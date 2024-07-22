package tls_client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"sync"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/httputil"
	"github.com/bogdanfinn/tls-client/bandwidth"
	"github.com/bogdanfinn/tls-client/profiles"
	"golang.org/x/net/proxy"
)

var defaultRedirectFunc = func(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

type HttpClient interface {
	GetCookies(u *url.URL) []*http.Cookie
	SetCookies(u *url.URL, cookies []*http.Cookie)
	SetCookieJar(jar http.CookieJar)
	GetCookieJar() http.CookieJar
	SetProxy(proxyUrl string) error
	GetProxy() string
	SetFollowRedirect(followRedirect bool)
	GetFollowRedirect() bool
	CloseIdleConnections()
	Do(req *http.Request) (*http.Response, error)
	Get(url string) (resp *http.Response, err error)
	Head(url string) (resp *http.Response, err error)
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)

	GetBandwidthTracker() bandwidth.BandwidthTracker
}

// Interface guards are a cheap way to make sure all methods are implemented, this is a static check and does not affect runtime performance.
var _ HttpClient = (*httpClient)(nil)

type httpClient struct {
	http.Client
	headerLck sync.Mutex
	logger    Logger
	config    *httpClientConfig

	bandwidthTracker bandwidth.BandwidthTracker
}

var DefaultTimeoutSeconds = 30

var DefaultOptions = []HttpClientOption{
	WithTimeoutSeconds(DefaultTimeoutSeconds),
	WithClientProfile(profiles.DefaultClientProfile),
	WithRandomTLSExtensionOrder(),
	WithNotFollowRedirects(),
}

func ProvideDefaultClient(logger Logger) (HttpClient, error) {
	jar := NewCookieJar()

	return NewHttpClient(logger, append(DefaultOptions, WithCookieJar(jar))...)
}

// NewHttpClient constructs a new HTTP client with the given logger and client options.
func NewHttpClient(logger Logger, options ...HttpClientOption) (HttpClient, error) {
	config := &httpClientConfig{
		followRedirects:    true,
		badPinHandler:      nil,
		customRedirectFunc: nil,
		defaultHeaders:     make(http.Header),
		clientProfile:      profiles.DefaultClientProfile,
		timeout:            time.Duration(DefaultTimeoutSeconds) * time.Second,
	}

	for _, opt := range options {
		opt(config)
	}

	if err := validateConfig(config); err != nil {
		return nil, err
	}

	client, bandwidthTracker, clientProfile, err := buildFromConfig(logger, config)
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
		Client:           *client,
		logger:           logger,
		config:           config,
		headerLck:        sync.Mutex{},
		bandwidthTracker: bandwidthTracker,
	}, nil
}

func validateConfig(_ *httpClientConfig) error {
	return nil
}

func buildFromConfig(logger Logger, config *httpClientConfig) (*http.Client, bandwidth.BandwidthTracker, profiles.ClientProfile, error) {
	var dialer proxy.ContextDialer
	dialer = newDirectDialer(config.timeout, config.localAddr, config.dialer)

	if config.proxyUrl != "" {
		proxyDialer, err := newConnectDialer(config.proxyUrl, config.timeout, config.localAddr, config.dialer, logger, config.userAgent)
		if err != nil {
			return nil, nil, profiles.ClientProfile{}, err
		}

		dialer = proxyDialer
	}

	var redirectFunc func(req *http.Request, via []*http.Request) error
	if !config.followRedirects {
		redirectFunc = defaultRedirectFunc
	} else {
		redirectFunc = nil

		if config.customRedirectFunc != nil {
			redirectFunc = config.customRedirectFunc
		}
	}

	var bandwidthTracker bandwidth.BandwidthTracker
	if config.enabledBandwidthTracker {
		bandwidthTracker = bandwidth.NewTracker()
	} else {
		bandwidthTracker = bandwidth.NewNopeTracker()
	}

	clientProfile := config.clientProfile

	transport, err := newRoundTripper(clientProfile, config.transportOptions, config.serverNameOverwrite, config.insecureSkipVerify, config.withRandomTlsExtensionOrder, config.forceHttp1, config.certificatePins, config.badPinHandler, config.disableIPV6, config.disableIPV4, bandwidthTracker, dialer)
	if err != nil {
		return nil, nil, clientProfile, err
	}

	client := &http.Client{
		Timeout:       config.timeout,
		Transport:     transport,
		CheckRedirect: redirectFunc,
	}

	if config.cookieJar != nil {
		client.Jar = config.cookieJar
	}

	return client, bandwidthTracker, clientProfile, nil
}

// CloseIdleConnections closes all idle connections of the underlying http client.
func (c *httpClient) CloseIdleConnections() {
	c.Client.CloseIdleConnections()
}

// SetFollowRedirect configures the client's HTTP redirect following policy.
func (c *httpClient) SetFollowRedirect(followRedirect bool) {
	c.logger.Debug("set follow redirect from %v to %v", c.config.followRedirects, followRedirect)

	c.config.followRedirects = followRedirect
	c.applyFollowRedirect()
}

// GetFollowRedirect returns the client's HTTP redirect following policy.
func (c *httpClient) GetFollowRedirect() bool {
	return c.config.followRedirects
}

func (c *httpClient) applyFollowRedirect() {
	if c.config.followRedirects {
		c.logger.Debug("automatic redirect following is enabled")
		c.CheckRedirect = nil
	} else {
		c.logger.Debug("automatic redirect following is disabled")
		c.CheckRedirect = defaultRedirectFunc
	}

	if c.config.customRedirectFunc != nil && c.config.followRedirects {
		c.CheckRedirect = c.config.customRedirectFunc
	}
}

// SetProxy configures the client to use the given proxy URL.
//
// proxyUrl should be formatted as:
//
//	"http://user:pass@host:port"
func (c *httpClient) SetProxy(proxyUrl string) error {
	currentProxy := c.config.proxyUrl

	c.logger.Debug("set proxy from %s to %s", c.config.proxyUrl, proxyUrl)
	c.config.proxyUrl = proxyUrl

	err := c.applyProxy()
	if err != nil {
		c.logger.Error("failed to apply new proxy. rolling back to previous used proxy: %w", err)
		c.config.proxyUrl = currentProxy

		return c.applyProxy()
	}

	return nil
}

// GetProxy returns the proxy URL used by the client.
func (c *httpClient) GetProxy() string {
	return c.config.proxyUrl
}

func (c *httpClient) applyProxy() error {
	var dialer proxy.ContextDialer
	dialer = proxy.Direct

	if c.config.proxyUrl != "" {
		c.logger.Debug("proxy url %s supplied - using proxy connect dialer", c.config.proxyUrl)
		proxyDialer, err := newConnectDialer(c.config.proxyUrl, c.config.timeout, c.config.localAddr, c.config.dialer, c.logger, c.config.userAgent)
		if err != nil {
			c.logger.Error("failed to create proxy connect dialer: %s", err.Error())
			return err
		}

		dialer = proxyDialer
	}

	transport, err := newRoundTripper(c.config.clientProfile, c.config.transportOptions, c.config.serverNameOverwrite, c.config.insecureSkipVerify, c.config.withRandomTlsExtensionOrder, c.config.forceHttp1, c.config.certificatePins, c.config.badPinHandler, c.config.disableIPV6, c.config.disableIPV4, c.bandwidthTracker, dialer)
	if err != nil {
		return err
	}

	c.Transport = transport

	return nil
}

// GetCookies returns the cookies in the client's cookie jar for a given URL.
func (c *httpClient) GetCookies(u *url.URL) []*http.Cookie {
	c.logger.Debug(fmt.Sprintf("get cookies for url: %s", u.String()))
	if c.Jar == nil {
		c.logger.Warn("you did not setup a cookie jar")
		return nil
	}

	return c.Jar.Cookies(u)
}

// SetCookies sets a list of cookies for a given URL in the client's cookie jar.
func (c *httpClient) SetCookies(u *url.URL, cookies []*http.Cookie) {
	c.logger.Debug(fmt.Sprintf("set cookies for url: %s", u.String()))

	if c.Jar == nil {
		c.logger.Warn("you did not setup a cookie jar")
		return
	}

	c.Jar.SetCookies(u, cookies)
}

// SetCookieJar sets a jar as the clients cookie jar. This is the recommended way when you want to "clear" the existing cookiejar
func (c *httpClient) SetCookieJar(jar http.CookieJar) {
	c.Jar = jar
}

// GetCookieJar returns the jar the client is currently using
func (c *httpClient) GetCookieJar() http.CookieJar {
	return c.Jar
}

// GetBandwidthTracker returns the bandwidth tracker
func (c *httpClient) GetBandwidthTracker() bandwidth.BandwidthTracker {
	return c.bandwidthTracker
}

// Do issues a given HTTP request and returns the corresponding response.
//
// If the returned error is nil, the response contains a non-nil body, which the user is expected to close.
func (c *httpClient) Do(req *http.Request) (*http.Response, error) {
	if c.config.catchPanics {
		defer func() {
			err := recover()

			if err != nil && c.config.debug {
				c.logger.Debug(fmt.Sprintf("panic occurred in tls client request handling: %s", err))
			}

			if err != nil && !c.config.debug {
				c.logger.Info("critical error during request handling")
			}
		}()
	}

	// Header order must be defined in all lowercase. On HTTP 1 people sometimes define them also in uppercase and then ordering does not work.
	c.headerLck.Lock()

	if len(req.Header) == 0 {
		req.Header = c.config.defaultHeaders
	}

	req.Header[http.HeaderOrderKey] = allToLower(req.Header[http.HeaderOrderKey])
	c.headerLck.Unlock()

	if c.config.debug {
		debugReq := req.Clone(context.Background())

		if req.Body != nil {
			buf, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}

			debugBody := io.NopCloser(bytes.NewBuffer(buf))
			requestBody := io.NopCloser(bytes.NewBuffer(buf))

			c.logger.Debug("request body payload: %s", string(buf))

			debugReq.Body = debugBody
			req.Body = requestBody
		}

		requestBytes, err := httputil.DumpRequestOut(debugReq, debugReq.ContentLength > 0)
		if err != nil {
			return nil, err
		}

		c.logger.Debug("raw request bytes sent over wire: %d (%d kb)", len(requestBytes), len(requestBytes)/1024)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		c.logger.Debug("failed to do request: %s", err.Error())
		return nil, err
	}

	c.logger.Debug("headers on request:\n%v", req.Header)
	c.logger.Debug("cookies on request:\n%v", resp.Request.Cookies())
	c.logger.Debug("headers on response:\n%v", resp.Header)
	c.logger.Debug("cookies on response:\n%v", resp.Cookies())
	c.logger.Debug("requested %s : status %d", req.URL.String(), resp.StatusCode)

	if c.config.debug {
		responseBytes, err := httputil.DumpResponse(resp, resp.ContentLength > 0)
		if err != nil {
			return nil, err
		}

		if resp.Body != nil {
			buf, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			responseBody := io.NopCloser(bytes.NewBuffer(buf))

			c.logger.Debug("response body payload: %s", string(buf))

			resp.Body = responseBody
		}

		c.logger.Debug("raw response bytes received over wire: %d (%d kb)", len(responseBytes), len(responseBytes)/1024)
	}

	return resp, nil
}

func (c *httpClient) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

func (c *httpClient) Head(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

func (c *httpClient) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	return c.Do(req)
}

func allToLower(list []string) []string {
	lower := make([]string, len(list))

	for i, elem := range list {
		lower[i] = strings.ToLower(elem)
	}

	return lower
}
