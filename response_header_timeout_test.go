package tls_client

import (
	"errors"
	"io"
	"net"
	stdhttp "net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bogdanfinn/tls-client/profiles"
)

// TestResponseHeaderTimeout checks that TransportOptions.ResponseHeaderTimeout
// bounds the wait for response headers over both HTTP/1.1 and HTTP/2.
func TestResponseHeaderTimeout(t *testing.T) {
	for _, tt := range []struct {
		name string
		h2   bool
	}{
		{"http2", true},
		{"http1", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			release := make(chan struct{})
			srv := httptest.NewUnstartedServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
				if r.URL.Path == "/slow" {
					<-release
				}
				_, _ = io.WriteString(w, "ok")
			}))
			srv.EnableHTTP2 = tt.h2
			srv.StartTLS()
			t.Cleanup(srv.Close)
			t.Cleanup(func() { close(release) })

			options := []HttpClientOption{
				WithClientProfile(profiles.Chrome_133),
				WithInsecureSkipVerify(),
				WithTimeoutSeconds(3),
				WithTransportOptions(&TransportOptions{ResponseHeaderTimeout: 200 * time.Millisecond}),
			}
			if !tt.h2 {
				options = append(options, WithForceHttp1())
			}
			client, err := NewHttpClient(NewNoopLogger(), options...)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := client.Get(srv.URL + "/fast")
			if err != nil {
				t.Fatalf("prompt response: %v", err)
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			wantProto := "HTTP/1.1"
			if tt.h2 {
				wantProto = "HTTP/2.0"
			}
			if resp.Proto != wantProto {
				t.Fatalf("proto = %s, want %s", resp.Proto, wantProto)
			}

			start := time.Now()
			resp, err = client.Get(srv.URL + "/slow")
			if err == nil {
				_ = resp.Body.Close()
				t.Fatal("delayed response headers: got a response, want a timeout")
			}
			if elapsed := time.Since(start); elapsed > 2*time.Second {
				t.Fatalf("request failed after %v, want about 200ms: %v", elapsed, err)
			}
			var netErr net.Error
			if !errors.As(err, &netErr) || !netErr.Timeout() {
				t.Fatalf("delayed response headers: got %v, want a timeout error", err)
			}
		})
	}
}

// TestHTTP2DisableKeepAlives checks that TransportOptions.DisableKeepAlives
// also stops HTTP/2 connections from being reused.
func TestHTTP2DisableKeepAlives(t *testing.T) {
	for _, tt := range []struct {
		name              string
		disableKeepAlives bool
		wantConns         int64
	}{
		{"keep-alives enabled", false, 1},
		{"keep-alives disabled", true, 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var conns atomic.Int64
			srv := httptest.NewUnstartedServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
				_, _ = io.WriteString(w, "ok")
			}))
			srv.EnableHTTP2 = true
			srv.Config.ConnState = func(_ net.Conn, state stdhttp.ConnState) {
				if state == stdhttp.StateNew {
					conns.Add(1)
				}
			}
			srv.StartTLS()
			t.Cleanup(srv.Close)

			client, err := NewHttpClient(NewNoopLogger(),
				WithClientProfile(profiles.Chrome_133),
				WithInsecureSkipVerify(),
				WithTimeoutSeconds(5),
				WithTransportOptions(&TransportOptions{DisableKeepAlives: tt.disableKeepAlives}),
			)
			if err != nil {
				t.Fatal(err)
			}

			for i := 0; i < 2; i++ {
				resp, err := client.Get(srv.URL + "/")
				if err != nil {
					t.Fatal(err)
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				if resp.Proto != "HTTP/2.0" {
					t.Fatalf("proto = %s, want HTTP/2.0", resp.Proto)
				}
			}

			if got := conns.Load(); got != tt.wantConns {
				t.Fatalf("server saw %d connections, want %d", got, tt.wantConns)
			}
		})
	}
}
