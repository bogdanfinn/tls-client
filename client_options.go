package tls_client

import (
	"time"

	http "github.com/bogdanfinn/fhttp"
)

type HttpClientOption func(config *httpClientConfig)

type TransportOptions struct {
	DisableKeepAlives      bool
	DisableCompression     bool
	MaxIdleConns           int
	MaxIdleConnsPerHost    int
	MaxConnsPerHost        int
	MaxResponseHeaderBytes int64 // Zero means to use a default limit.
	WriteBufferSize        int   // If zero, a default (currently 4KB) is used.
	ReadBufferSize         int   // If zero, a default (currently 4KB) is used.
}

type httpClientConfig struct {
	debug                       bool
	followRedirects             bool
	insecureSkipVerify          bool
	proxyUrl                    string
	serverNameOverwrite         string
	transportOptions            *TransportOptions
	cookieJar                   http.CookieJar
	clientProfile               ClientProfile
	withRandomTlsExtensionOrder bool
	forceHttp1                  bool
	timeout                     time.Duration
}

// WithProxyUrl configures a HTTP client to use the specified proxy URL.
//
// proxyUrl should be formatted as:
//
//	"http://user:pass@host:port"
func WithProxyUrl(proxyUrl string) HttpClientOption {
	return func(config *httpClientConfig) {
		config.proxyUrl = proxyUrl
	}
}

// WithCookieJar configures a HTTP client to use the specified cookie jar.
func WithCookieJar(jar http.CookieJar) HttpClientOption {
	return func(config *httpClientConfig) {
		config.cookieJar = jar
	}
}

// WithTimeout configures a HTTP client to use the specified request timeout.
//
// timeout is the request timeout in seconds.
func WithTimeout(timeout int) HttpClientOption {
	return func(config *httpClientConfig) {
		config.timeout = time.Second * time.Duration(timeout)
	}
}

// WithNotFollowRedirects configures a HTTP client to not follow HTTP redirects.
func WithNotFollowRedirects() HttpClientOption {
	return func(config *httpClientConfig) {
		config.followRedirects = false
	}
}

// WithRandomTLSExtensionOrder configures a TLS client to randomize the order of TLS extensions being sent in the ClientHello.
//
// Placement of GREASE and padding is fixed and will not be affected by this.
func WithRandomTLSExtensionOrder() HttpClientOption {
	return func(config *httpClientConfig) {
		config.withRandomTlsExtensionOrder = true
	}
}

// WithDebug configures a client to log debugging information.
func WithDebug() HttpClientOption {
	return func(config *httpClientConfig) {
		config.debug = true
	}
}

// WithTransportOptions configures a client to use the specified transport options.
func WithTransportOptions(transportOptions *TransportOptions) HttpClientOption {
	return func(config *httpClientConfig) {
		config.transportOptions = transportOptions
	}
}

// WithInsecureSkipVerify configures a client to skip SSL certificate verification.
func WithInsecureSkipVerify() HttpClientOption {
	return func(config *httpClientConfig) {
		config.insecureSkipVerify = true
	}
}

// WithForceHttp1 configures a client to force HTTP/1.1 as the used protocol.
func WithForceHttp1() HttpClientOption {
	return func(config *httpClientConfig) {
		config.forceHttp1 = true
	}
}

// WithClientProfile configures a TLS client to use the specified client profile.
func WithClientProfile(clientProfile ClientProfile) HttpClientOption {
	return func(config *httpClientConfig) {
		config.clientProfile = clientProfile
	}
}

// WithServerNameOverwrite configures a TLS client to overwrite the server name being used for certificate verification and in the client hello.
func WithServerNameOverwrite(serverName string) HttpClientOption {
	return func(config *httpClientConfig) {
		config.serverNameOverwrite = serverName
	}
}
