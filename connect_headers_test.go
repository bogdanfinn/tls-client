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
	"github.com/bogdanfinn/tls-client/profiles"
)

// Chrome 152 sends three headers on a CONNECT to an HTTP proxy, read off the
// wire through a recording proxy on 2026-09-08, from Chrome and Edge alike:
//
//	Host: host:443
//	Proxy-Connection: keep-alive
//	User-Agent: Mozilla/5.0 ...
//
// That order happens to be alphabetical, which is also what the header writer
// falls back to when no order is given, so the test below asks for an order
// that alphabetical would not produce.

// TestConnectHeadersFollowTheGivenOrder holds the CONNECT request against the
// order named under http.HeaderOrderKey, spelled the way a caller would spell
// it. The order is looked up by lower-cased name, and for ordinary requests the
// client lower-cases the given names first; the CONNECT request did not, so a
// capitalised order was ignored and the headers went out alphabetically.
func TestConnectHeadersFollowTheGivenOrder(t *testing.T) {
	proxy := startRecordingProxy(t)
	target := startTargetServer(t)

	got := connectThrough(t, proxy, target, http.Header{
		"User-Agent":        {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36"},
		"Proxy-Connection":  {"keep-alive"},
		http.HeaderOrderKey: {"Host", "User-Agent", "Proxy-Connection"},
	})
	if want := "Host User-Agent Proxy-Connection"; names(got) != want {
		t.Errorf("the CONNECT carried %s, the order asked for %s", names(got), want)
	}
	for _, kv := range got {
		if kv[0] == "User-Agent" && !strings.Contains(kv[1], "Chrome/152.0.0.0") {
			t.Errorf("User-Agent arrived as %q", kv[1])
		}
	}
}

// recordingProxy is an HTTP proxy that keeps the header lines of every
// CONNECT it receives, in arrival order and with their spelling, and then
// tunnels the connection to the target.
type recordingProxy struct {
	ln       net.Listener
	connects chan [][2]string
}

func startRecordingProxy(t *testing.T) *recordingProxy {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	p := &recordingProxy{ln: ln, connects: make(chan [][2]string, 8)}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go p.serve(c)
		}
	}()
	return p
}

func (p *recordingProxy) serve(c net.Conn) {
	defer c.Close()
	c.SetDeadline(time.Now().Add(10 * time.Second))
	br := bufio.NewReader(c)
	line, err := br.ReadString('\n')
	if err != nil {
		return
	}
	var hdrs [][2]string
	for {
		l, err := br.ReadString('\n')
		if err != nil {
			return
		}
		l = strings.TrimRight(l, "\r\n")
		if l == "" {
			break
		}
		name, value, _ := strings.Cut(l, ":")
		hdrs = append(hdrs, [2]string{name, strings.TrimSpace(value)})
	}
	p.connects <- hdrs
	parts := strings.Fields(line)
	if len(parts) < 2 || parts[0] != "CONNECT" {
		io.WriteString(c, "HTTP/1.1 405 Method Not Allowed\r\nContent-Length: 0\r\n\r\n")
		return
	}
	target, err := net.DialTimeout("tcp", parts[1], 5*time.Second)
	if err != nil {
		io.WriteString(c, "HTTP/1.1 502 Bad Gateway\r\nContent-Length: 0\r\n\r\n")
		return
	}
	defer target.Close()
	io.WriteString(c, "HTTP/1.1 200 Connection established\r\n\r\n")
	c.SetDeadline(time.Time{})
	done := make(chan struct{}, 2)
	go func() { io.Copy(target, br); done <- struct{}{} }()
	go func() { io.Copy(c, target); done <- struct{}{} }()
	<-done
}

// startTargetServer answers any HTTP/1.1 request over TLS with an empty 200.
func startTargetServer(t *testing.T) string {
	t.Helper()
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{testCert(t)},
		NextProtos:   []string{"http/1.1"},
	})
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
			go func() {
				defer c.Close()
				c.SetDeadline(time.Now().Add(10 * time.Second))
				br := bufio.NewReader(c)
				for {
					l, err := br.ReadString('\n')
					if err != nil || strings.TrimRight(l, "\r\n") == "" {
						break
					}
				}
				io.WriteString(c, "HTTP/1.1 200 OK\r\nContent-Length: 0\r\nConnection: close\r\n\r\n")
			}()
		}
	}()
	return ln.Addr().String()
}

// connectThrough makes one request through the proxy with the given CONNECT
// headers and returns what the proxy saw on the CONNECT.
func connectThrough(t *testing.T, proxy *recordingProxy, target string, connectHeaders http.Header) [][2]string {
	t.Helper()
	client, err := NewHttpClient(NewNoopLogger(),
		WithClientProfile(profiles.Chrome_152),
		WithProxyUrl("http://"+proxy.ln.Addr().String()),
		WithConnectHeaders(connectHeaders),
		WithInsecureSkipVerify(),
		WithNotFollowRedirects(),
		WithTimeoutSeconds(10),
	)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodGet, "https://"+target+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	select {
	case hdrs := <-proxy.connects:
		return hdrs
	case <-time.After(5 * time.Second):
		t.Fatal("the proxy saw no CONNECT")
		return nil
	}
}

func names(hdrs [][2]string) string {
	out := make([]string, len(hdrs))
	for i, h := range hdrs {
		out[i] = h[0]
	}
	return strings.Join(out, " ")
}
