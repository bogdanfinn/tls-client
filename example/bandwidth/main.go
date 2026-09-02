// Package main shows the bandwidth tracker: how many bytes the client read
// and wrote on the wire, measured on the connection rather than on the
// response body, so TLS and header overhead are included.
//
// Run: go run ./example/bandwidth
package main

import (
	"io"
	"log"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

func main() {
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_150),
		// Without this option the client uses a no-op tracker that always
		// reports zero.
		tls_client.WithBandwidthTracker(),
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

	// The body has to be drained before reading the counters, otherwise the
	// bytes still in flight are not accounted for yet.
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	tracker := client.GetBandwidthTracker()
	log.Printf("written: %d bytes, read: %d bytes, total: %d bytes\n",
		tracker.GetWriteBytes(), tracker.GetReadBytes(), tracker.GetTotalBandwidth())

	// Reset zeroes the counters, e.g. to measure per request instead of per
	// client lifetime.
	tracker.Reset()
	log.Printf("after reset: total %d bytes\n", tracker.GetTotalBandwidth())
}
