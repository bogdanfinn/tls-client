package tls_client

import (
	"time"
)

type HttpClientOption func(config *httpClientConfig)

type httpClientConfig struct {
	debug              bool
	followRedirects    bool
	insecureSkipVerify bool
	proxyUrl           string
	clientProfile      ClientProfile
	timeout            time.Duration
}

func WithProxyUrl(proxyUrl string) HttpClientOption {
	return func(config *httpClientConfig) {
		config.proxyUrl = proxyUrl
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
