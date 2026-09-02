// Package main shows manual response body decompression. The client sets
// accept-encoding and decompresses for you - but only as long as you do not
// set that header yourself. On an HTTP/1.1 connection with your own
// accept-encoding the body arrives compressed and is yours to decode.
//
// Run: go run ./example/decompress
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
		// On HTTP/2 decompression is applied automatically based on the
		// response content-encoding, so force HTTP/1.1 to see the raw body.
		tls_client.WithForceHttp1(),
	}...)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://tls.peet.ws/api/all", nil)
	if err != nil {
		log.Fatal(err)
	}

	// Setting accept-encoding yourself hands the decompression back to you.
	req.Header = http.Header{
		"accept":          {"*/*"},
		"accept-encoding": {"gzip"},
		"user-agent":      {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-encoding",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// DecompressBody picks the decoder from the response content-encoding and
	// returns the body unchanged when it is not compressed.
	decompressed := http.DecompressBody(resp)
	defer decompressed.Close()

	body, err := io.ReadAll(decompressed)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("status: %d, content-encoding: %q, %d bytes decoded\n%s\n",
		resp.StatusCode, resp.Header.Get("content-encoding"), len(body), body)
}
