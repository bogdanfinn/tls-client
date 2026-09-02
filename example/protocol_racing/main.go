// Package main shows protocol racing: HTTP/3 (QUIC) and HTTP/2 (TCP) are
// started in parallel and whichever connects first serves the request,
// similar to Chrome's "Happy Eyeballs" approach. The winning protocol is
// cached per host and reused on later requests.
//
// Run: go run ./example/protocol_racing
package main

import (
	"log"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

func main() {
	// Only chrome_144, chrome_144_PSK, firefox_147, firefox_147_PSK and
	// firefox_148 carry HTTP/3 settings. Every other profile falls back to a
	// minimal SETTINGS frame, so its HTTP/3 fingerprint does not resemble the
	// browser it imitates. Pick one of those five when the HTTP/3 fingerprint
	// matters, or turn HTTP/3 off with WithDisableHttp3().
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_144),
		tls_client.WithProtocolRacing(),
	}...)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://cloudflare-quic.com/", nil)
	if err != nil {
		log.Fatal(err)
	}

	// First request: both legs are started, HTTP/2 delayed by 300ms to give
	// HTTP/3 a head start.
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
	log.Printf("first request: status %d over %s\n", resp.StatusCode, resp.Proto)

	// Second request: the protocol that won the race is cached for the host
	// and used directly, no racing.
	resp, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
	log.Printf("second request: status %d over %s\n", resp.StatusCode, resp.Proto)

	// The HTTP/3 leg is UDP, and of the supported proxy schemes only SOCKS5
	// can carry UDP (via UDP ASSOCIATE). Combining racing with any other
	// scheme is rejected instead of silently sending the HTTP/3 leg outside
	// the proxy and leaking the real IP address.
	if err := client.SetProxy("http://user:pass@proxy.example.com:8000"); err != nil {
		log.Printf("expected rejection: %v\n", err)
	}

	// A socks5:// proxy is accepted - the proxy server still has to support
	// UDP ASSOCIATE, which not every provider enables.
	if err := client.SetProxy("socks5://user:pass@proxy.example.com:1080"); err != nil {
		log.Fatal(err)
	}
	log.Printf("proxy accepted: %s\n", client.GetProxy())
}
