package tls_client

import (
	"bytes"
	"crypto/tls"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/bogdanfinn/fhttp/http2"
	"github.com/bogdanfinn/fhttp/http2/hpack"
	"github.com/bogdanfinn/tls-client/profiles"
)

// TestHTTP2ReadIdleTimeoutPingsIdleConnection checks that ReadIdleTimeout and
// PingTimeout reach the HTTP/2 transport: an idle connection sends PING frames,
// and a connection whose PING is never acknowledged is closed by the client.
func TestHTTP2ReadIdleTimeoutPingsIdleConnection(t *testing.T) {
	tests := []struct {
		name        string
		readIdle    time.Duration
		ackPings    bool
		wantPing    bool
		wantClosure bool
	}{
		{name: "ping when enabled", readIdle: 100 * time.Millisecond, ackPings: true, wantPing: true},
		{name: "no ping by default", readIdle: 0, ackPings: true},
		{name: "unacknowledged ping closes the connection", readIdle: 100 * time.Millisecond, wantPing: true, wantClosure: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, pings, closed := serveHTTP2(t, tt.ackPings)

			client, err := NewHttpClient(NewNoopLogger(),
				WithClientProfile(profiles.Chrome_133),
				WithInsecureSkipVerify(),
				WithTransportOptions(&TransportOptions{ReadIdleTimeout: tt.readIdle, PingTimeout: 200 * time.Millisecond}),
			)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := client.Get("https://" + addr + "/")
			if err != nil {
				t.Fatal(err)
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()

			select {
			case <-pings:
				if !tt.wantPing {
					t.Fatal("idle connection sent a PING without ReadIdleTimeout")
				}
			case <-time.After(time.Second):
				if tt.wantPing {
					t.Fatal("idle connection sent no PING after ReadIdleTimeout")
				}
			}

			select {
			case <-closed:
				if !tt.wantClosure {
					t.Fatal("client closed a healthy idle connection")
				}
			case <-time.After(time.Second):
				if tt.wantClosure {
					t.Fatal("client kept a connection whose PING was not acknowledged")
				}
			}
		})
	}
}

// serveHTTP2 accepts one HTTP/2 connection, answers every request with an empty
// 200, and reports each PING it receives and the end of the connection.
func serveHTTP2(t *testing.T, ackPings bool) (addr string, pings <-chan struct{}, closed <-chan struct{}) {
	t.Helper()

	listener, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{testCert(t)},
		NextProtos:   []string{http2.NextProtoTLS},
	})
	if err != nil {
		t.Fatal(err)
	}
	var (
		mu       sync.Mutex
		accepted net.Conn
	)
	t.Cleanup(func() {
		_ = listener.Close()
		mu.Lock()
		defer mu.Unlock()
		if accepted != nil {
			_ = accepted.Close()
		}
	})

	pingCh := make(chan struct{}, 1)
	closedCh := make(chan struct{})

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		mu.Lock()
		accepted = conn
		mu.Unlock()
		defer close(closedCh)

		preface := make([]byte, len(http2.ClientPreface))
		if _, err := io.ReadFull(conn, preface); err != nil {
			return
		}

		framer := http2.NewFramer(conn, conn)
		if err := framer.WriteSettings(); err != nil {
			return
		}

		for {
			frame, err := framer.ReadFrame()
			if err != nil {
				return
			}

			switch f := frame.(type) {
			case *http2.SettingsFrame:
				if !f.IsAck() {
					_ = framer.WriteSettingsAck()
				}
			case *http2.HeadersFrame:
				var block bytes.Buffer
				_ = hpack.NewEncoder(&block).WriteField(hpack.HeaderField{Name: ":status", Value: "200"})
				_ = framer.WriteHeaders(http2.HeadersFrameParam{StreamID: f.StreamID, BlockFragment: block.Bytes(), EndHeaders: true, EndStream: true})
			case *http2.PingFrame:
				if f.IsAck() {
					continue
				}
				select {
				case pingCh <- struct{}{}:
				default:
				}
				if ackPings {
					_ = framer.WritePing(true, f.Data)
				}
			}
		}
	}()

	return listener.Addr().String(), pingCh, closedCh
}
