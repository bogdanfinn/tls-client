package tests

import (
	"crypto/tls"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/http2"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

// TestClient_ConcurrentReconnectsDoNotRaceOnCachedTransports drives the map
// from both sides at once.
//
// cachedTransports is written in two places under two different mutexes.
// RoundTrip reads and writes it under cachedTransportsLck, while dialTLS writes
// it under the roundTripper's own embedded mutex. That is safe only while every
// dialTLS call happens with cachedTransportsLck held, which is true of the
// first dial, since it goes through getTransport, and false of every reconnect:
// dialTLS is the transport's own DialTLSContext, so it runs from inside
// t.RoundTrip, after RoundTrip has released the lock.
//
// A server that negotiates a different protocol on every handshake makes those
// reconnects happen constantly, so a few goroutines are enough to put a write
// from one and a read from another in the same window.
func TestClient_ConcurrentReconnectsDoNotRaceOnCachedTransports(t *testing.T) {
	var handshakes int64

	cert := selfSignedCert(t)
	listener, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{cert},
		GetConfigForClient: func(*tls.ClientHelloInfo) (*tls.Config, error) {
			config := &tls.Config{Certificates: []tls.Certificate{cert}}
			// Alternate on every handshake rather than once, so the cached
			// transport is wrong again as soon as it has been rebuilt.
			if atomic.AddInt64(&handshakes, 1)%2 == 1 {
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
		tls_client.WithTimeoutSeconds(30),
	)
	if err != nil {
		t.Fatal(err)
	}

	endpoint := fmt.Sprintf("https://%s/", listener.Addr().String())

	// The errors are not the point and are deliberately ignored: half of these
	// requests are expected to meet a protocol that has moved. The point is
	// what the race detector sees while they overlap.
	const goroutines, requests = 60, 8

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < requests; j++ {
				req, err := http.NewRequest(http.MethodGet, endpoint, nil)
				if err != nil {
					return
				}
				resp, err := client.Do(req)
				if err == nil {
					resp.Body.Close()
				}
			}
		}()
	}
	wg.Wait()
}
