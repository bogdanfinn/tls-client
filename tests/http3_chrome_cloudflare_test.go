package tests

import (
	"io"
	"strings"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

func TestHTTP3WithChromeOnCloudflare(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithProtocolRacing(),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://www.cloudflare.com/cdn-cgi/trace", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultHeader

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(body), "http=http/3") {
		t.Fatal("Response did not contain HTTP3 result")
	}
}

func TestHTTP2WithChromeOnCloudflare(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		// tls_client.WithProtocolRacing(), // we explicitly disable racing and stick to the "old" behavior
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://www.cloudflare.com/cdn-cgi/trace", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultHeader

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(body), "http=http/2") {
		t.Fatal("Response did contain HTTP3 result")
	}
}
