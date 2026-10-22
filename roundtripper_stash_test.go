package tls_client

import (
	"context"
	"io"
	"net"
	stdhttp "net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/tls-client/bandwidth"
	"github.com/bogdanfinn/tls-client/profiles"
)

// stashTestServer starts a local TLS server answering "ok", over HTTP/2 if h2.
func stashTestServer(t *testing.T, h2 bool, configure func(*stdhttp.Server)) *httptest.Server {
	t.Helper()

	srv := httptest.NewUnstartedServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	srv.EnableHTTP2 = h2
	if configure != nil {
		configure(srv.Config)
	}
	srv.StartTLS()
	t.Cleanup(srv.Close)

	return srv
}

type stashOutcome struct {
	resp  *http.Response
	err   error
	panic any
	hung  bool
}

// doWithin runs client.Do, reporting a panic or a hang instead of crashing.
func doWithin(client HttpClient, req *http.Request, limit time.Duration) stashOutcome {
	ch := make(chan stashOutcome, 1)
	go func() {
		var o stashOutcome
		defer func() {
			o.panic = recover()
			ch <- o
		}()
		o.resp, o.err = client.Do(req)
		if o.resp != nil {
			_, _ = io.Copy(io.Discard, o.resp.Body)
			_ = o.resp.Body.Close()
		}
	}()

	select {
	case o := <-ch:
		return o
	case <-time.After(limit):
		return stashOutcome{hung: true}
	}
}

// TestAnInternationalizedHostIsServedOnItsFirstConnection: a non-ASCII host used
// to miss the stash, fail, then panic on the next request with the lock held.
func TestAnInternationalizedHostIsServedOnItsFirstConnection(t *testing.T) {
	for _, tt := range []struct {
		name string
		h2   bool
	}{
		{"http2", true},
		{"http1", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			srv := stashTestServer(t, tt.h2, nil)
			target := srv.Listener.Addr().String()
			_, port, _ := net.SplitHostPort(target)

			var mu sync.Mutex
			var dialed []string

			options := []HttpClientOption{
				WithClientProfile(profiles.Chrome_133),
				WithInsecureSkipVerify(),
				WithTimeoutSeconds(5),
				WithDialContext(func(ctx context.Context, network, addr string) (net.Conn, error) {
					mu.Lock()
					dialed = append(dialed, addr)
					mu.Unlock()

					var d net.Dialer
					return d.DialContext(ctx, network, target)
				}),
			}
			if !tt.h2 {
				options = append(options, WithForceHttp1())
			}

			client, err := NewHttpClient(NewNoopLogger(), options...)
			if err != nil {
				t.Fatal(err)
			}

			wantProto := "HTTP/1.1"
			if tt.h2 {
				wantProto = "HTTP/2.0"
			}

			// A host's first request uses the stash; the next one reuses it.
			for i, step := range []struct {
				host      string
				wantDials []string
			}{
				{"bücher.test", []string{"xn--bcher-kva.test:" + port}},
				{"bücher.test", nil},
				{"plain.test", []string{"plain.test:" + port}},
			} {
				req, err := http.NewRequest(http.MethodGet, "https://"+step.host+":"+port+"/", nil)
				if err != nil {
					t.Fatal(err)
				}

				o := doWithin(client, req, 5*time.Second)
				switch {
				case o.hung:
					t.Fatalf("request %d to %s did not return", i+1, step.host)
				case o.panic != nil:
					t.Fatalf("request %d to %s panicked: %v", i+1, step.host, o.panic)
				case o.err != nil:
					t.Fatalf("request %d to %s: %v", i+1, step.host, o.err)
				case o.resp.Proto != wantProto:
					t.Fatalf("request %d to %s used %s, want %s", i+1, step.host, o.resp.Proto, wantProto)
				}

				mu.Lock()
				got := dialed
				dialed = nil
				mu.Unlock()

				if !slices.Equal(got, step.wantDials) {
					t.Fatalf("request %d to %s dialed %q, want %q", i+1, step.host, got, step.wantDials)
				}
			}
		})
	}
}

// TestARejectedFirstRequestDoesNotLeaveItsConnectionStashed: a request rejected
// before dialing used to leave its stash for the next request to fail on.
func TestARejectedFirstRequestDoesNotLeaveItsConnectionStashed(t *testing.T) {
	srv := stashTestServer(t, false, func(s *stdhttp.Server) {
		s.ReadTimeout = 200 * time.Millisecond
	})

	client, err := NewHttpClient(NewNoopLogger(),
		WithClientProfile(profiles.Chrome_133),
		WithInsecureSkipVerify(),
		WithTimeoutSeconds(5),
	)
	if err != nil {
		t.Fatal(err)
	}
	rt := client.(*httpClient).Transport.(*roundTripper)

	bad, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	bad.Header.Set("X-Bad", "bad\nvalue")

	if o := doWithin(client, bad, 3*time.Second); o.hung || o.panic != nil || o.err == nil {
		t.Fatalf("the invalid header was not rejected: hung=%v panic=%v err=%v", o.hung, o.panic, o.err)
	}

	rt.Lock()
	stashed := len(rt.cachedConnections)
	rt.Unlock()
	if stashed != 0 {
		t.Errorf("%d connection(s) left in the stash after the request was rejected", stashed)
	}

	// Let the server close the idle stashed connection.
	time.Sleep(500 * time.Millisecond)

	good, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if o := doWithin(client, good, 3*time.Second); o.hung || o.panic != nil || o.err != nil {
		t.Fatalf("the next request failed: hung=%v panic=%v err=%v", o.hung, o.panic, o.err)
	}
}

// TestASetupDialClosesAStaleStash: a setup dial used to return a stale stash,
// on which getTransport panicked.
func TestASetupDialClosesAStaleStash(t *testing.T) {
	srv := stashTestServer(t, true, nil)
	addr := srv.Listener.Addr().String()

	tripper, err := newRoundTripper(profiles.Chrome_133, nil, "", true, false, false, true, false, false,
		nil, nil, nil, false, false, bandwidth.NewNopeTracker(), "", 0, newDirectDialer(5*time.Second, nil, net.Dialer{}))
	if err != nil {
		t.Fatal(err)
	}
	rt := tripper.(*roundTripper)

	stale, peer := net.Pipe()
	defer peer.Close()
	rt.cachedConnections[addr] = stale

	req, err := http.NewRequest(http.MethodGet, "https://"+addr+"/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rt.cachedTransportsLck.Lock()
	conn, err := rt.getTransport(req, addr)
	rt.cachedTransportsLck.Unlock()
	if err != nil {
		t.Fatalf("getTransport: %v", err)
	}
	defer conn.Close()

	if conn == stale {
		t.Fatal("the setup dial returned the stale stashed connection")
	}
	if _, err := peer.Write([]byte{0}); err == nil {
		t.Error("the stale stashed connection was not closed")
	}
	if _, ok := rt.cachedTransports[addr]; !ok {
		t.Error("getTransport returned without caching a transport")
	}
}

// TestAPanicInTheSetupDialDoesNotLockTheClient: a panic while building a
// transport used to leave cachedTransportsLck locked.
func TestAPanicInTheSetupDialDoesNotLockTheClient(t *testing.T) {
	srv := stashTestServer(t, true, nil)
	target := srv.Listener.Addr().String()

	var mu sync.Mutex
	calls := 0

	client, err := NewHttpClient(NewNoopLogger(),
		WithClientProfile(profiles.Chrome_133),
		WithInsecureSkipVerify(),
		WithTimeoutSeconds(5),
		WithDialContext(func(ctx context.Context, network, _ string) (net.Conn, error) {
			mu.Lock()
			calls++
			first := calls == 1
			mu.Unlock()
			if first {
				panic("dialer panic")
			}

			var d net.Dialer
			return d.DialContext(ctx, network, target)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	first, err := http.NewRequest(http.MethodGet, "https://"+target+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	if o := doWithin(client, first, 3*time.Second); o.panic == nil {
		t.Fatalf("the dialer's panic did not reach the caller: hung=%v err=%v", o.hung, o.err)
	}

	second, err := http.NewRequest(http.MethodGet, "https://"+target+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	if o := doWithin(client, second, 3*time.Second); o.hung || o.panic != nil || o.err != nil {
		t.Fatalf("the request after the panic failed: hung=%v panic=%v err=%v", o.hung, o.panic, o.err)
	}
}
