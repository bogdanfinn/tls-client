package tests

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
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

// TestClient_ProtocolChangeBetweenConnections covers a server that negotiates
// http/1.1 on one connection and h2 on the next, which is ordinary behaviour
// for an address behind a load balancer.
//
// The transport is cached per address. Before this was fixed, the cached
// transport was reused on reconnect without checking the protocol that had just
// been negotiated, so an HTTP/1 transport was handed an HTTP/2 connection and
// read the SETTINGS frame as a status line:
//
//	malformed HTTP response "\x00\x00\x1e\x04\x00..."
//
// The address stayed broken until the whole client was replaced.
//
// The connection on which the change is discovered cannot be rescued, since
// that dial belongs to the transport speaking the wrong protocol. So what this
// checks is that the client RECOVERS: the request that runs into the change
// fails with a message saying what happened, and the one after it succeeds over
// the protocol the server is now offering. Without the fix the third request
// fails exactly like the second, and so does every one after it.
func TestClient_ProtocolChangeBetweenConnections(t *testing.T) {
	var handshakes int64

	cert := selfSignedCert(t)
	listener, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{cert},
		GetConfigForClient: func(*tls.ClientHelloInfo) (*tls.Config, error) {
			config := &tls.Config{Certificates: []tls.Certificate{cert}}
			if atomic.AddInt64(&handshakes, 1) == 1 {
				config.NextProtos = []string{"http/1.1"}
			} else {
				config.NextProtos = []string{http2.NextProtoTLS}
			}
			return config, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go serveChangingProtocol(listener)

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(),
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithInsecureSkipVerify(),
		tls_client.WithTimeoutSeconds(10),
	)
	if err != nil {
		t.Fatal(err)
	}

	endpoint := fmt.Sprintf("https://%s/", listener.Addr().String())

	// The first request negotiates http/1.1 and caches an HTTP/1 transport.
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "HTTP/1.1", resp.Proto)

	// The server closed that connection, so the second request dials again and
	// this time negotiates h2. The cached HTTP/1 transport cannot use it, so
	// this request fails and the cache entry is dropped.
	req2, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp2, err := client.Do(req2)
	if err == nil {
		resp2.Body.Close()
		t.Fatal("expected the request that runs into the protocol change to fail")
	}
	assert.Contains(t, err.Error(), "cached transport speaks")

	// The third request is the point of the fix. Without it this one, and every
	// one after it, fails the same way as the second.
	req3, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp3, err := client.Do(req3)
	if err != nil {
		t.Fatalf("third request, after the cached transport was dropped: %v", err)
	}
	defer resp3.Body.Close()

	assert.Equal(t, 200, resp3.StatusCode)
	assert.Equal(t, "HTTP/2.0", resp3.Proto)
}

// serveChangingProtocol answers each connection with whichever protocol its
// handshake settled on, and closes an HTTP/1 connection after one response so
// the client has to dial again.
func serveChangingProtocol(listener net.Listener) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "ok")
	})

	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		go func(conn net.Conn) {
			tlsConn, ok := conn.(*tls.Conn)
			if !ok {
				_ = conn.Close()
				return
			}
			if err := tlsConn.Handshake(); err != nil {
				_ = conn.Close()
				return
			}

			if tlsConn.ConnectionState().NegotiatedProtocol == http2.NextProtoTLS {
				(&http2.Server{}).ServeConn(conn, &http2.ServeConnOpts{Handler: handler})
				return
			}

			server := &http.Server{Handler: handler}
			_ = server.Serve(&singleConnListener{Conn: conn})
		}(conn)
	}
}

// singleConnListener hands one already accepted connection to an http.Server
// and then reports that it is finished, so the server serves exactly that
// connection and closes it.
type singleConnListener struct {
	net.Conn
	used bool
}

func (l *singleConnListener) Accept() (net.Conn, error) {
	if l.used {
		return nil, fmt.Errorf("no more connections")
	}
	l.used = true
	return l.Conn, nil
}

func (l *singleConnListener) Close() error   { return nil }
func (l *singleConnListener) Addr() net.Addr { return l.Conn.LocalAddr() }

func selfSignedCert(t *testing.T) tls.Certificate {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "localhost"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}

	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
}
