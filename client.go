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

var DefaultTimeoutSeconds = 30

var DefaultOptions = []HttpClientOption{
	WithTimeout(DefaultTimeoutSeconds),
	WithClientProfile(DefaultClientProfile),
	WithNotFollowRedirects(),
}

func ProvideDefaultClient(logger Logger) (HttpClient, error) {
	return NewHttpClient(logger, DefaultOptions...)
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

	clientProfile := config.clientProfile

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
		c.logger.Info("automatic redirect following is enabled")
		c.CheckRedirect = nil
	} else {
		c.logger.Info("automatic redirect following is disabled")
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

	var filteredCookies []*http.Cookie

	existingCookies := c.Jar.Cookies(u)

	for _, cookie := range cookies {
		alreadyInJar := false

		for _, existingCookie := range existingCookies {
			alreadyInJar = cookie.Name == existingCookie.Name

			if alreadyInJar {
				break
			}
		}

		if alreadyInJar {
			continue
		}

		filteredCookies = append(filteredCookies, cookie)
	}

	c.Jar.SetCookies(u, filteredCookies)
}

func (c *httpClient) GetCookieJar() http.CookieJar {
	return c.Jar
}
