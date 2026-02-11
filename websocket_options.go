package tls_client

import (
	"time"

	http "github.com/bogdanfinn/fhttp"
)

type WebsocketOption func(config *websocketConfig)

type websocketConfig struct {
	url              string
	tlsClient        HttpClient
	headers          http.Header
	readBufferSize   int
	writeBufferSize  int
	handshakeTimeout time.Duration
	cookieJar        http.CookieJar
}

func WithUrl(url string) WebsocketOption {
	return func(config *websocketConfig) {
		config.url = url
	}
}

// WithTlsClient sets the tls-client HttpClient to use for the WebSocket connection.
// The underlying dialer from this client will be used to establish the connection,
// preserving TLS fingerprinting and other client configurations.
//
// IMPORTANT: WebSocket connections require HTTP/1.1. When creating your HttpClient,
// you MUST use WithForceHttp1() option to ensure compatibility:
//
//	client, _ := NewHttpClient(nil,
//	    WithClientProfile(profiles.Chrome_133),
//	    WithForceHttp1(), // Required for WebSocket!
//	)
func WithTlsClient(tlsClient HttpClient) WebsocketOption {
	return func(config *websocketConfig) {
		config.tlsClient = tlsClient
	}
}

func WithHeaders(headers http.Header) WebsocketOption {
	return func(config *websocketConfig) {
		config.headers = headers
	}
}

func WithReadBufferSize(readBufferSize int) WebsocketOption {
	return func(config *websocketConfig) {
		config.readBufferSize = readBufferSize
	}
}

func WithWriteBufferSize(writeBufferSize int) WebsocketOption {
	return func(config *websocketConfig) {
		config.writeBufferSize = writeBufferSize
	}
}

func WithHandshakeTimeoutMilliseconds(timeout int) WebsocketOption {
	return func(config *websocketConfig) {
		config.handshakeTimeout = time.Millisecond * time.Duration(timeout)
	}
}

func WithCookiejar(cookiejar http.CookieJar) WebsocketOption {
	return func(config *websocketConfig) {
		config.cookieJar = cookiejar
	}
}
