// Package main shows a WebSocket connection that reuses the TLS fingerprint
// of a tls-client. The handshake goes through the client's dialer, so the
// connection is indistinguishable from the browser the profile imitates.
//
// Run: go run ./example/websocket
package main

import (
	"context"
	"log"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/bogdanfinn/websocket"
)

func main() {
	// WebSocket needs HTTP/1.1, so WithForceHttp1 is mandatory here.
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_150),
		tls_client.WithForceHttp1(),
	}...)
	if err != nil {
		log.Fatal(err)
	}

	// http.HeaderOrderKey works for the handshake headers just like it does
	// for a regular request.
	headers := http.Header{
		"user-agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"host",
			"upgrade",
			"connection",
			"sec-websocket-key",
			"sec-websocket-version",
			"user-agent",
		},
	}

	ws, err := tls_client.NewWebsocket(tls_client.NewNoopLogger(),
		tls_client.WithUrl("wss://echo.websocket.events"),
		tls_client.WithTlsClient(client),
		tls_client.WithHeaders(headers),
		tls_client.WithHandshakeTimeoutMilliseconds(10000),
	)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := ws.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if err := conn.WriteMessage(websocket.TextMessage, []byte("hello from tls-client")); err != nil {
		log.Fatal(err)
	}

	// echo.websocket.events sends a greeting first, then echoes back.
	for i := 0; i < 2; i++ {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("received: %s\n", message)
	}
}
