// Package main shows certificate pinning: the request is rejected if the
// server does not present one of the configured public key hashes. Pins
// can be generated with `hpkp-pins -server=host:443`.
//
// Run: go run ./example/certificate_pinning
package main

import (
	"log"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

func main() {
	pins := map[string][]string{
		"bstn.com": {
			"NQvy9sFS99nBqk/nZCUF44hFhshrkvxqYtfrZq3i+Ww=",
			"4a6cPehI7OG6cuDZka5NDZ7FR8a60d3auda+sKfg4Ng=",
			"x4QzPSC810K5/cMjb05Qm4k3Bw5zBn4lTdO/nEW/Td4=",
		},
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_150),
		// DefaultBadPinHandler closes the connection as soon as a pin
		// mismatch is detected. Supply your own func(req) to log/alert instead.
		tls_client.WithCertificatePinning(pins, tls_client.DefaultBadPinHandler),
	}...)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://bstn.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		// A pin mismatch surfaces here as a request error.
		log.Fatal(err)
	}
	defer resp.Body.Close()

	log.Printf("status: %d (certificate pins matched)\n", resp.StatusCode)
}
