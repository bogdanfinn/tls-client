package tls_client_cffi_src

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/websocket"
	"github.com/google/uuid"
)

// wsConnections maps connectionId -> *wsConn.
var wsConnections sync.Map

// ErrWsReadTimeout is the error message returned by WsRead when no message arrived in time.
// The connection stays usable after it.
const ErrWsReadTimeout = "websocket read timeout"

type wsMessage struct {
	err         error
	data        []byte
	messageType int
}

// wsConn wraps a websocket connection with a reader goroutine. The goroutine keeps reading so
// ping/pong and close frames are handled even while the caller is not inside WsRead, and so a
// WsRead timeout does not poison the connection (a gorilla read deadline error is permanent).
type wsConn struct {
	conn      *websocket.Conn
	messages  chan wsMessage
	done      chan struct{}
	closeOnce sync.Once
	writeLck  sync.Mutex

	// counters for WsStats; times are unix milliseconds, 0 = never
	connectedAt     int64
	messagesRead    atomic.Int64
	messagesWritten atomic.Int64
	bytesRead       atomic.Int64
	bytesWritten    atomic.Int64
	lastReadAt      atomic.Int64
	lastWriteAt     atomic.Int64
	lastWriteDurMs  atomic.Int64
	maxWriteDurMs   atomic.Int64
	// readPump holds a message nobody has collected with WsRead yet when pumped > collected. Two
	// counters, not a flag: a flag cleared by the pump after the handover is still set when the
	// caller that just got the message asks for stats.
	pumped    atomic.Int64
	collected atomic.Int64
}

func (c *wsConn) readPump() {
	for {
		mt, data, err := c.conn.ReadMessage()
		if err == nil {
			c.messagesRead.Add(1)
			c.bytesRead.Add(int64(len(data)))
			c.lastReadAt.Store(time.Now().UnixMilli())
		}

		c.pumped.Add(1)
		select {
		case c.messages <- wsMessage{messageType: mt, data: data, err: err}:
		case <-c.done:
			return
		}

		if err != nil {
			return
		}
	}
}

func (c *wsConn) close() {
	c.closeOnce.Do(func() {
		close(c.done)
		// best effort close frame; WriteControl is safe concurrently with other writes
		err := c.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
		if err != nil {
			// the path is dead or stalled: close the socket under TLS first, otherwise the TLS close
			// blocks for another 5s trying to send close_notify
			if tlsConn, ok := c.conn.NetConn().(interface{ NetConn() net.Conn }); ok {
				_ = tlsConn.NetConn().Close()
			}
		}
		_ = c.conn.Close()
	})
}

func getWsConn(connectionId string) (*wsConn, *TLSClientError) {
	val, ok := wsConnections.Load(connectionId)
	if !ok {
		return nil, NewTLSClientError(fmt.Errorf("no websocket connection found for connectionId: %s", connectionId))
	}

	return val.(*wsConn), nil
}

// closeWsConn removes the connection from the store and closes it. Safe to call multiple times.
func closeWsConn(connectionId string) bool {
	val, ok := wsConnections.LoadAndDelete(connectionId)
	if !ok {
		return false
	}

	val.(*wsConn).close()

	return true
}

// WsConnect establishes a WebSocket connection using either an existing session or an inline
// TLS client configuration. It returns a connectionId that must be used for subsequent
// WsRead, WsWrite, and WsClose calls.
//
// When SessionId is provided the existing TLS client is reused as-is and never modified: the
// caller must have created that session with forceHttp1: true, otherwise the handshake fails
// when the server negotiates h2. When creating an inline client, HTTP/1.1 is enforced
// automatically because WebSocket requires it.
func WsConnect(input WsConnectInput) (WsConnectOutput, *TLSClientError) {
	var tlsClient tls_client.HttpClient
	sessionId := ""
	withSession := false

	if input.SessionId != nil && *input.SessionId != "" {
		sessionId = *input.SessionId
		withSession = true

		var err error
		tlsClient, err = GetClient(sessionId)
		if err != nil {
			return WsConnectOutput{}, NewTLSClientError(err)
		}
	} else {
		// Build an inline client. Always enforce HTTP/1.1 since WebSocket requires it.
		requestInput := RequestInput{
			TLSClientIdentifier:         input.TLSClientIdentifier,
			CustomTlsClient:             input.CustomTlsClient,
			ProxyUrl:                    input.ProxyUrl,
			ForceHttp1:                  true,
			InsecureSkipVerify:          input.InsecureSkipVerify,
			WithRandomTLSExtensionOrder: input.WithRandomTLSExtensionOrder,
			WithoutCookieJar:            true,
			FollowRedirects:             false,
		}

		var clientErr *TLSClientError
		tlsClient, sessionId, withSession, clientErr = CreateClient(requestInput)
		if clientErr != nil {
			return WsConnectOutput{}, clientErr
		}
	}

	headers := http.Header{}
	for k, v := range input.Headers {
		headers[k] = []string{v}
	}
	if len(input.HeaderOrder) > 0 {
		headers[http.HeaderOrderKey] = input.HeaderOrder
	}

	wsOptions := []tls_client.WebsocketOption{
		tls_client.WithTlsClient(tlsClient),
		tls_client.WithUrl(input.Url),
		tls_client.WithHeaders(headers),
		tls_client.WithEnableCompression(),
	}

	if input.HandshakeTimeoutMilliseconds > 0 {
		wsOptions = append(wsOptions, tls_client.WithHandshakeTimeoutMilliseconds(input.HandshakeTimeoutMilliseconds))
	}
	if input.ReadBufferSize > 0 {
		wsOptions = append(wsOptions, tls_client.WithReadBufferSize(input.ReadBufferSize))
	}
	if input.WriteBufferSize > 0 {
		wsOptions = append(wsOptions, tls_client.WithWriteBufferSize(input.WriteBufferSize))
	}

	ws, err := tls_client.NewWebsocket(nil, wsOptions...)
	if err != nil {
		return WsConnectOutput{}, NewTLSClientError(fmt.Errorf("failed to create websocket: %w", err))
	}

	conn, err := ws.Connect(context.Background())
	if err != nil {
		return WsConnectOutput{}, NewTLSClientError(fmt.Errorf("failed to connect websocket: %w", err))
	}

	connectionId := uuid.New().String()
	c := &wsConn{
		conn:        conn,
		messages:    make(chan wsMessage),
		done:        make(chan struct{}),
		connectedAt: time.Now().UnixMilli(),
	}
	wsConnections.Store(connectionId, c)
	go c.readPump()

	out := WsConnectOutput{
		Id:           uuid.New().String(),
		ConnectionId: connectionId,
		Status:       101,
	}
	if withSession {
		out.SessionId = sessionId
	}

	return out, nil
}

// WsRead returns the next message from an active WebSocket connection, blocking until one arrives.
// If TimeoutMilliseconds is > 0 and nothing arrives in time, an ErrWsReadTimeout error is returned
// and the connection stays open. Any other read error (including a close from the peer) closes the
// connection and removes it from the store.
// Text messages are returned as-is; binary messages are base64-encoded.
func WsRead(input WsReadInput) (WsReadOutput, *TLSClientError) {
	c, clientErr := getWsConn(input.ConnectionId)
	if clientErr != nil {
		return WsReadOutput{}, clientErr
	}

	var timeout <-chan time.Time
	if input.TimeoutMilliseconds > 0 {
		timer := time.NewTimer(time.Duration(input.TimeoutMilliseconds) * time.Millisecond)
		defer timer.Stop()
		timeout = timer.C
	}

	var msg wsMessage
	select {
	case msg = <-c.messages:
		c.collected.Add(1)
	case <-timeout:
		return WsReadOutput{}, NewTLSClientError(fmt.Errorf(ErrWsReadTimeout))
	case <-c.done:
		return WsReadOutput{}, NewTLSClientError(fmt.Errorf("websocket connection closed: %s", input.ConnectionId))
	}

	if msg.err != nil {
		closeWsConn(input.ConnectionId)

		return WsReadOutput{}, NewTLSClientError(fmt.Errorf("failed to read message: %w", msg.err))
	}

	data := string(msg.data)
	if msg.messageType == websocket.BinaryMessage {
		data = base64.StdEncoding.EncodeToString(msg.data)
	}

	return WsReadOutput{
		Id:           uuid.New().String(),
		ConnectionId: input.ConnectionId,
		MessageType:  msg.messageType,
		Data:         data,
	}, nil
}

// WsWrite sends a message over an active WebSocket connection. Safe to call concurrently with
// WsRead and with other WsWrite calls on the same connection. A failed write closes the
// connection and removes it from the store.
// For binary messages (MessageType 2) the Data field must be base64-encoded.
func WsWrite(input WsWriteInput) (WsWriteOutput, *TLSClientError) {
	c, clientErr := getWsConn(input.ConnectionId)
	if clientErr != nil {
		return WsWriteOutput{}, clientErr
	}

	if input.MessageType != websocket.TextMessage && input.MessageType != websocket.BinaryMessage {
		return WsWriteOutput{}, NewTLSClientError(fmt.Errorf("unsupported messageType %d: use 1 (text) or 2 (binary)", input.MessageType))
	}

	msgData := []byte(input.Data)
	if input.MessageType == websocket.BinaryMessage {
		var decodeErr error
		msgData, decodeErr = base64.StdEncoding.DecodeString(input.Data)
		if decodeErr != nil {
			return WsWriteOutput{}, NewTLSClientError(fmt.Errorf("failed to base64 decode binary message: %w", decodeErr))
		}
	}

	writeStart := time.Now() // includes the wait for the write lock
	var deadline time.Time   // zero = none; always set, a deadline from an earlier write would stick
	if input.TimeoutMilliseconds > 0 {
		deadline = writeStart.Add(time.Duration(input.TimeoutMilliseconds) * time.Millisecond)
	}

	c.writeLck.Lock()
	_ = c.conn.SetWriteDeadline(deadline)
	writeErr := c.conn.WriteMessage(input.MessageType, msgData)
	c.writeLck.Unlock()

	durMs := time.Since(writeStart).Milliseconds()
	c.lastWriteDurMs.Store(durMs)
	if durMs > c.maxWriteDurMs.Load() {
		c.maxWriteDurMs.Store(durMs) // racy max is fine for a diagnostic
	}
	if writeErr == nil {
		c.messagesWritten.Add(1)
		c.bytesWritten.Add(int64(len(msgData)))
		c.lastWriteAt.Store(time.Now().UnixMilli())
	}

	if writeErr != nil {
		// write errors are permanent on a websocket conn: drop it so a blocked WsRead returns
		// and the caller does not keep writing into a dead socket
		closeWsConn(input.ConnectionId)

		return WsWriteOutput{}, NewTLSClientError(fmt.Errorf("failed to write message: %w", writeErr))
	}

	return WsWriteOutput{
		Id:           uuid.New().String(),
		ConnectionId: input.ConnectionId,
		Success:      true,
	}, nil
}

// WsClose closes an active WebSocket connection and removes it from the connection store.
// A WsRead blocked on the same connection returns with an error.
func WsClose(input WsCloseInput) (WsCloseOutput, *TLSClientError) {
	if !closeWsConn(input.ConnectionId) {
		return WsCloseOutput{}, NewTLSClientError(fmt.Errorf("no websocket connection found for connectionId: %s", input.ConnectionId))
	}

	return WsCloseOutput{
		Id:      uuid.New().String(),
		Success: true,
	}, nil
}

// WsStats returns a snapshot of a live connection for diagnosing silent stalls: message counters,
// how long ago the last message was read / written, how long writes block, whether a message is
// waiting for WsRead, and (linux) the kernel's TCP_INFO for the socket to the server or proxy.
// It never touches the wire. A connection that already failed or was closed is gone from the store.
func WsStats(input WsStatsInput) (WsStatsOutput, *TLSClientError) {
	c, clientErr := getWsConn(input.ConnectionId)
	if clientErr != nil {
		return WsStatsOutput{}, clientErr
	}

	now := time.Now().UnixMilli()
	since := func(at int64) int64 {
		if at == 0 {
			return -1
		}
		return now - at
	}

	return WsStatsOutput{
		Id:                  uuid.New().String(),
		ConnectionId:        input.ConnectionId,
		AgeMs:               now - c.connectedAt,
		MessagesRead:        c.messagesRead.Load(),
		MessagesWritten:     c.messagesWritten.Load(),
		BytesRead:           c.bytesRead.Load(),
		BytesWritten:        c.bytesWritten.Load(),
		MsSinceLastRead:     since(c.lastReadAt.Load()),
		MsSinceLastWrite:    since(c.lastWriteAt.Load()),
		LastWriteDurationMs: c.lastWriteDurMs.Load(),
		MaxWriteDurationMs:  c.maxWriteDurMs.Load(),
		UnreadPending:       c.pumped.Load() > c.collected.Load(),
		LocalAddr:           c.conn.LocalAddr().String(),
		RemoteAddr:          c.conn.RemoteAddr().String(),
		Tcp:                 wsTcpInfo(c.conn.UnderlyingConn()),
	}, nil
}
