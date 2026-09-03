package tls_client

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"testing"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/http2"
	"github.com/bogdanfinn/tls-client/bandwidth"
	"github.com/bogdanfinn/tls-client/profiles"
	"golang.org/x/net/proxy"
)

// TestGetTransportDoesNotPanicWhenOnlyTheKindSurvivedADrop covers the window
// inside dropCachedTransport.
//
// It takes cachedTransportsLck to delete the transport and cachedKindsLck to
// delete the kind, one after the other, so in between an address has a kind and
// no transport. A RoundTrip landing there finds nothing cached, calls
// getTransport, and the setup dial used to see the surviving kind, decide the
// connection was reusable and return it with a nil error. getTransport panics
// on exactly that.
//
// The state is set up directly rather than raced into, because a window between
// two mutex releases is not something a test can hit on purpose.
func TestGetTransportDoesNotPanicWhenOnlyTheKindSurvivedADrop(t *testing.T) {
	cert := testCert(t)
	listener, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{http2.NextProtoTLS},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				if tlsConn, ok := c.(*tls.Conn); ok {
					_ = tlsConn.Handshake()
				}
			}(conn)
		}
	}()

	addr := listener.Addr().String()

	tripper, err := newRoundTripper(profiles.Chrome_133, nil, "", true, false, false, true, false, false,
		nil, nil, false, false, bandwidth.NewNopeTracker(), "", proxy.Direct)
	if err != nil {
		t.Fatal(err)
	}
	rt := tripper.(*roundTripper)

	// The state the drop leaves behind for a moment: a kind, and no transport.
	rt.setCachedKind(addr, transportHTTP2)

	req, err := http.NewRequest(http.MethodGet, "https://"+addr+"/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Control: the state really is the one described, or the call below proves
	// nothing about it.
	if _, ok := rt.cachedKind(addr); !ok {
		t.Fatal("no cached kind, so this does not exercise the window")
	}
	rt.cachedTransportsLck.Lock()
	_, cached := rt.cachedTransports[addr]
	rt.cachedTransportsLck.Unlock()
	if cached {
		t.Fatal("a transport is already cached, so getTransport would not be called")
	}

	rt.cachedTransportsLck.Lock()
	defer rt.cachedTransportsLck.Unlock()

	// A panic here is the failure. An error is fine: the point is that the setup
	// dial does not report success without caching a transport.
	if err := rt.getTransport(req, addr); err != nil {
		t.Fatalf("getTransport: %v", err)
	}

	if _, ok := rt.cachedTransports[addr]; !ok {
		t.Error("getTransport returned without caching a transport")
	}
}

func testCert(t *testing.T) tls.Certificate {
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
