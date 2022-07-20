package tls_client

import (
	"io"
	"net/url"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/cookiejar"
	"golang.org/x/net/proxy"
)

type HttpClient interface {
	GetCookies(u *url.URL) []*http.Cookie
	SetCookies(u *url.URL, cookies []*http.Cookie)
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

	cJar, _ := cookiejar.New(nil)

	return &http.Client{
		Jar:           cJar,
		Timeout:       config.timeout,
		Transport:     newRoundTripper(config.clientProfile.clientHelloId, config.clientProfile.settings, config.clientProfile.settingsOrder, config.clientProfile.pseudoHeaderOrder, config.clientProfile.priorities, config.clientProfile.connectionFlow, config.insecureSkipVerify, dialer),
		CheckRedirect: redirectFunc,
	}, nil
}

func (c *httpClient) GetCookies(u *url.URL) []*http.Cookie {
	return c.Jar.Cookies(u)
}

func (c *httpClient) SetCookies(u *url.URL, cookies []*http.Cookie) {
	c.Jar.SetCookies(u, cookies)
}
