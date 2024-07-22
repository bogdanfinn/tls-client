package tls_client

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"sync"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"golang.org/x/net/proxy"

	"github.com/bogdanfinn/fhttp/http2"
)

type directDialer struct {
	dialer net.Dialer
}

func newDirectDialer(timeout time.Duration, localAddr *net.TCPAddr, _dialer net.Dialer) proxy.ContextDialer {
	_dialer.Timeout = timeout
	if nil != localAddr {
		_dialer.LocalAddr = localAddr
	}

	return &directDialer{
		dialer: _dialer,
	}
}

func (d *directDialer) Dial(network, addr string) (net.Conn, error) {
	return d.dialer.Dial(network, addr)
}

func (d *directDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return d.dialer.DialContext(ctx, network, addr)
}

type socksContextDialer struct {
	socksDialer proxy.Dialer
}

func newSocksContextDialer(socksDialer proxy.Dialer) socksContextDialer {
	return socksContextDialer{
		socksDialer: socksDialer,
	}
}

func (s *socksContextDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return s.socksDialer.Dial(network, address)
}

// Copyright 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// stolen from https://github.com/caddyserver/forwardproxy/blob/master/httpclient/httpclient.go

// connectDialer allows to configure one-time use HTTP CONNECT client
type connectDialer struct {
	logger        Logger
	ProxyUrl      url.URL
	DefaultHeader http.Header

	Dialer net.Dialer // overridden dialer allow to control establishment of TCP connection

	// overridden DialTLS allows user to control establishment of TLS connection
	// MUST return connection with completed Handshake, and NegotiatedProtocol
	DialTLS            func(network string, address string) (net.Conn, string, error)
	Timeout            time.Duration
	EnableH2ConnReuse  bool
	cacheH2Mu          sync.Mutex
	cachedH2ClientConn *http2.ClientConn
	cachedH2RawConn    net.Conn
}

// newConnectDialer creates a dialer to issue CONNECT requests and tunnel traffic via HTTP/S proxy.
// proxyUrlStr must provide Scheme and Host, may provide credentials and port.
// Example: https://username:password@golang.org:443
func newConnectDialer(proxyUrlStr string, timeout time.Duration, localAddr *net.TCPAddr, configDialer net.Dialer, logger Logger, userAgent string) (proxy.ContextDialer, error) {
	proxyUrl, err := url.Parse(proxyUrlStr)
	if err != nil {
		return nil, err
	}

	if proxyUrl.Host == "" {
		return nil, errors.New("invalid url `" + proxyUrlStr +
			"`, make sure to specify full url like https://username:password@hostname.com:443/")
	}

	switch proxyUrl.Scheme {
	case "http":
		if proxyUrl.Port() == "" {
			proxyUrl.Host = net.JoinHostPort(proxyUrl.Host, "80")
		}
	case "https":
		if proxyUrl.Port() == "" {
			proxyUrl.Host = net.JoinHostPort(proxyUrl.Host, "443")
		}
	case "socks5", "socks5h":
		return handleSocks5ProxyDialer(proxyUrl, localAddr)
	case "":
		return nil, errors.New("specify scheme explicitly (https://)")
	default:
		return nil, errors.New("scheme " + proxyUrl.Scheme + " is not supported")
	}
	_dialer := configDialer
	_dialer.Timeout = timeout
	if nil != localAddr {
		_dialer.LocalAddr = localAddr
	}

	dialer := &connectDialer{
		logger:            logger,
		ProxyUrl:          *proxyUrl,
		Dialer:            _dialer,
		Timeout:           timeout,
		DefaultHeader:     make(http.Header),
		EnableH2ConnReuse: true,
	}

	if proxyUrl.User != nil {
		if proxyUrl.User.Username() != "" {
			//example format (with credentials): http://root:xxx@127.0.0.1:12312
			password, _ := proxyUrl.User.Password()
			dialer.DefaultHeader.Set("Proxy-Authorization", "Basic "+
				base64.StdEncoding.EncodeToString([]byte(proxyUrl.User.Username()+":"+password)))

		}
	} else {
		//example format (without credentials): http://127.0.0.1:12312
		dialer.DefaultHeader.Set("Proxy-Authorization", "")
	}
	if userAgent != "" {
		dialer.DefaultHeader.Set("User-Agent", userAgent)
	}
	return dialer, nil
}

func handleSocks5ProxyDialer(proxyUrl *url.URL, localAddr *net.TCPAddr) (proxy.ContextDialer, error) {
	var proxyAuth *proxy.Auth
	if proxyUrl.User != nil {
		password, _ := proxyUrl.User.Password()
		proxyAuth = &proxy.Auth{
			User:     proxyUrl.User.Username(),
			Password: password,
		}
	} else {
		proxyAuth = nil
	}
	_dialer := proxy.Dialer(proxy.Direct)
	if nil != localAddr {
		_dialer = &net.Dialer{
			LocalAddr: localAddr,
		}
	}
	socksDialer, err := proxy.SOCKS5("tcp", proxyUrl.Host, proxyAuth, _dialer)
	if err != nil {
		return nil, fmt.Errorf("failed to create socks5 proxy: %w", err)
	}

	scd := newSocksContextDialer(socksDialer)

	return &scd, nil
}

func (c *connectDialer) Dial(network, address string) (net.Conn, error) {
	return c.DialContext(context.Background(), network, address)
}

// Users of context.WithValue should define their own types for keys
type ContextKeyHeader struct{}

// ctx.Value will be inspected for optional ContextKeyHeader{} key, with `http.Header` value,
// which will be added to outgoing request headers, overriding any colliding c.DefaultHeader
func (c *connectDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	req := (&http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Host: address},
		Header: make(http.Header),
		Host:   address,
	}).WithContext(ctx)

	for k, v := range c.DefaultHeader {
		req.Header[k] = v
	}

	if ctxHeader, ctxHasHeader := ctx.Value(ContextKeyHeader{}).(http.Header); ctxHasHeader {
		for k, v := range ctxHeader {
			req.Header[k] = v
		}
	}

	connectHttp2 := func(rawConn net.Conn, h2clientConn *http2.ClientConn) (net.Conn, error) {
		req.Proto = "HTTP/2.0"
		req.ProtoMajor = 2
		req.ProtoMinor = 0
		pr, pw := io.Pipe()
		req.Body = pr

		resp, err := h2clientConn.RoundTrip(req)
		if err != nil {
			_ = rawConn.Close()
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			_ = rawConn.Close()
			return nil, errors.New("Proxy responded with non 200 code: " + resp.Status)
		}

		return newHttp2Conn(rawConn, pw, resp.Body), nil
	}

	connectHttp1 := func(rawConn net.Conn) (net.Conn, error) {
		req.Proto = "HTTP/1.1"
		req.ProtoMajor = 1
		req.ProtoMinor = 1

		deadline := time.Now().Add(c.Timeout)
		err := rawConn.SetDeadline(deadline)
		if err != nil {
			_ = rawConn.Close()
			return nil, err
		}

		err = req.Write(rawConn)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				c.logger.Error("deadline exceeded while trying to write proxy connection")
			}

			_ = rawConn.Close()
			return nil, err
		}

		resp, err := http.ReadResponse(bufio.NewReader(rawConn), req)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				c.logger.Error("deadline exceeded while trying to read proxy connection")
			}

			_ = rawConn.Close()
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			_ = rawConn.Close()
			return nil, errors.New("Proxy responded with non 200 code: " + resp.Status)
		}

		rawConn.SetDeadline(time.Time{})

		return rawConn, nil
	}

	if c.EnableH2ConnReuse {
		c.cacheH2Mu.Lock()
		unlocked := false
		if c.cachedH2ClientConn != nil && c.cachedH2RawConn != nil {
			if c.cachedH2ClientConn.CanTakeNewRequest() {
				rc := c.cachedH2RawConn
				cc := c.cachedH2ClientConn
				c.cacheH2Mu.Unlock()
				unlocked = true

				proxyConn, err := connectHttp2(rc, cc)
				if err == nil {
					return proxyConn, err
				}
				// else: carry on and try again
			}
		}
		if !unlocked {
			c.cacheH2Mu.Unlock()
		}
	}

	var err error
	var rawConn net.Conn
	negotiatedProtocol := ""
	switch c.ProxyUrl.Scheme {
	case "http":
		rawConn, err = c.Dialer.DialContext(ctx, network, c.ProxyUrl.Host)

		if err != nil {
			return nil, err
		}
	case "https":
		if c.DialTLS != nil {
			rawConn, negotiatedProtocol, err = c.DialTLS(network, c.ProxyUrl.Host)
			if err != nil {
				return nil, err
			}
		} else {
			tlsConf := tls.Config{
				NextProtos: []string{"h2", "http/1.1"},
				ServerName: c.ProxyUrl.Hostname(),
			}
			tlsConn, err := tls.Dial(network, c.ProxyUrl.Host, &tlsConf)
			if err != nil {
				return nil, err
			}
			err = tlsConn.HandshakeContext(ctx)
			if err != nil {
				return nil, err
			}
			negotiatedProtocol = tlsConn.ConnectionState().NegotiatedProtocol
			rawConn = tlsConn
		}
	default:
		return nil, errors.New("scheme " + c.ProxyUrl.Scheme + " is not supported")
	}

	switch negotiatedProtocol {
	case "":
		fallthrough
	case "http/1.1":
		return connectHttp1(rawConn)
	case "h2":
		t := http2.Transport{}
		h2clientConn, err := t.NewClientConn(rawConn)
		if err != nil {
			_ = rawConn.Close()
			return nil, err
		}

		proxyConn, err := connectHttp2(rawConn, h2clientConn)
		if err != nil {
			_ = rawConn.Close()
			return nil, err
		}

		if c.EnableH2ConnReuse {
			c.cacheH2Mu.Lock()
			c.cachedH2ClientConn = h2clientConn
			c.cachedH2RawConn = rawConn
			c.cacheH2Mu.Unlock()
		}
		return proxyConn, err
	default:
		_ = rawConn.Close()
		return nil, errors.New("negotiated unsupported application layer protocol: " +
			negotiatedProtocol)
	}
}

func newHttp2Conn(c net.Conn, pipedReqBody *io.PipeWriter, respBody io.ReadCloser) net.Conn {
	return &http2Conn{Conn: c, in: pipedReqBody, out: respBody}
}

type http2Conn struct {
	net.Conn
	in  *io.PipeWriter
	out io.ReadCloser
}

func (h *http2Conn) Read(p []byte) (n int, err error) {
	return h.out.Read(p)
}

func (h *http2Conn) Write(p []byte) (n int, err error) {
	return h.in.Write(p)
}

func (h *http2Conn) Close() error {
	var retErr error = nil
	if err := h.in.Close(); err != nil {
		retErr = err
	}
	if err := h.out.Close(); err != nil {
		retErr = err
	}
	return retErr
}

func (h *http2Conn) CloseConn() error {
	return h.Conn.Close()
}

func (h *http2Conn) CloseWrite() error {
	return h.in.Close()
}

func (h *http2Conn) CloseRead() error {
	return h.out.Close()
}
