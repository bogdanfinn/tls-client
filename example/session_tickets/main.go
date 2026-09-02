// Package main shows how to turn off TLS session tickets. By default the
// client caches the tickets a server issues and resumes the session on the
// next handshake to that host, which is what a browser does. Disabling them
// forces a full handshake every time.
//
// Run: go run ./example/session_tickets
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
		// Without this the second handshake below resumes the first session
		// via the ticket the server issued.
		tls_client.WithDisableSessionTickets(),
	}...)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://tls.peet.ws/api/all", nil)
	if err != nil {
		log.Fatal(err)
	}

	for i := 1; i <= 2; i++ {
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()

		log.Printf("request %d: status %d\n", i, resp.StatusCode)

		// Drop the pooled connection so the next request has to handshake
		// again - otherwise it just reuses the open one and there is nothing
		// to resume in the first place.
		client.CloseIdleConnections()
	}

	// Whether a handshake was full or resumed is not visible on the response.
	// Add tls_client.WithDebug() to see the handshakes in the client log, or
	// watch for the pre_shared_key extension in a packet capture.
	log.Println("both requests performed a full handshake")
}
