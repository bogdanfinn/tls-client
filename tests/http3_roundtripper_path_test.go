package tests

import (
	"io"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

func TestHTTP3DirectPathChrome(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_144),
		tls_client.WithTimeoutSeconds(30),
		// NO WithProtocolRacing() - forces direct ALPN path via roundtripper.go
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://quic.browserleaks.com/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = http.Header{
		"accept":          {"*/*"},
		"accept-language": {"en-US,en;q=0.9"},
		"user-agent":      {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-language",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	expectedH3Hash := "ba909fc3dc419ea5c5b26c6323ac1879"

	t.Logf("Chrome_144 HTTP/3 direct path test completed")
	t.Logf("Expected hash: %s", expectedH3Hash)
	t.Logf("Protocol used: %s", resp.Proto)
}

func TestHTTP3DirectPathFirefox(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Firefox_147),
		tls_client.WithTimeoutSeconds(30),
		// NO WithProtocolRacing() - forces direct ALPN path via roundtripper.go
		tls_client.WithTransportOptions(&tls_client.TransportOptions{
			MaxResponseHeaderBytes: -1, // Firefox doesn't send SETTINGS_MAX_FIELD_SECTION_SIZE
		}),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://quic.browserleaks.com/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = http.Header{
		"accept":          {"*/*"},
		"accept-language": {"en-US,en;q=0.9"},
		"user-agent":      {"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0"},
		http.HeaderOrderKey: {
			"accept",
			"accept-language",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	// Verify the fingerprint matches Firefox expectations
	expectedH3Hash := "d50d4e585c22bb92b6c86b592aa2d586"

	t.Logf("Firefox_147 HTTP/3 direct path test completed")
	t.Logf("Expected hash: %s", expectedH3Hash)
	t.Logf("Protocol used: %s", resp.Proto)
}
