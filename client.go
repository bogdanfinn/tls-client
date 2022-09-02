package tls_client

import (
	"fmt"
	"io"
	"net/url"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/cookiejar"
	"golang.org/x/net/proxy"
)

var defaultRedirectFunc = func(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

type HttpClient interface {
	GetCookieJar() http.CookieJar
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

var DefaultOptions = []HttpClientOption{
	WithTimeout(30),
	WithClientProfile(DefaultClientProfile),
}

func ProvideDefaultClient(logger Logger) (HttpClient, error) {
	return NewHttpClient(logger, DefaultOptions...)
}

func NewHttpClient(logger Logger, options ...HttpClientOption) (HttpClient, error) {
	config := &httpClientConfig{
		followRedirects: true,
		timeout:         30 * time.Second,
	}

	for _, opt := range options {
		opt(config)
	}

	client, clientProfile, err := buildFromConfig(config)

	if err != nil {
		return nil, err
	}

	config.clientProfile = clientProfile

	return &httpClient{
		Client: *client,
		logger: logger,
		config: config,
	}, nil
}

func buildFromConfig(config *httpClientConfig) (*http.Client, ClientProfile, error) {
	if config.IsClientProfileSet() && config.IsJa3StringSet() {
		return nil, ClientProfile{}, fmt.Errorf("you can not create http client out of clientProfile option and ja3string option. decide for one of them")
	}

	if !config.IsClientProfileSet() && !config.IsJa3StringSet() {
		return nil, ClientProfile{}, fmt.Errorf("you can not create http client without clientProfile option and without ja3string option. decide for one of them")
	}

	var dialer proxy.ContextDialer
	dialer = proxy.Direct

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

	var cJar http.CookieJar

	if config.cookieJar != nil {
		cJar = config.cookieJar
	} else {
		cJar, _ = cookiejar.New(nil)
	}

	var clientProfile ClientProfile

	if config.IsClientProfileSet() {
		clientProfile = config.clientProfile
	}

	if config.IsJa3StringSet() {
		var decodeErr error
		clientProfile, decodeErr = GetClientProfileFromJa3String(config.ja3String)

		if decodeErr != nil {
			return nil, ClientProfile{}, fmt.Errorf("can not build http client out of ja3 string: %w", decodeErr)
		}
	}

	return &http.Client{
		Jar:           cJar,
		Timeout:       config.timeout,
		Transport:     newRoundTripper(clientProfile, config.insecureSkipVerify, dialer),
		CheckRedirect: redirectFunc,
	}, clientProfile, nil
}

func (c *httpClient) SetFollowRedirect(followRedirect bool) {
	c.config.followRedirects = followRedirect
	c.applyFollowRedirect()
}

func (c *httpClient) GetFollowRedirect() bool {
	return c.config.followRedirects
}

func (c *httpClient) applyFollowRedirect() {
	if c.config.followRedirects {
		c.logger.Info("automatic redirect following is disabled")
		c.CheckRedirect = nil
	} else {
		c.logger.Info("automatic redirect following is enabled")
		c.CheckRedirect = defaultRedirectFunc
	}
}

func (c *httpClient) SetProxy(proxyUrl string) error {
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
		proxyDialer, err := newConnectDialer(c.config.proxyUrl, c.config.timeout)
		if err != nil {
			return err
		}

		dialer = proxyDialer
	}

	c.Transport = newRoundTripper(c.config.clientProfile, c.config.insecureSkipVerify, dialer)

	return nil
}

func (c *httpClient) GetCookies(u *url.URL) []*http.Cookie {
	c.logger.Info(fmt.Sprintf("get cookies for url: %s", u.String()))
	return c.Jar.Cookies(u)
}

func (c *httpClient) SetCookies(u *url.URL, cookies []*http.Cookie) {
	c.logger.Info(fmt.Sprintf("set cookies for url: %s", u.String()))
	c.Jar.SetCookies(u, cookies)
}

func (c *httpClient) GetCookieJar() http.CookieJar {
	return c.Jar
}
