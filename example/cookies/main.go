// Package main shows automatic cookie persistence via the cookie jar, plus
// reading and manually seeding cookies for a URL.
//
// Run: go run ./example/cookies
package main

import (
	"log"
	"net/url"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

func main() {
	jar := tls_client.NewCookieJar()

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_150),
		tls_client.WithCookieJar(jar),
	}...)
	if err != nil {
		log.Fatal(err)
	}

	// The server sets a cookie on this request; the jar stores it
	// automatically and replays it on every later request to the domain.
	req, err := http.NewRequest(http.MethodGet, "https://httpbin.org/cookies/set?session=abc123", nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	u, _ := url.Parse("https://httpbin.org")
	log.Printf("cookies stored for %s: %v\n", u, client.GetCookies(u))

	// You can also seed cookies manually without a prior request.
	client.SetCookies(u, []*http.Cookie{{Name: "manual", Value: "value"}})
	log.Printf("cookies after manual seeding: %v\n", client.GetCookies(u))
}
