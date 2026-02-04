package tests

import (
	"encoding/json"
	"io"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

type BrowserLeaksResponse struct {
	H3Hash string `json:"h3_hash"`
	H3Text string `json:"h3_text"`
}

func TestHTTP3FingerprintChrome443(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_144),
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithProtocolRacing(),
		tls_client.WithEnableHttp3(),
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
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var browserLeaksResp BrowserLeaksResponse
	if err := json.Unmarshal(body, &browserLeaksResp); err != nil {
		t.Logf("JSON unmarshal error: %v", err)
		t.Fatal(err)
	}

	expectedH3Hash := "ba909fc3dc419ea5c5b26c6323ac1879"
	expectedH3Text := "1:65536;6:262144;7:100;51:1;GREASE|GREASE|984832|m,a,s,p"

	if browserLeaksResp.H3Hash != expectedH3Hash {
		t.Errorf("Chrome_144 HTTP/3 hash mismatch.\nexpected: %s\nactual  : %s", expectedH3Hash, browserLeaksResp.H3Hash)
	}

	if browserLeaksResp.H3Text != expectedH3Text {
		t.Errorf("Chrome_144 HTTP/3 fingerprint mismatch.\nexpected: %s\nactual  : %s", expectedH3Text, browserLeaksResp.H3Text)
	}

	t.Logf("Chrome_144 HTTP/3 fingerprint: %s", browserLeaksResp.H3Text)
	t.Logf("Chrome_144 HTTP/3 hash: %s", browserLeaksResp.H3Hash)
}

func TestHTTP3FingerprintFirefox135(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Firefox_147),
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithProtocolRacing(),
		tls_client.WithEnableHttp3(),
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
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var browserLeaksResp BrowserLeaksResponse
	if err := json.Unmarshal(body, &browserLeaksResp); err != nil {
		t.Logf("JSON unmarshal error: %v", err)
		t.Fatal(err)
	}

	expectedH3Hash := "d50d4e585c22bb92b6c86b592aa2d586"
	expectedH3Text := "1:65536;7:20;727725890:0;16765559:1;51:1;8:1|GREASE|m,s,a,p"

	if browserLeaksResp.H3Hash != expectedH3Hash {
		t.Errorf("Firefox_147 HTTP/3 hash mismatch.\nexpected: %s\nactual  : %s", expectedH3Hash, browserLeaksResp.H3Hash)
	}

	if browserLeaksResp.H3Text != expectedH3Text {
		t.Errorf("Firefox_147 HTTP/3 fingerprint mismatch.\nexpected: %s\nactual  : %s", expectedH3Text, browserLeaksResp.H3Text)
	}

	t.Logf("Firefox_147 HTTP/3 fingerprint: %s", browserLeaksResp.H3Text)
	t.Logf("Firefox_147 HTTP/3 hash: %s", browserLeaksResp.H3Hash)
}

func TestHTTP3FingerprintWithDefaultValuesForChrome(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithEnableHttp3(),
		tls_client.WithProtocolRacing(),
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
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var browserLeaksResp BrowserLeaksResponse
	if err := json.Unmarshal(body, &browserLeaksResp); err != nil {
		t.Logf("JSON unmarshal error: %v", err)
		t.Fatal(err)
	}

	expectedH3Hash := "5d290f560382da8cfd5d89b4f7c8bbe0"
	expectedH3Text := "6:262144;51:1|m,a,s,p"

	if browserLeaksResp.H3Hash != expectedH3Hash {
		t.Errorf("Chrome_133 HTTP/3 hash mismatch.\nexpected: %s\nactual  : %s", expectedH3Hash, browserLeaksResp.H3Hash)
	}

	if browserLeaksResp.H3Text != expectedH3Text {
		t.Errorf("Chrome_133 HTTP/3 fingerprint mismatch.\nexpected: %s\nactual  : %s", expectedH3Text, browserLeaksResp.H3Text)
	}

	t.Logf("Chrome_133 HTTP/3 fingerprint: %s", browserLeaksResp.H3Text)
	t.Logf("Chrome_133 HTTP/3 hash: %s", browserLeaksResp.H3Hash)
}
