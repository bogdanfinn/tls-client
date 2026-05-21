package tls_client_cffi_src

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/websocket"
	"github.com/google/uuid"
)

var wsConnections sync.Map

// WsConnect establishes a WebSocket connection using either an existing session or an inline
// TLS client configuration. It returns a connectionId that must be used for subsequent
// WsRead, WsWrite, and WsClose calls.
//
// When SessionId is provided the existing TLS client (including its ForceHttp1 setting) is
// reused. When creating an inline client, HTTP/1.1 is enforced automatically because
// WebSocket requires it.
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
	wsConnections.Store(connectionId, conn)

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

// WsRead reads a single message from an active WebSocket connection.
// If TimeoutMilliseconds is > 0 a read deadline is applied for that call.
// Text messages are returned as-is; binary messages are base64-encoded.
func WsRead(input WsReadInput) (WsReadOutput, *TLSClientError) {
	val, ok := wsConnections.Load(input.ConnectionId)
	if !ok {
		return WsReadOutput{}, NewTLSClientError(fmt.Errorf("no websocket connection found for connectionId: %s", input.ConnectionId))
	}

	conn := val.(*websocket.Conn)

	if input.TimeoutMilliseconds > 0 {
		conn.SetReadDeadline(time.Now().Add(time.Duration(input.TimeoutMilliseconds) * time.Millisecond))
	}

	mt, msg, err := conn.ReadMessage()
	if err != nil {
		return WsReadOutput{}, NewTLSClientError(fmt.Errorf("failed to read message: %w", err))
	}

	data := string(msg)
	if mt == websocket.BinaryMessage {
		data = base64.StdEncoding.EncodeToString(msg)
	}

	return WsReadOutput{
		Id:           uuid.New().String(),
		ConnectionId: input.ConnectionId,
		MessageType:  mt,
		Data:         data,
	}, nil
}

// WsWrite sends a message over an active WebSocket connection.
// For binary messages (MessageType 2) the Data field must be base64-encoded.
func WsWrite(input WsWriteInput) (WsWriteOutput, *TLSClientError) {
	val, ok := wsConnections.Load(input.ConnectionId)
	if !ok {
		return WsWriteOutput{}, NewTLSClientError(fmt.Errorf("no websocket connection found for connectionId: %s", input.ConnectionId))
	}

	conn := val.(*websocket.Conn)

	var msgData []byte
	if input.MessageType == websocket.BinaryMessage {
		var decodeErr error
		msgData, decodeErr = base64.StdEncoding.DecodeString(input.Data)
		if decodeErr != nil {
			return WsWriteOutput{}, NewTLSClientError(fmt.Errorf("failed to base64 decode binary message: %w", decodeErr))
		}
	} else {
		msgData = []byte(input.Data)
	}

	if writeErr := conn.WriteMessage(input.MessageType, msgData); writeErr != nil {
		return WsWriteOutput{}, NewTLSClientError(fmt.Errorf("failed to write message: %w", writeErr))
	}

	return WsWriteOutput{
		Id:           uuid.New().String(),
		ConnectionId: input.ConnectionId,
		Success:      true,
	}, nil
}

// WsClose closes an active WebSocket connection and removes it from the connection store.
func WsClose(input WsCloseInput) (WsCloseOutput, *TLSClientError) {
	val, ok := wsConnections.Load(input.ConnectionId)
	if !ok {
		return WsCloseOutput{}, NewTLSClientError(fmt.Errorf("no websocket connection found for connectionId: %s", input.ConnectionId))
	}

	conn := val.(*websocket.Conn)
	conn.Close()
	wsConnections.Delete(input.ConnectionId)

	return WsCloseOutput{
		Id:      uuid.New().String(),
		Success: true,
	}, nil
}
