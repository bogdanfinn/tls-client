package tls_client_cffi_src

import (
	"encoding/base64"
	"runtime"
	"strings"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/httptest"
	"github.com/bogdanfinn/websocket"
	"github.com/stretchr/testify/require"
)

// wsServer starts a local TLS websocket server and returns its wss:// url. handle runs once per
// upgraded connection, which is closed when it returns.
func wsServer(t *testing.T, upgrader websocket.Upgrader, handle func(conn *websocket.Conn)) string {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		handle(conn)
	}))
	t.Cleanup(server.Close)

	return "wss" + strings.TrimPrefix(server.URL, "https")
}

func wsEcho(conn *websocket.Conn) {
	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if err := conn.WriteMessage(mt, msg); err != nil {
			return
		}
	}
}

func wsEchoServer(t *testing.T) string {
	return wsServer(t, websocket.Upgrader{}, wsEcho)
}

func wsDial(url string) (WsConnectOutput, *TLSClientError) {
	return WsConnect(WsConnectInput{
		TLSClientIdentifier:          "chrome_133",
		Url:                          url,
		Headers:                      map[string]string{"User-Agent": "cffi-ws-test"},
		HeaderOrder:                  []string{"host", "upgrade", "connection", "user-agent"},
		HandshakeTimeoutMilliseconds: 30000,
		InsecureSkipVerify:           true,
	})
}

func wsConnectTo(t *testing.T, url string) string {
	out, err := wsDial(url)
	require.Nil(t, err)
	require.Equal(t, 101, out.Status)
	require.NotEmpty(t, out.ConnectionId)
	require.Empty(t, out.SessionId)

	return out.ConnectionId
}

func wsTestConnect(t *testing.T) string {
	return wsConnectTo(t, wsEchoServer(t))
}

func TestWsEchoTextAndBinary(t *testing.T) {
	connectionId := wsTestConnect(t)

	_, err := WsWrite(WsWriteInput{ConnectionId: connectionId, MessageType: websocket.TextMessage, Data: "hello"})
	require.Nil(t, err)

	read, err := WsRead(WsReadInput{ConnectionId: connectionId, TimeoutMilliseconds: 5000})
	require.Nil(t, err)
	require.Equal(t, websocket.TextMessage, read.MessageType)
	require.Equal(t, "hello", read.Data)

	binary := base64.StdEncoding.EncodeToString([]byte{0x00, 0xff, 0x10, 0x80})
	_, err = WsWrite(WsWriteInput{ConnectionId: connectionId, MessageType: websocket.BinaryMessage, Data: binary})
	require.Nil(t, err)

	read, err = WsRead(WsReadInput{ConnectionId: connectionId, TimeoutMilliseconds: 5000})
	require.Nil(t, err)
	require.Equal(t, websocket.BinaryMessage, read.MessageType)
	require.Equal(t, binary, read.Data)

	_, err = WsWrite(WsWriteInput{ConnectionId: connectionId, MessageType: websocket.BinaryMessage, Data: "not base64!"})
	require.NotNil(t, err)

	_, err = WsClose(WsCloseInput{ConnectionId: connectionId})
	require.Nil(t, err)
}

func TestWsReadTimeoutKeepsConnectionUsable(t *testing.T) {
	connectionId := wsTestConnect(t)

	_, err := WsRead(WsReadInput{ConnectionId: connectionId, TimeoutMilliseconds: 50})
	require.NotNil(t, err)
	require.Equal(t, ErrWsReadTimeout, err.Error())

	_, err = WsWrite(WsWriteInput{ConnectionId: connectionId, MessageType: websocket.TextMessage, Data: "still alive"})
	require.Nil(t, err)

	read, err := WsRead(WsReadInput{ConnectionId: connectionId, TimeoutMilliseconds: 5000})
	require.Nil(t, err)
	require.Equal(t, "still alive", read.Data)

	_, err = WsClose(WsCloseInput{ConnectionId: connectionId})
	require.Nil(t, err)
}

func TestWsUnknownAndClosedConnection(t *testing.T) {
	_, err := WsRead(WsReadInput{ConnectionId: "nope"})
	require.NotNil(t, err)
	_, err = WsWrite(WsWriteInput{ConnectionId: "nope", MessageType: websocket.TextMessage})
	require.NotNil(t, err)
	_, err = WsClose(WsCloseInput{ConnectionId: "nope"})
	require.NotNil(t, err)

	connectionId := wsTestConnect(t)

	// a blocked read must be released by close
	readErr := make(chan *TLSClientError)
	go func() {
		_, err := WsRead(WsReadInput{ConnectionId: connectionId})
		readErr <- err
	}()

	_, err = WsClose(WsCloseInput{ConnectionId: connectionId})
	require.Nil(t, err)
	require.NotNil(t, <-readErr)

	_, err = WsRead(WsReadInput{ConnectionId: connectionId})
	require.NotNil(t, err)
	_, err = WsWrite(WsWriteInput{ConnectionId: connectionId, MessageType: websocket.TextMessage, Data: "x"})
	require.NotNil(t, err)
	_, err = WsClose(WsCloseInput{ConnectionId: connectionId})
	require.NotNil(t, err)

	_, ok := wsConnections.Load(connectionId)
	require.False(t, ok)
}

func TestWsConnectFailureLeavesNoConnection(t *testing.T) {
	_, err := WsConnect(WsConnectInput{TLSClientIdentifier: "chrome_133", Url: "wss://127.0.0.1:1/ws", HandshakeTimeoutMilliseconds: 1000})
	require.NotNil(t, err)

	missing := "missing-session"
	_, err = WsConnect(WsConnectInput{SessionId: &missing, Url: "wss://127.0.0.1:1/ws"})
	require.NotNil(t, err)

	count := 0
	wsConnections.Range(func(_, _ any) bool { count++; return true })
	require.Zero(t, count)
}

func TestWsWriteErrorClosesConnection(t *testing.T) {
	connectionId := wsTestConnect(t)

	c, _ := getWsConn(connectionId)
	_ = c.conn.NetConn().Close() // simulate a dead tunnel

	readErr := make(chan *TLSClientError)
	go func() {
		_, err := WsRead(WsReadInput{ConnectionId: connectionId})
		readErr <- err
	}()

	_, err := WsWrite(WsWriteInput{ConnectionId: connectionId, MessageType: websocket.TextMessage, Data: "x"})
	require.NotNil(t, err)
	require.NotNil(t, <-readErr)

	_, ok := wsConnections.Load(connectionId)
	require.False(t, ok)
}

func TestWsStats(t *testing.T) {
	connectionId := wsTestConnect(t)

	_, err := WsWrite(WsWriteInput{ConnectionId: connectionId, MessageType: websocket.TextMessage, Data: "hello"})
	require.Nil(t, err)
	_, err = WsRead(WsReadInput{ConnectionId: connectionId, TimeoutMilliseconds: 5000})
	require.Nil(t, err)

	stats, err := WsStats(WsStatsInput{ConnectionId: connectionId})
	require.Nil(t, err)
	require.EqualValues(t, 1, stats.MessagesRead)
	require.EqualValues(t, 1, stats.MessagesWritten)
	require.EqualValues(t, 5, stats.BytesWritten)
	require.GreaterOrEqual(t, stats.MsSinceLastRead, int64(0))
	require.False(t, stats.UnreadPending)
	if runtime.GOOS == "linux" {
		require.NotNil(t, stats.Tcp)
		require.EqualValues(t, 1, stats.Tcp.State)
	}

	_, err = WsClose(WsCloseInput{ConnectionId: connectionId})
	require.Nil(t, err)
	_, err = WsStats(WsStatsInput{ConnectionId: connectionId})
	require.NotNil(t, err)
}
