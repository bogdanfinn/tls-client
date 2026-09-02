// Package main shows downloading a binary response (e.g. an image) and
// writing it to disk.
//
// Run: go run ./example/download
package main

import (
	"io"
	"log"
	"os"
	"path/filepath"

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

	req, err := http.NewRequest(http.MethodGet, "https://avatars.githubusercontent.com/u/17678241?v=4", nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	dest := filepath.Join(os.TempDir(), "tls-client-example.jpg")

	file, err := os.Create(dest)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	written, err := io.Copy(file, resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("status: %d, wrote %d bytes to %s\n", resp.StatusCode, written, dest)
}
