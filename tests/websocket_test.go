package tests

import (
	"context"
	"testing"
	"time"

	"github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/httptest"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	gorillaWebsocket "github.com/bogdanfinn/websocket"
	"github.com/stretchr/testify/require"
)

var upgrader = gorillaWebsocket.Upgrader{}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

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

func TestWebSocketEcho(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(echoHandler))
	defer server.Close()

	url := "ws" + server.URL[len("http"):]

	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithRandomTLSExtensionOrder(),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	websocketOptions := []tls_client.WebsocketOption{
		tls_client.WithTlsClient(client),
		tls_client.WithUrl(url),
		tls_client.WithHeaders(http.Header{}),
		tls_client.WithHandshakeTimeoutMilliseconds(1000),
	}

	ws, err := tls_client.NewWebsocket(nil, websocketOptions...)
	if err != nil {
		t.Fatal(err)
	}

	wsConnection, err := ws.Connect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	defer wsConnection.Close()

	expected := "hello world"
	err = wsConnection.WriteMessage(gorillaWebsocket.TextMessage, []byte(expected))
	require.NoError(t, err)

	_, msg, err := wsConnection.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, expected, string(msg))

	wsConnection.SetReadDeadline(time.Now().Add(2 * time.Second))
}

func TestWebSocketEchoRealWebserver(t *testing.T) {
	url := "wss://echo.websocket.org"

	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithRandomTLSExtensionOrder(),
		tls_client.WithForceHttp1(), // WebSocket requires HTTP/1.1
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	websocketOptions := []tls_client.WebsocketOption{
		tls_client.WithTlsClient(client),
		tls_client.WithUrl(url),
		tls_client.WithHeaders(http.Header{}),
		tls_client.WithHandshakeTimeoutMilliseconds(1000),
	}

	ws, err := tls_client.NewWebsocket(nil, websocketOptions...)
	if err != nil {
		t.Fatal(err)
	}

	wsConnection, err := ws.Connect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	defer wsConnection.Close()

	testMessage := "hello world"
	err = wsConnection.WriteMessage(gorillaWebsocket.TextMessage, []byte(testMessage))
	require.NoError(t, err)

	_, msg, err := wsConnection.ReadMessage()
	require.NoError(t, err)
	// echo.websocket.org sends its own message instead of echoing, so just verify we got a response
	require.NotEmpty(t, string(msg))
	t.Logf("Received message from server: %s", string(msg))

	wsConnection.SetReadDeadline(time.Now().Add(2 * time.Second))
}

func TestWebSocketWithHeaderOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(echoHandler))
	defer server.Close()

	url := "ws" + server.URL[len("http"):]

	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithRandomTLSExtensionOrder(),
		tls_client.WithForceHttp1(),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	require.NoError(t, err)

	customHeaders := http.Header{
		"User-Agent":    {"CustomBot/1.0"},
		"Custom-Header": {"CustomValue"},
		http.HeaderOrderKey: {
			"host",
			"upgrade",
			"connection",
			"sec-websocket-key",
			"sec-websocket-version",
			"user-agent",
			"custom-header",
		},
	}

	websocketOptions := []tls_client.WebsocketOption{
		tls_client.WithTlsClient(client),
		tls_client.WithUrl(url),
		tls_client.WithHeaders(customHeaders),
		tls_client.WithHandshakeTimeoutMilliseconds(1000),
	}

	ws, err := tls_client.NewWebsocket(nil, websocketOptions...)
	require.NoError(t, err)

	wsConnection, err := ws.Connect(context.Background())
	require.NoError(t, err)
	defer wsConnection.Close()

	expected := "header order test"
	err = wsConnection.WriteMessage(gorillaWebsocket.TextMessage, []byte(expected))
	require.NoError(t, err)

	_, msg, err := wsConnection.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, expected, string(msg))
}

func TestWebSocketWithoutHeaderOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(echoHandler))
	defer server.Close()

	url := "ws" + server.URL[len("http"):]

	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithRandomTLSExtensionOrder(),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	require.NoError(t, err)

	websocketOptions := []tls_client.WebsocketOption{
		tls_client.WithTlsClient(client),
		tls_client.WithUrl(url),
		tls_client.WithHeaders(http.Header{
			"User-Agent": {"TestBot/1.0"},
		}),
		tls_client.WithHandshakeTimeoutMilliseconds(1000),
	}

	ws, err := tls_client.NewWebsocket(nil, websocketOptions...)
	require.NoError(t, err)

	wsConnection, err := ws.Connect(context.Background())
	require.NoError(t, err)
	defer wsConnection.Close()

	expected := "no header order test"
	err = wsConnection.WriteMessage(gorillaWebsocket.TextMessage, []byte(expected))
	require.NoError(t, err)

	_, msg, err := wsConnection.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, expected, string(msg))
}
