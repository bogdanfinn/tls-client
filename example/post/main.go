// Package main shows a POST request with a JSON body.
//
// Run: go run ./example/post
package main

import (
	"io"
	"log"
	"strings"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

func main() {
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_150),
	}...)
	if err != nil {
		log.Fatal(err)
	}

	body := strings.NewReader(`{"foo":"bar"}`)

	req, err := http.NewRequest(http.MethodPost, "https://httpbin.org/post", body)
	if err != nil {
		log.Fatal(err)
	}

	req.Header = http.Header{
		"content-type": {"application/json"},
		"accept":       {"application/json"},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("status: %d\n%s\n", resp.StatusCode, respBody)
}
