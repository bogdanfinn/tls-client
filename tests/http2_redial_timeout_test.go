package tests

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/http2"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/stretchr/testify/assert"
)

// TestClient_HTTP2RedialThroughStalledProxyIsBounded covers an HTTP/2
// connection that has to be dialed again through a proxy which answers CONNECT
// with 200 and then sends nothing.
//
// The HTTP/2 transport makes that dial through dialTLSHTTP2, which used to dial
// with context.Background(). The TLS handshake on the silent tunnel then had no
// deadline, and it held the roundTripper's lock, so the request that started the
// dial and every later request on the same client hung past the client's
// timeout, for as long as the proxy kept the tunnel open.
func TestClient_HTTP2RedialThroughStalledProxyIsBounded(t *testing.T) {
	cert := selfSignedCert(t)
	origin, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{http2.NextProtoTLS},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer origin.Close()

	go serveChangingProtocol(origin)

	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer proxyListener.Close()

	release := make(chan struct{})
	defer close(release)

	go serveStallingProxy(proxyListener, origin.Addr().String(), 2, release)

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(),
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithInsecureSkipVerify(),
		tls_client.WithTimeoutSeconds(1),
		tls_client.WithProxyUrl("http://"+proxyListener.Addr().String()),
	)
	if err != nil {
		t.Fatal(err)
	}

	endpoint := fmt.Sprintf("https://%s/", origin.Addr().String())
	limit := 5 * time.Second

	resp, err := getWithin(t, client, endpoint, limit)
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	_ = resp.Body.Close()
	assert.Equal(t, "HTTP/2.0", resp.Proto)

	// Closing the pooled connection makes the next request dial again from
	// inside the HTTP/2 transport, which is the path under test. That dial is
	// the proxy's second connection, the one it stalls.
	client.CloseIdleConnections()

	resp, err = getWithin(t, client, endpoint, limit)
	if err == nil {
		_ = resp.Body.Close()
		t.Fatal("expected the request through the stalled tunnel to fail")
	}

	// Without the fix the stalled handshake still holds the lock here, so this
	// request never starts its own dial.
	resp, err = getWithin(t, client, endpoint, limit)
	if err != nil {
		t.Fatalf("request after the stalled tunnel: %v", err)
	}
	_ = resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "HTTP/2.0", resp.Proto)
}

// getWithin fails the test if the request has not returned after limit. The
// client's own timeout is shorter, so hitting limit means the client hung.
func getWithin(t *testing.T, client tls_client.HttpClient, url string, limit time.Duration) (*http.Response, error) {
	t.Helper()

	type result struct {
		resp *http.Response
		err  error
	}
	done := make(chan result, 1)

	go func() {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			done <- result{err: err}
			return
		}
		resp, err := client.Do(req)
		done <- result{resp: resp, err: err}
	}()

	select {
	case r := <-done:
		return r.resp, r.err
	case <-time.After(limit):
		t.Fatalf("request still blocked after %s", limit)
		return nil, nil
	}
}

// serveStallingProxy is an HTTP CONNECT proxy that tunnels every connection to
// origin except connection number stall, which it answers with 200 and then
// leaves silent until release is closed.
func serveStallingProxy(listener net.Listener, origin string, stall int64, release <-chan struct{}) {
	var count int64

	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		n := atomic.AddInt64(&count, 1)

		go func(conn net.Conn) {
			defer conn.Close()

			reader := bufio.NewReader(conn)
			if _, err := http.ReadRequest(reader); err != nil {
				return
			}
			if _, err := io.WriteString(conn, "HTTP/1.1 200 Connection established\r\n\r\n"); err != nil {
				return
			}

			if n == stall {
				<-release
				return
			}

			upstream, err := net.Dial("tcp", origin)
			if err != nil {
				return
			}
			defer upstream.Close()

			go func() {
				_, _ = io.Copy(upstream, reader)
				_ = upstream.Close()
			}()
			_, _ = io.Copy(conn, upstream)
		}(conn)
	}
}
