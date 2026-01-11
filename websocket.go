package tls_client

import (
	"context"
	"fmt"

	"github.com/bogdanfinn/websocket"
)

type Websocket struct {
	config websocketConfig
	dialer *websocket.Dialer
}

// New creates a new WebSocket wrapper that uses tls-client for connections.
// This allows WebSocket connections to use the same TLS fingerprinting and
// configuration as regular HTTP requests.
//
// Example usage:
//
//	// Create HTTP client with ForceHttp1 (required for WebSocket!)
//	client, _ := NewHttpClient(nil,
//	    WithClientProfile(profiles.Chrome_133),
//	    WithForceHttp1(),
//	)
//
//	// Create WebSocket with optional header ordering
//	headers := http.Header{
//	    "User-Agent": {"MyBot/1.0"},
//	    http.HeaderOrderKey: {"host", "upgrade", "connection", "user-agent"},
//	}
//
//	ws, _ := New(nil,
//	    WithTlsClient(client),
//	    WithUrl("wss://example.com/ws"),
//	    WithHeaders(headers),
//	)
//
//	conn, _ := ws.Connect(context.Background())
//	defer conn.Close()
func New(logger Logger, options ...WebsocketOption) (*Websocket, error) {
	config := &websocketConfig{}

	for _, opt := range options {
		opt(config)
	}

	if err := validateWebsocketConfig(config); err != nil {
		return nil, err
	}

	dialer := &websocket.Dialer{
		HandshakeTimeout:  config.handshakeTimeout,
		Jar:               config.cookieJar,
		ReadBufferSize:    config.readBufferSize,
		WriteBufferSize:   config.writeBufferSize,
		NetDialTLSContext: config.tlsClient.GetDialer().DialContext,
		NetDialContext:    config.tlsClient.GetDialer().DialContext,
	}

	return &Websocket{
		config: *config,
		dialer: dialer,
	}, nil
}

func (w *Websocket) Connect(ctx context.Context) (*websocket.Conn, error) {
	c, _, err := w.dialer.DialContext(ctx, w.config.url, w.config.headers)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func validateWebsocketConfig(config *websocketConfig) error {
	if config.tlsClient == nil {
		return fmt.Errorf("tlsClient cannot be nil for websocket connection")
	}

	if config.url == "" {
		return fmt.Errorf("url cannot be empty for websocket connection")
	}

	return nil
}
