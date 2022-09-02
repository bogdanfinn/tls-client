package tls_client

import (
	"time"

	http "github.com/bogdanfinn/fhttp"
)

type HttpClientOption func(config *httpClientConfig)

type httpClientConfig struct {
	debug              bool
	followRedirects    bool
	insecureSkipVerify bool
	proxyUrl           string
	cookieJar          http.CookieJar
	clientProfile      ClientProfile
	ja3String          string
	timeout            time.Duration
}

func (o *httpClientConfig) IsJa3StringSet() bool {
	return o.ja3String != ""
}

func (o *httpClientConfig) IsClientProfileSet() bool {
	return o.clientProfile.clientHelloId.Str() != ""
}

func WithProxyUrl(proxyUrl string) HttpClientOption {
	return func(config *httpClientConfig) {
		config.proxyUrl = proxyUrl
	}
}

func WithCookieJar(jar http.CookieJar) HttpClientOption {
	return func(config *httpClientConfig) {
		config.cookieJar = jar
	}
}

func WithTimeout(timeout int) HttpClientOption {
	return func(config *httpClientConfig) {
		config.timeout = time.Second * time.Duration(timeout)
	}
}

func WithNotFollowRedirects() HttpClientOption {
	return func(config *httpClientConfig) {
		config.followRedirects = false
	}
}

func WithDebug() HttpClientOption {
	return func(config *httpClientConfig) {
		config.debug = true
	}
}

func WithInsecureSkipVerify() HttpClientOption {
	return func(config *httpClientConfig) {
		config.insecureSkipVerify = true
	}
}

func WithClientProfile(clientProfile ClientProfile) HttpClientOption {
	return func(config *httpClientConfig) {
		config.clientProfile = clientProfile
	}
}

func WithJa3String(ja3String string) HttpClientOption {
	return func(config *httpClientConfig) {
		config.ja3String = ja3String
	}
}
