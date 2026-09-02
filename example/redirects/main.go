// Package main shows how to toggle automatic redirect following at
// runtime, without rebuilding the client.
//
// Run: go run ./example/redirects
package main

import (
	"log"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

func main() {
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_150),
		tls_client.WithNotFollowRedirects(),
	}...)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://httpbin.org/redirect/1", nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	log.Printf("not following redirects: status %d, Location: %s\n", resp.StatusCode, resp.Header.Get("Location"))

	client.SetFollowRedirect(true)

	resp, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	log.Printf("following redirects: final status %d\n", resp.StatusCode)
}
