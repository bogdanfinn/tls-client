package tls_client

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"strings"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/http2"
	"github.com/bogdanfinn/tls-client/bandwidth"
	"github.com/bogdanfinn/tls-client/profiles"
)

// portMappingDialer sends the two well-known ports to local test servers, so a
// URL that names no port can be requested without binding 80 or 443. That is
// the case the test needs: before the fix a URL without a port was keyed under
// 443 whichever scheme it had, and only then did the two schemes collide.
type portMappingDialer struct {
	tlsAddr   string
	plainAddr string
}

func (d portMappingDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	switch {
	case strings.HasSuffix(addr, ":443"):
		addr = d.tlsAddr
	case strings.HasSuffix(addr, ":80"):
		addr = d.plainAddr
	}
	var nd net.Dialer

	return nd.DialContext(ctx, network, addr)
}

// TestARedirectToPlainHTTPDoesNotReuseTheTLSTransport covers the report under
// issue #135: an https page answers with a Location pointing at an http URL, and
// the client fails the redirect with "http2: unsupported scheme".
//
// The transport cache is keyed by host and port, and a URL with no port fell
// back to 443 whatever its scheme said. So https://host/ and http://host/ shared
// a key, the redirect found the HTTP/2 transport the first request had cached,
// and that transport refuses a plain http request before dialing anything.
func TestARedirectToPlainHTTPDoesNotReuseTheTLSTransport(t *testing.T) {
	cert := testCert(t)

	// The https side speaks HTTP/2 and answers every request with a redirect to
	// the plain side.
	tlsLn, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{http2.NextProtoTLS},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer tlsLn.Close()

	go func() {
		redirect := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "http://127.0.0.1/plain", http.StatusFound)
		})
		for {
			conn, err := tlsLn.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				tlsConn := conn.(*tls.Conn)
				if err := tlsConn.Handshake(); err != nil {
					_ = conn.Close()

					return
				}
				(&http2.Server{}).ServeConn(conn, &http2.ServeConnOpts{Handler: redirect})
			}(conn)
		}
	}()

	plainLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer plainLn.Close()

	plain := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "plain")
	})}
	go func() { _ = plain.Serve(plainLn) }()
	defer plain.Close()

	dialer := portMappingDialer{tlsAddr: tlsLn.Addr().String(), plainAddr: plainLn.Addr().String()}

	tripper, err := newRoundTripper(profiles.Chrome_133, nil, "", true, false, false, true, false, false,
		nil, nil, false, false, bandwidth.NewNopeTracker(), "", dialer)
	if err != nil {
		t.Fatal(err)
	}
	rt := tripper.(*roundTripper)

	client := &http.Client{Transport: rt}
	defer client.CloseIdleConnections()

	req, err := http.NewRequest(http.MethodGet, "https://127.0.0.1/", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)

	// The control comes first, whatever Do returned: the first hop has to have
	// negotiated HTTP/2 and cached that transport under the key the redirect
	// used to collide with. Without it the redirect could succeed on an HTTP/1
	// transport by accident and the test would say nothing.
	if kind, ok := rt.cachedKind("127.0.0.1:443"); !ok || kind != transportHTTP2 {
		t.Fatalf("the https hop did not leave an HTTP/2 transport under 127.0.0.1:443 (kind=%v ok=%v), so this does not exercise the collision", kind, ok)
	}

	if err != nil {
		t.Fatalf("following the redirect: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK || string(body) != "plain" {
		t.Errorf("got %d %q from %s, want 200 \"plain\" from the plain server", resp.StatusCode, body, resp.Request.URL)
	}
	if resp.Request.URL.Scheme != "http" {
		t.Errorf("the final request went to %s, want the http redirect target", resp.Request.URL)
	}
}

// The cache key is also the address the TLS dial uses, so it has to carry a
// port, and the port has to follow the scheme when the URL leaves it out.
func TestTheTransportKeyDefaultsThePortFromTheScheme(t *testing.T) {
	rt := &roundTripper{}

	tests := []struct {
		url  string
		want string
	}{
		{"https://example.com/", "example.com:443"},
		{"http://example.com/", "example.com:80"},
		{"HTTP://example.com/", "example.com:80"},
		{"https://example.com:8443/", "example.com:8443"},
		{"http://example.com:8080/", "example.com:8080"},
		{"http://[::1]/", "[::1]:80"},
	}

	for _, tt := range tests {
		req, err := http.NewRequest(http.MethodGet, tt.url, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got := rt.getDialTLSAddr(req); got != tt.want {
			t.Errorf("getDialTLSAddr(%s) = %q, want %q", tt.url, got, tt.want)
		}
	}
}
