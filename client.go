package tls_client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/httputil"
	"github.com/bogdanfinn/tls-client/bandwidth"
	"github.com/bogdanfinn/tls-client/profiles"
	"golang.org/x/net/proxy"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

var defaultRedirectFunc = func(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

// TLSDialerFunc is a function that dials a TLS connection to the given address.
// It's used for WebSocket connections to ensure they use the same TLS fingerprinting
// as regular HTTP requests.
type TLSDialerFunc func(ctx context.Context, network, addr string) (net.Conn, error)

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
	GetDialer() proxy.ContextDialer
	GetTLSDialer() TLSDialerFunc

	AddPreRequestHook(hook PreRequestHookFunc)
	AddPostResponseHook(hook PostResponseHookFunc)
}

// Interface guards are a cheap way to make sure all methods are implemented, this is a static check and does not affect runtime performance.
var _ HttpClient = (*httpClient)(nil)

type httpClient struct {
	http.Client
	logger           Logger
	bandwidthTracker bandwidth.BandwidthTracker
	config           *httpClientConfig
	headerLck        sync.Mutex
	dialer           proxy.ContextDialer

	preHooksLck  sync.RWMutex
	postHooksLck sync.RWMutex
	preHooks     []PreRequestHookFunc
	postHooks    []PostResponseHookFunc
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
		connectHeaders:     make(http.Header),
		clientProfile:      profiles.DefaultClientProfile,
		timeout:            time.Duration(DefaultTimeoutSeconds) * time.Second,
	}

	for _, opt := range options {
		opt(config)
	}

	if err := validateConfig(config); err != nil {
		return nil, err
	}

	client, dialer, bandwidthTracker, clientProfile, err := buildFromConfig(logger, config)
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
		dialer:           dialer,
		preHooksLck:      sync.RWMutex{},
		postHooksLck:     sync.RWMutex{},
		preHooks:         append([]PreRequestHookFunc{}, config.preHooks...),
		postHooks:        append([]PostResponseHookFunc{}, config.postHooks...),
	}, nil
}

func validateConfig(config *httpClientConfig) error {
	if config.enableProtocolRacing && config.disableHttp3 {
		return fmt.Errorf("invalid config: HTTP/3 racing cannot be enabled when HTTP/3 is disabled")
	}

	if config.enableProtocolRacing && config.forceHttp1 {
		return fmt.Errorf("invalid config: HTTP/3 racing cannot be enabled when HTTP/1 is forced")
	}

	if config.disableIPV4 && config.disableIPV6 {
		return fmt.Errorf("invalid config: cannot disable both IPv4 and IPv6")
	}

	if len(config.certificatePins) > 0 && config.insecureSkipVerify {
		return fmt.Errorf("invalid config: certificate pinning cannot be used with insecure skip verify")
	}

	if config.proxyUrl != "" && config.proxyDialerFactory != nil {
		return fmt.Errorf("invalid config: cannot set both proxy URL and custom proxy dialer factory (only one will be used)")
	}

	return nil
}

func buildFromConfig(logger Logger, config *httpClientConfig) (*http.Client, proxy.ContextDialer, bandwidth.BandwidthTracker, profiles.ClientProfile, error) {
	var dialer proxy.ContextDialer
	dialer = newDirectDialer(config.timeout, config.localAddr, config.dialer)

	if config.proxyUrl != "" && config.proxyDialerFactory == nil {
		proxyDialer, err := newConnectDialer(config.proxyUrl, config.timeout, config.localAddr, config.dialer, config.connectHeaders, logger)
		if err != nil {
			return nil, nil, nil, profiles.ClientProfile{}, err
		}

		dialer = proxyDialer
	}

	if config.proxyDialerFactory != nil {
		proxyDialer, err := config.proxyDialerFactory(config.proxyUrl, config.timeout, config.localAddr, config.connectHeaders, logger)
		if err != nil {
			return nil, nil, nil, profiles.ClientProfile{}, err
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

	transport, err := newRoundTripper(clientProfile, config.transportOptions, config.serverNameOverwrite, config.insecureSkipVerify, config.withRandomTlsExtensionOrder, config.forceHttp1, config.disableHttp3, config.enableProtocolRacing, config.certificatePins, config.badPinHandler, config.disableIPV6, config.disableIPV4, bandwidthTracker, dialer)
	if err != nil {
		return nil, nil, nil, clientProfile, err
	}

	client := &http.Client{
		Timeout:       config.timeout,
		Transport:     transport,
		CheckRedirect: redirectFunc,
	}

	if config.cookieJar != nil {
		client.Jar = config.cookieJar
	}

	return client, dialer, bandwidthTracker, clientProfile, nil
}

// CloseIdleConnections closes all idle connections of the underlying http client.
func (c *httpClient) CloseIdleConnections() {
	c.Client.CloseIdleConnections()
}

// GetDialer() returns the underlying Dialer
func (c *httpClient) GetDialer() proxy.ContextDialer {
	return c.dialer
}

// GetTLSDialer returns a TLS dialer function that uses the same TLS fingerprinting
// as regular HTTP requests. This is essential for WebSocket connections to maintain
// consistent fingerprinting.
func (c *httpClient) GetTLSDialer() TLSDialerFunc {
	// Get the roundTripper from the client's transport
	rt, ok := c.Transport.(*roundTripper)
	if !ok {
		// Fallback to a simple TLS dialer if the transport is not a roundTripper
		return func(ctx context.Context, network, addr string) (net.Conn, error) {
			return c.dialer.DialContext(ctx, network, addr)
		}
	}

	// Return a function that uses the roundTripper's dialTLSForWebsocket method
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return rt.dialTLSForWebsocket(ctx, network, addr)
	}
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

	if c.config.proxyUrl != "" && c.config.proxyDialerFactory == nil {
		c.logger.Debug("proxy url %s supplied - using proxy connect dialer", c.config.proxyUrl)
		proxyDialer, err := newConnectDialer(c.config.proxyUrl, c.config.timeout, c.config.localAddr, c.config.dialer, c.config.connectHeaders, c.logger)
		if err != nil {
			c.logger.Error("failed to create proxy connect dialer: %s", err.Error())
			return err
		}

		dialer = proxyDialer
	}

	if c.config.proxyDialerFactory != nil {
		c.logger.Debug("using custom proxy connect dialer")
		proxyDialer, err := c.config.proxyDialerFactory(c.config.proxyUrl, c.config.timeout, c.config.localAddr, c.config.connectHeaders, c.logger)
		if err != nil {
			c.logger.Error("failed to create proxy connect dialer: %s", err.Error())
			return err
		}

		dialer = proxyDialer
	}

	transport, err := newRoundTripper(c.config.clientProfile, c.config.transportOptions, c.config.serverNameOverwrite, c.config.insecureSkipVerify, c.config.withRandomTlsExtensionOrder, c.config.forceHttp1, c.config.disableHttp3, c.config.enableProtocolRacing, c.config.certificatePins, c.config.badPinHandler, c.config.disableIPV6, c.config.disableIPV4, c.bandwidthTracker, dialer)
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

// AddPreRequestHook adds a pre-request hook that is called before each request is sent.
// Multiple hooks can be added and they will be executed in the order they were added.
// If any hook returns an error, the request is aborted and subsequent hooks are not called.
// This method is thread-safe.
func (c *httpClient) AddPreRequestHook(hook PreRequestHookFunc) {
	c.preHooksLck.Lock()
	defer c.preHooksLck.Unlock()
	c.preHooks = append(c.preHooks, hook)
}

// AddPostResponseHook adds a post-response hook that is called after each request completes.
// Multiple hooks can be added and they will be executed in the order they were added.
// All hooks are always executed, even if the request failed or a previous hook panicked.
// This method is thread-safe.
func (c *httpClient) AddPostResponseHook(hook PostResponseHookFunc) {
	c.postHooksLck.Lock()
	defer c.postHooksLck.Unlock()
	c.postHooks = append(c.postHooks, hook)
}

// executePreHooks runs all registered pre-request hooks in order.
// Returns an error if any hook returns an error or panics, aborting subsequent hooks.
func (c *httpClient) executePreHooks(req *http.Request) error {
	c.preHooksLck.RLock()
	hooks := c.preHooks
	c.preHooksLck.RUnlock()

	for _, hook := range hooks {
		if err := c.runPreHook(hook, req); err != nil {
			return err
		}
	}
	return nil
}

func (c *httpClient) runPreHook(hook PreRequestHookFunc, req *http.Request) (err error) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Error("panic in pre-request hook: %v", r)
			err = fmt.Errorf("panic in pre-request hook: %v", r)
		}
	}()
	return hook(req)
}

// executePostHooks runs all registered post-response hooks in order.
// If any hook panics, it is recovered and subsequent hooks are not called.
func (c *httpClient) executePostHooks(originalReq *http.Request, resp *http.Response, requestErr error) {
	c.postHooksLck.RLock()
	hooks := c.postHooks
	c.postHooksLck.RUnlock()

	if len(hooks) == 0 {
		return
	}

	ctx := &PostResponseContext{
		Request:  originalReq,
		Response: resp,
		Error:    requestErr,
	}

	for _, hook := range hooks {
		if err := c.runPostHook(hook, ctx); err != nil {
			return
		}
	}
}

func (c *httpClient) runPostHook(hook PostResponseHookFunc, ctx *PostResponseContext) (err error) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Error("panic in post-response hook: %v", r)
			err = fmt.Errorf("panic in post-response hook: %v", r)
		}
	}()
	hook(ctx)
	return nil
}

// Do issues a given HTTP request and returns the corresponding response.
//
// If the returned error is nil, the response contains a non-nil body, which the user is expected to close.
func (c *httpClient) Do(req *http.Request) (*http.Response, error) {
	if err := c.executePreHooks(req); err != nil {
		return nil, err
	}

	resp, err := c.do(req)

	c.executePostHooks(req, resp, err)

	return resp, err
}

func (c *httpClient) do(req *http.Request) (*http.Response, error) {
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
		req.Header = c.config.defaultHeaders.Clone()
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

			finalResponse := string(buf)

			if c.config.euckrResponse {
				var bufs bytes.Buffer
				wr := transform.NewWriter(&bufs, korean.EUCKR.NewDecoder())
				wr.Write(buf)
				wr.Close()
				finalResponse = bufs.String()
			}

			c.logger.Debug("response body payload: %s", finalResponse)

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
