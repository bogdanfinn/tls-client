package tls_client

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/tls-client/profiles"
)

// silentListener accepts connections, reads whatever arrives and never answers.
func silentListener(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var conns []net.Conn
	t.Cleanup(func() {
		ln.Close()
		mu.Lock()
		defer mu.Unlock()
		for _, c := range conns {
			c.Close()
		}
	})

	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			mu.Lock()
			conns = append(conns, c)
			mu.Unlock()
			go io.Copy(io.Discard, c)
		}
	}()

	return ln.Addr().String()
}

// startSOCKS5Proxy is a no-auth SOCKS5 proxy that relays CONNECT requests.
func startSOCKS5Proxy(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })

	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go serveSOCKS5(c)
		}
	}()

	return ln.Addr().String()
}

func serveSOCKS5(c net.Conn) {
	defer c.Close()

	buf := make([]byte, 262)
	if _, err := io.ReadFull(c, buf[:2]); err != nil {
		return
	}
	if _, err := io.ReadFull(c, buf[:buf[1]]); err != nil {
		return
	}
	if _, err := c.Write([]byte{5, 0}); err != nil {
		return
	}

	if _, err := io.ReadFull(c, buf[:4]); err != nil {
		return
	}
	var host string
	switch buf[3] {
	case 1:
		if _, err := io.ReadFull(c, buf[:4]); err != nil {
			return
		}
		host = net.IP(buf[:4]).String()
	case 3:
		if _, err := io.ReadFull(c, buf[:1]); err != nil {
			return
		}
		n := int(buf[0])
		if _, err := io.ReadFull(c, buf[:n]); err != nil {
			return
		}
		host = string(buf[:n])
	default:
		return
	}
	if _, err := io.ReadFull(c, buf[:2]); err != nil {
		return
	}
	port := binary.BigEndian.Uint16(buf[:2])

	target, err := net.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(int(port))))
	if err != nil {
		return
	}
	defer target.Close()

	if _, err := c.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0}); err != nil {
		return
	}

	done := make(chan struct{}, 2)
	go func() { io.Copy(target, c); done <- struct{}{} }()
	go func() { io.Copy(c, target); done <- struct{}{} }()
	<-done
}

type proxyResult struct {
	err     error
	elapsed time.Duration
	hung    bool
}

// proxyGet makes one GET through client, giving up after limit.
func proxyGet(t *testing.T, client HttpClient, ctx context.Context, url string, limit time.Duration) proxyResult {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan proxyResult, 1)
	start := time.Now()
	go func() {
		resp, err := client.Do(req)
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
		done <- proxyResult{err: err, elapsed: time.Since(start)}
	}()

	select {
	case r := <-done:
		return r
	case <-time.After(limit):
		return proxyResult{hung: true, elapsed: limit}
	}
}

func newProxyClient(t *testing.T, proxyURL string, timeoutSeconds int, extra ...HttpClientOption) HttpClient {
	t.Helper()
	options := append([]HttpClientOption{
		WithClientProfile(profiles.Chrome_133),
		WithInsecureSkipVerify(),
		WithTimeoutSeconds(timeoutSeconds),
		WithProxyUrl(proxyURL),
	}, extra...)

	client, err := NewHttpClient(NewNoopLogger(), options...)
	if err != nil {
		t.Fatal(err)
	}

	return client
}

// TestAStalledProxyDialStopsAtTheDeadline checks that a SOCKS or HTTPS proxy
// that never answers fails at the client timeout or the context deadline.
func TestAStalledProxyDialStopsAtTheDeadline(t *testing.T) {
	target := "https://" + startTargetServer(t) + "/"

	for _, scheme := range []string{"socks5", "socks4", "https"} {
		t.Run(scheme+"/client timeout", func(t *testing.T) {
			client := newProxyClient(t, scheme+"://"+silentListener(t), 1)

			r := proxyGet(t, client, context.Background(), target, 5*time.Second)
			if r.hung || r.err == nil || r.elapsed > 2*time.Second {
				t.Fatalf("err=%v after %v (hung=%v), want an error within ~1s", r.err, r.elapsed, r.hung)
			}
		})

		t.Run(scheme+"/context deadline", func(t *testing.T) {
			client := newProxyClient(t, scheme+"://"+silentListener(t), 30)

			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			r := proxyGet(t, client, ctx, target, 5*time.Second)
			if r.hung || r.err == nil || r.elapsed > 1500*time.Millisecond {
				t.Fatalf("err=%v after %v (hung=%v), want an error within ~500ms", r.err, r.elapsed, r.hung)
			}
		})
	}
}

// TestAnHTTPSProxyUsesTheConfiguredDialer checks that HTTPS proxies are dialed
// with WithDialer's dialer.
func TestAnHTTPSProxyUsesTheConfiguredDialer(t *testing.T) {
	var dials atomic.Int32
	dialer := net.Dialer{Control: func(string, string, syscall.RawConn) error {
		dials.Add(1)
		return nil
	}}

	client := newProxyClient(t, "https://"+silentListener(t), 1, WithDialer(dialer))
	proxyGet(t, client, context.Background(), "https://"+startTargetServer(t)+"/", 5*time.Second)

	if dials.Load() == 0 {
		t.Fatal("the HTTPS proxy was dialed without the configured dialer")
	}
}

// TestAStalledCONNECTStopsAtTheContext checks that a CONNECT that is never
// answered stops at the request's deadline or cancellation.
func TestAStalledCONNECTStopsAtTheContext(t *testing.T) {
	target := "https://" + startTargetServer(t) + "/"

	t.Run("deadline", func(t *testing.T) {
		client := newProxyClient(t, "http://"+silentListener(t), 10)

		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		r := proxyGet(t, client, ctx, target, 5*time.Second)
		if r.hung || r.err == nil || r.elapsed > 1500*time.Millisecond {
			t.Fatalf("err=%v after %v (hung=%v), want an error within ~500ms", r.err, r.elapsed, r.hung)
		}
	})

	t.Run("cancel", func(t *testing.T) {
		client := newProxyClient(t, "http://"+silentListener(t), 10)

		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(300*time.Millisecond, cancel)

		r := proxyGet(t, client, ctx, target, 5*time.Second)
		if r.hung || r.err == nil || r.elapsed > 1500*time.Millisecond {
			t.Fatalf("err=%v after %v (hung=%v), want an error soon after the cancel at 300ms", r.err, r.elapsed, r.hung)
		}
	})
}

// TestProxiesWorkWithoutAClientTimeout checks that WithTimeoutSeconds(0) does
// not make every request through an HTTP proxy fail.
func TestProxiesWorkWithoutAClientTimeout(t *testing.T) {
	target := "https://" + startTargetServer(t) + "/"

	for name, proxyURL := range map[string]string{
		"http":   "http://" + startRecordingProxy(t).ln.Addr().String(),
		"socks5": "socks5://" + startSOCKS5Proxy(t),
	} {
		t.Run(name, func(t *testing.T) {
			client := newProxyClient(t, proxyURL, 0)

			if r := proxyGet(t, client, context.Background(), target, 5*time.Second); r.hung || r.err != nil {
				t.Fatalf("err=%v (hung=%v)", r.err, r.hung)
			}
		})
	}
}
