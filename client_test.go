package tls_client_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	http "github.com/Carcraftz/fhttp"
	tls "github.com/Carcraftz/utls"
	tls_client "github.com/bogdanfinn/tls-client"
)

const apiEndpoint = "https://tls.peet.ws/api/all"

type TlsApiResponse struct {
	Donate      string `json:"donate"`
	IP          string `json:"ip"`
	HTTPVersion string `json:"http_version"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	TLS         struct {
		Ciphers    []string `json:"ciphers"`
		Curves     []string `json:"curves"`
		Extensions []string `json:"extensions"`
		Points     []string `json:"points"`
		Version    string   `json:"version"`
		Protocols  []string `json:"protocols"`
		Versions   []string `json:"versions"`
		Ja3        string   `json:"ja3"`
		Ja3Hash    string   `json:"ja3_hash"`
	} `json:"tls"`
	HTTP2 struct {
		AkamaiFingerprint     string `json:"akamai_fingerprint"`
		AkamaiFingerprintHash string `json:"akamai_fingerprint_hash"`
		SentFrames            []struct {
			FrameType string   `json:"frame_type"`
			Length    int      `json:"length"`
			Settings  []string `json:"settings,omitempty"`
			Increment int      `json:"increment,omitempty"`
			StreamID  int      `json:"stream_id,omitempty"`
			Headers   []string `json:"headers,omitempty"`
		} `json:"sent_frames"`
	} `json:"http2"`
}

func TestClient_Chrome103(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Chrome_103),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, apiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-encoding":           {"gzip"},
		"Accept-Encoding":           {"gzip"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"cache-control":             {"max-age=0"},
		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
		"sec-ch-ua":                 {`" Not A;Brand";v="99", "Chromium";v="101", "Google Chrome";v="101"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.75 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-encoding",
			"accept-language",
			"cache-control",
			"if-none-match",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"upgrade-insecure-requests",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, tls_client.BrowserFingerprints[tls_client.Chrome][tls.HelloChrome_103], resp)
}

func TestClient_Safari_15_3(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Safari_15_3),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, apiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-encoding":           {"gzip"},
		"Accept-Encoding":           {"gzip"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"cache-control":             {"max-age=0"},
		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
		"sec-ch-ua":                 {`" Not A;Brand";v="99", "Chromium";v="101", "Google Chrome";v="101"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.75 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-encoding",
			"accept-language",
			"cache-control",
			"if-none-match",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"upgrade-insecure-requests",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, tls_client.BrowserFingerprints[tls_client.Safari][tls.HelloSafari_15_3], resp)
}

func TestClient_Safari_15_5(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Safari_15_5),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, apiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-encoding":           {"gzip"},
		"Accept-Encoding":           {"gzip"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"cache-control":             {"max-age=0"},
		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
		"sec-ch-ua":                 {`" Not A;Brand";v="99", "Chromium";v="101", "Google Chrome";v="101"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.75 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-encoding",
			"accept-language",
			"cache-control",
			"if-none-match",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"upgrade-insecure-requests",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, tls_client.BrowserFingerprints[tls_client.Safari][tls.HelloSafari_15_5], resp)
}

func TestClient_Safari_iOS_15_5(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Safari_IOS_15_5),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, apiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-encoding":           {"gzip"},
		"Accept-Encoding":           {"gzip"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"cache-control":             {"max-age=0"},
		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
		"sec-ch-ua":                 {`" Not A;Brand";v="99", "Chromium";v="101", "Google Chrome";v="101"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.75 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-encoding",
			"accept-language",
			"cache-control",
			"if-none-match",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"upgrade-insecure-requests",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, tls_client.BrowserFingerprints[tls_client.SafariIOS][tls.HelloIOS_15_5], resp)
}

func TestClient_Firefox_102(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Firefox_102),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, apiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-encoding":           {"gzip"},
		"Accept-Encoding":           {"gzip"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"cache-control":             {"max-age=0"},
		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
		"sec-ch-ua":                 {`" Not A;Brand";v="99", "Chromium";v="101", "Google Chrome";v="101"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.75 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-encoding",
			"accept-language",
			"cache-control",
			"if-none-match",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"upgrade-insecure-requests",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, tls_client.BrowserFingerprints[tls_client.Firefox][tls.HelloFirefox_102], resp)
}

func TestClient_Opera_89(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Opera_89),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, apiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-encoding":           {"gzip"},
		"Accept-Encoding":           {"gzip"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"cache-control":             {"max-age=0"},
		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
		"sec-ch-ua":                 {`" Not A;Brand";v="99", "Chromium";v="101", "Google Chrome";v="101"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.75 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-encoding",
			"accept-language",
			"cache-control",
			"if-none-match",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"upgrade-insecure-requests",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, tls_client.BrowserFingerprints[tls_client.Opera][tls.HelloOpera_89], resp)
}

func compareResponse(t *testing.T, expectedValues map[string]string, resp *http.Response) {
	defer resp.Body.Close()

	readBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	tlsApiResponse := TlsApiResponse{}
	if err := json.Unmarshal(readBytes, &tlsApiResponse); err != nil {
		t.Fatal(err)
	}

	for key, expectedValue := range expectedValues {
		switch key {
		case tls_client.Ja3String:
			if tlsApiResponse.TLS.Ja3 != expectedValue {
				t.Errorf("TLS Ja3 mismatch.\nexpected: %s\nactual  : %s", expectedValue, tlsApiResponse.TLS.Ja3)
			}
		case tls_client.Ja3Hash:
			if tlsApiResponse.TLS.Ja3Hash != expectedValue {
				t.Errorf("TLS Ja3 hash mismatch.\nexpected: %s\nactual  : %s", expectedValue, tlsApiResponse.TLS.Ja3Hash)
			}
		case tls_client.AkamaiFingerprint:
			if tlsApiResponse.HTTP2.AkamaiFingerprint != expectedValue {
				t.Errorf("akamai fingerprint mismatch.\nexpected: %s\nactual  : %s", expectedValue, tlsApiResponse.HTTP2.AkamaiFingerprint)
			}
		case tls_client.AkamaiFingerprintHash:
			if tlsApiResponse.HTTP2.AkamaiFingerprintHash != expectedValue {
				t.Errorf("akamai fingerprint hash mismatch.\nexpected: %s\nactual  : %s", expectedValue, tlsApiResponse.HTTP2.AkamaiFingerprintHash)
			}
		}
	}
}
