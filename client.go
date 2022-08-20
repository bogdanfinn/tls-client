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

type HttpClient interface {
	GetCookieJar() http.CookieJar
	GetCookies(u *url.URL) []*http.Cookie
	SetCookies(u *url.URL, cookies []*http.Cookie)
	SetProxy(proxyUrl string) error
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

	client, err := buildFromConfig(config)

	if err != nil {
		return nil, err
	}

	return &httpClient{
		Client: *client,
		logger: logger,
		config: config,
	}, nil
}

func buildFromConfig(config *httpClientConfig) (*http.Client, error) {
	if config.clientProfileSet && config.ja3StringSet {
		return nil, fmt.Errorf("you can not create http client out of clientProfile option and ja3string option. decide for one of them")
	}

	if !config.clientProfileSet && !config.ja3StringSet {
		return nil, fmt.Errorf("you can not create http client without clientProfile option and without ja3string option. decide for one of them")
	}

	var dialer proxy.ContextDialer
	dialer = proxy.Direct

	if config.proxyUrl != "" {
		proxyDialer, err := newConnectDialer(config.proxyUrl)
		if err != nil {
			return nil, err
		}

		dialer = proxyDialer
	}

	var redirectFunc func(req *http.Request, via []*http.Request) error
	if !config.followRedirects {
		redirectFunc = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
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

	if config.clientProfileSet {
		clientProfile = config.clientProfile
	}

	if config.ja3StringSet {
		var decodeErr error
		clientProfile, decodeErr = GetClientProfileFromJa3String(config.ja3String)

		if decodeErr != nil {
			return nil, fmt.Errorf("can not build http client out of ja3 string: %w", decodeErr)
		}
	}

	return &http.Client{
		Jar:           cJar,
		Timeout:       config.timeout,
		Transport:     newRoundTripper(clientProfile, config.insecureSkipVerify, dialer),
		CheckRedirect: redirectFunc,
	}, nil
}

func (c *httpClient) SetProxy(proxyUrl string) error {
	c.config.proxyUrl = proxyUrl

	client, err := buildFromConfig(c.config)

	if err != nil {
		return err
	}

	c.Client = *client
	return nil
}

func (c *httpClient) GetCookies(u *url.URL) []*http.Cookie {
	return c.Jar.Cookies(u)
}

func (c *httpClient) SetCookies(u *url.URL, cookies []*http.Cookie) {
	c.Jar.SetCookies(u, cookies)
}

func (c *httpClient) GetCookieJar() http.CookieJar {
	return c.Jar
}
