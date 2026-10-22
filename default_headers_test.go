package tls_client

import (
	"bufio"
	"crypto/tls"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/http2"
	"github.com/bogdanfinn/fhttp/http2/hpack"
	"github.com/bogdanfinn/tls-client/profiles"
)

// What Chrome 152 on Windows sent to a local server for a top-level navigation
// on 2026-09-08, read off the wire: the HPACK block decoded field by field over
// HTTP/2, the raw lines over HTTP/1.1. HOST stands for the address. A client
// given the profile's DefaultHeaders and nothing else has to arrive the same
// way, on both protocols, because what leaves the map is not what reaches the
// wire: the transport writes Host, strips Connection over HTTP/2, and would add
// its own accept-encoding if it did not recognise ours.
var chrome152OverHTTP2 = [][2]string{
	{":method", "GET"}, {":authority", "HOST"}, {":scheme", "https"}, {":path", "/"},
	{"sec-ch-ua", `"Chromium";v="152", "Not?A_Brand";v="24", "Google Chrome";v="152"`},
	{"sec-ch-ua-mobile", "?0"},
	{"sec-ch-ua-platform", `"Windows"`},
	{"upgrade-insecure-requests", "1"},
	{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36"},
	{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
	{"sec-fetch-site", "none"},
	{"sec-fetch-mode", "navigate"},
	{"sec-fetch-user", "?1"},
	{"sec-fetch-dest", "document"},
	{"accept-encoding", "gzip, deflate, br, zstd"},
	{"accept-language", "en-US,en;q=0.9"},
	{"priority", "u=0, i"},
}

var chrome152OverHTTP1 = [][2]string{
	{"Host", "HOST"},
	{"Connection", "keep-alive"},
	{"sec-ch-ua", `"Chromium";v="152", "Not?A_Brand";v="24", "Google Chrome";v="152"`},
	{"sec-ch-ua-mobile", "?0"},
	{"sec-ch-ua-platform", `"Windows"`},
	{"Upgrade-Insecure-Requests", "1"},
	{"User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36"},
	{"Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
	{"Sec-Fetch-Site", "none"},
	{"Sec-Fetch-Mode", "navigate"},
	{"Sec-Fetch-User", "?1"},
	{"Sec-Fetch-Dest", "document"},
	{"Accept-Encoding", "gzip, deflate, br, zstd"},
	{"Accept-Language", "en-US,en;q=0.9"},
}

func TestDefaultHeadersArriveTheWayChromeSendsThem(t *testing.T) {
	headers, ok := profiles.Chrome_152.DefaultHeaders()
	if !ok {
		t.Fatal("Chrome_152 has no default headers")
	}

	t.Run("HTTP/2", func(t *testing.T) {
		got := recordOneRequest(t, "h2", func(addr string) {
			get(t, addr, WithClientProfile(profiles.Chrome_152), WithDefaultHeaders(headers))
		})
		compareWire(t, got, chrome152OverHTTP2)
	})

	t.Run("HTTP/1.1", func(t *testing.T) {
		// The documented step: the browser sends no priority over HTTP/1.1.
		h1 := headers.Clone()
		delete(h1, "priority")
		got := recordOneRequest(t, "http/1.1", func(addr string) {
			get(t, addr, WithClientProfile(profiles.Chrome_152), WithDefaultHeaders(h1), WithForceHttp1())
		})
		compareWire(t, got, chrome152OverHTTP1)
	})
}

// A request with its own headers is left alone: the defaults only stand in for
// an empty set.
func TestDefaultHeadersDoNotTouchARequestThatHasHeaders(t *testing.T) {
	headers, _ := profiles.Chrome_152.DefaultHeaders()
	got := recordOneRequest(t, "h2", func(addr string) {
		client, err := NewHttpClient(NewNoopLogger(), WithClientProfile(profiles.Chrome_152), WithDefaultHeaders(headers), WithInsecureSkipVerify(), WithNotFollowRedirects())
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest(http.MethodGet, "https://"+addr+"/", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header = http.Header{"x-mine": {"1"}}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	})
	names := make([]string, 0, len(got))
	for _, kv := range got {
		names = append(names, kv[0])
	}
	joined := strings.Join(names, " ")
	if !strings.Contains(joined, "x-mine") || strings.Contains(joined, "sec-ch-ua") {
		t.Errorf("the request's own headers were replaced: %s", joined)
	}
}

func get(t *testing.T, addr string, options ...HttpClientOption) {
	t.Helper()
	options = append(options, WithInsecureSkipVerify(), WithNotFollowRedirects(), WithTimeoutSeconds(10))
	client, err := NewHttpClient(NewNoopLogger(), options...)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodGet, "https://"+addr+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func compareWire(t *testing.T, got, want [][2]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%d headers arrived, Chrome sends %d:\n%s", len(got), len(want), render(got))
	}
	for i := range want {
		if i >= len(got) {
			break
		}
		name, value := want[i][0], want[i][1]
		if got[i][0] != name {
			t.Errorf("position %d: %s, Chrome sends %s", i, got[i][0], name)
			continue
		}
		if value != "HOST" && got[i][1] != value {
			t.Errorf("%s: %q, Chrome sends %q", name, got[i][1], value)
		}
	}
}

func render(kvs [][2]string) string {
	var b strings.Builder
	for _, kv := range kvs {
		b.WriteString("  " + kv[0] + ": " + kv[1] + "\n")
	}
	return b.String()
}

// recordOneRequest serves one TLS connection speaking only proto, records the
// first request's headers in arrival order, answers it with an empty 200, and
// returns what arrived. Over HTTP/2 the HPACK block is decoded field by field,
// so pseudo-headers and order survive; over HTTP/1.1 the lines are read raw,
// so the spelling survives too.
func recordOneRequest(t *testing.T, proto string, do func(addr string)) [][2]string {
	t.Helper()
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{testCert(t)},
		NextProtos:   []string{proto},
		MinVersion:   tls.VersionTLS12,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	recorded := make(chan [][2]string, 1)
	failed := make(chan error, 1)
	done := make(chan struct{})
	defer close(done)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			failed <- err
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(10 * time.Second))
		tlsConn := conn.(*tls.Conn)
		if err := tlsConn.Handshake(); err != nil {
			failed <- err
			return
		}
		var hdrs [][2]string
		if tlsConn.ConnectionState().NegotiatedProtocol == "h2" {
			hdrs, err = serveOneHTTP2(tlsConn)
		} else {
			hdrs, err = serveOneHTTP1(tlsConn)
		}
		if err != nil {
			failed <- err
			return
		}
		recorded <- hdrs
		// The client keeps talking after its request, a settings ack for
		// one. Closing with that unread turns the close into a reset, and
		// the client can then lose the response it was about to read. So
		// drain until the test is done with the client, and close then.
		go io.Copy(io.Discard, tlsConn)
		<-done
	}()

	do(ln.Addr().String())

	select {
	case hdrs := <-recorded:
		return hdrs
	case err := <-failed:
		t.Fatalf("the server did not get a request: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("the server did not get a request in time")
	}
	return nil
}

func serveOneHTTP1(conn net.Conn) ([][2]string, error) {
	br := bufio.NewReader(conn)
	if _, err := br.ReadString('\n'); err != nil { // the request line
		return nil, err
	}
	var hdrs [][2]string
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		name, value, _ := strings.Cut(line, ":")
		hdrs = append(hdrs, [2]string{name, strings.TrimSpace(value)})
	}
	_, err := io.WriteString(conn, "HTTP/1.1 200 OK\r\nContent-Length: 0\r\nConnection: close\r\n\r\n")
	return hdrs, err
}

func serveOneHTTP2(conn net.Conn) ([][2]string, error) {
	preface := make([]byte, len(http2.ClientPreface))
	if _, err := io.ReadFull(conn, preface); err != nil {
		return nil, err
	}
	fr := http2.NewFramer(conn, conn)
	if err := fr.WriteSettings(); err != nil {
		return nil, err
	}
	var hdrs [][2]string
	dec := hpack.NewDecoder(65536, func(f hpack.HeaderField) {
		hdrs = append(hdrs, [2]string{f.Name, f.Value})
	})
	for {
		f, err := fr.ReadFrame()
		if err != nil {
			return nil, err
		}
		var block []byte
		var stream uint32
		var ended bool
		switch f := f.(type) {
		case *http2.SettingsFrame:
			if !f.IsAck() {
				if err := fr.WriteSettingsAck(); err != nil {
					return nil, err
				}
			}
			continue
		case *http2.HeadersFrame:
			block, stream, ended = f.HeaderBlockFragment(), f.StreamID, f.HeadersEnded()
		case *http2.ContinuationFrame:
			block, stream, ended = f.HeaderBlockFragment(), f.StreamID, f.HeadersEnded()
		default:
			continue
		}
		if _, err := dec.Write(block); err != nil {
			return nil, err
		}
		if !ended {
			continue
		}
		var buf strings.Builder
		enc := hpack.NewEncoder(&buf)
		enc.WriteField(hpack.HeaderField{Name: ":status", Value: "200"})
		enc.WriteField(hpack.HeaderField{Name: "content-length", Value: "0"})
		err = fr.WriteHeaders(http2.HeadersFrameParam{StreamID: stream, BlockFragment: []byte(buf.String()), EndHeaders: true, EndStream: true})
		return hdrs, err
	}
}
