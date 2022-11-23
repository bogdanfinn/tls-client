package tls_client

import (
	"fmt"
	"io"
	"net/url"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"golang.org/x/net/proxy"
)

var defaultRedirectFunc = func(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

type HttpClient interface {
	GetCookies(u *url.URL) []*http.Cookie
	SetCookies(u *url.URL, cookies []*http.Cookie)
	SetProxy(proxyUrl string) error
	GetProxy() string
	SetFollowRedirect(followRedirect bool)
	GetFollowRedirect() bool
	Do(req *http.Request) (*http.Response, error)
	Get(url string) (resp *http.Response, err error)
	Head(url string) (resp *http.Response, err error)
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

type httpClient struct {
	http.Client
	logger Logger
	config *httpClientConfig
}

var DefaultTimeoutSeconds = 30

var DefaultOptions = []HttpClientOption{
	WithTimeout(DefaultTimeoutSeconds),
	WithClientProfile(DefaultClientProfile),
	WithNotFollowRedirects(),
}

func ProvideDefaultClient(logger Logger) (HttpClient, error) {
	jar := NewCookieJar(nil)

	return NewHttpClient(logger, append(DefaultOptions, WithCookieJar(jar))...)
}

func NewHttpClient(logger Logger, options ...HttpClientOption) (HttpClient, error) {
	config := &httpClientConfig{
		followRedirects: true,
		timeout:         time.Duration(DefaultTimeoutSeconds) * time.Second,
	}

	for _, opt := range options {
		opt(config)
	}

	err := validateConfig(config)

	if err != nil {
		return nil, err
	}

	client, clientProfile, err := buildFromConfig(config)

	if err != nil {
		return nil, err
	}

	config.clientProfile = clientProfile

	if config.debug {
		if logger == nil {
			logger = NewLogger()
		}

		logger = NewDebugLogger(logger)
	}

	if logger == nil {
		logger = NewNoopLogger()
	}

	return &httpClient{
		Client: *client,
		logger: logger,
		config: config,
	}, nil
}

func validateConfig(config *httpClientConfig) error {
	return nil
}

func buildFromConfig(config *httpClientConfig) (*http.Client, ClientProfile, error) {
	var dialer proxy.ContextDialer
	dialer = newDirectDialer(config.timeout)

	if config.proxyUrl != "" {
		proxyDialer, err := newConnectDialer(config.proxyUrl, config.timeout)
		if err != nil {
			return nil, ClientProfile{}, err
		}

		dialer = proxyDialer
	}

	var redirectFunc func(req *http.Request, via []*http.Request) error
	if !config.followRedirects {
		redirectFunc = defaultRedirectFunc
	} else {
		redirectFunc = nil
	}

	clientProfile := config.clientProfile

	client := &http.Client{
		Timeout:       config.timeout,
		Transport:     newRoundTripper(clientProfile, config.transportOptions, config.serverNameOverwrite, config.insecureSkipVerify, config.withRandomTlsExtensionOrder, config.forceHttp1, dialer),
		CheckRedirect: redirectFunc,
	}

	if config.cookieJar != nil {
		client.Jar = config.cookieJar
	}

	return client, clientProfile, nil
}

func (c *httpClient) SetFollowRedirect(followRedirect bool) {
	c.logger.Debug("set follow redirect from %v to %v", c.config.followRedirects, followRedirect)

	c.config.followRedirects = followRedirect
	c.applyFollowRedirect()
}

func (c *httpClient) GetFollowRedirect() bool {
	return c.config.followRedirects
}

func (c *httpClient) applyFollowRedirect() {
	if c.config.followRedirects {
		c.logger.Info("automatic redirect following is enabled")
		c.CheckRedirect = nil
	} else {
		c.logger.Info("automatic redirect following is disabled")
		c.CheckRedirect = defaultRedirectFunc
	}
}

func (c *httpClient) SetProxy(proxyUrl string) error {
	c.logger.Debug("set proxy from %s to %s", c.config.proxyUrl, proxyUrl)
	c.config.proxyUrl = proxyUrl
	c.logger.Info(fmt.Sprintf("set proxy to: %s", proxyUrl))

	return c.applyProxy()
}

func (c *httpClient) GetProxy() string {
	return c.config.proxyUrl
}

func (c *httpClient) applyProxy() error {
	var dialer proxy.ContextDialer
	dialer = proxy.Direct

	if c.config.proxyUrl != "" {
		c.logger.Debug("proxy url %s supplied - using proxy connect dialer", c.config.proxyUrl)
		proxyDialer, err := newConnectDialer(c.config.proxyUrl, c.config.timeout)
		if err != nil {
			c.logger.Error("failed to create proxy connect dialer: %s", err.Error())
			return err
		}

		dialer = proxyDialer
	}

	c.Transport = newRoundTripper(c.config.clientProfile, c.config.transportOptions, c.config.serverNameOverwrite, c.config.insecureSkipVerify, c.config.withRandomTlsExtensionOrder, c.config.forceHttp1, dialer)

	return nil
}

func (c *httpClient) GetCookies(u *url.URL) []*http.Cookie {
	c.logger.Info(fmt.Sprintf("get cookies for url: %s", u.String()))
	if c.Jar == nil {
		c.logger.Warn("you did not setup a cookie jar")
		return nil
	}

	return c.Jar.Cookies(u)
}

func (c *httpClient) SetCookies(u *url.URL, cookies []*http.Cookie) {
	c.logger.Info(fmt.Sprintf("set cookies for url: %s", u.String()))

	if c.Jar == nil {
		c.logger.Warn("you did not setup a cookie jar")
		return
	}

	c.Jar.SetCookies(u, cookies)
}

func (c *httpClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.Client.Do(req)

	if err != nil {
		c.logger.Debug("failed to do request: %s", err.Error())
		return nil, err
	}

	c.logger.Debug("requested %s : status %d", req.URL.String(), resp.StatusCode)

	return resp, nil
}
