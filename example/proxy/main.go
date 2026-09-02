// Package main shows how to route requests through a proxy and how to
// rotate to a different proxy on an existing client without rebuilding it.
//
// Run: go run ./example/proxy
package main

import (
	"log"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

func main() {
	// http://, socks5:// and socks5h:// proxy URLs are all supported.
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_150),
		tls_client.WithProxyUrl("http://user:pass@proxy-one.example.com:8000"),
	}...)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://tls.peet.ws/api/all", nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
	log.Printf("request via proxy one: status %d\n", resp.StatusCode)

	// Swap the proxy for the same client - no new client needs to be built.
	if err := client.SetProxy("http://user:pass@proxy-two.example.com:8000"); err != nil {
		log.Fatal(err)
	}

	resp, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
	log.Printf("request via proxy two: status %d\n", resp.StatusCode)
}
