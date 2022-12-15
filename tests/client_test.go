package tests

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/shared"
	tls "github.com/bogdanfinn/utls"
)

func TestClients(t *testing.T) {
	t.Log("testing chrome 108")
	chrome108(t)
	time.Sleep(2 * time.Second)
	t.Log("testing chrome 107")
	chrome107(t)
	time.Sleep(2 * time.Second)
	t.Log("testing chrome 105")
	chrome105(t)
	time.Sleep(2 * time.Second)
	t.Log("testing chrome 104")
	chrome104(t)
	time.Sleep(2 * time.Second)
	t.Log("testing chrome 103")
	chrome103(t)
	time.Sleep(2 * time.Second)
	t.Log("testing safari 16")
	safari_16_0(t)
	time.Sleep(2 * time.Second)
	t.Log("testing safari ios 16")
	safari_iOS_16_0(t)
	time.Sleep(2 * time.Second)
	t.Log("testing firefox 105")
	firefox_105(t)
	time.Sleep(2 * time.Second)
	t.Log("testing firefox 106")
	firefox_106(t)
	time.Sleep(2 * time.Second)
	t.Log("testing opera 91")
	opera_91(t)
}

var defaultHeader = http.Header{
	"accept":          {"*/*"},
	"accept-encoding": {"gzip"},
	"accept-language": {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
	"user-agent":      {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) chrome/100.0.4896.75 safari/537.36"},
	http.HeaderOrderKey: {
		"accept",
		"accept-encoding",
		"accept-language",
		"user-agent",
	},
}

func chrome108(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Chrome_108),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultHeader

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, "chrome", browserFingerprints[chrome][tls.HelloChrome_108.Str()], resp)
}

func chrome107(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Chrome_107),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultHeader

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, "chrome", browserFingerprints[chrome][tls.HelloChrome_107.Str()], resp)
}

func chrome105(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Chrome_105),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultHeader

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, "chrome", browserFingerprints[chrome][tls.HelloChrome_105.Str()], resp)
}

func chrome104(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Chrome_104),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultHeader

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, "chrome", browserFingerprints[chrome][tls.HelloChrome_104.Str()], resp)
}

func chrome103(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Chrome_103),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultHeader

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, "chrome", browserFingerprints[chrome][tls.HelloChrome_103.Str()], resp)
}

func safari_16_0(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Safari_16_0),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultHeader

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, "safari", browserFingerprints[safari][tls.HelloSafari_16_0.Str()], resp)
}

func safari_iOS_16_0(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Safari_IOS_16_0),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultHeader

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, "safari ios", browserFingerprints[safariIos][tls.HelloIOS_16_0.Str()], resp)
}

func firefox_105(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Firefox_105),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultHeader

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, "firefox", browserFingerprints[firefox][tls.HelloFirefox_105.Str()], resp)
}

func firefox_106(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Firefox_106),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultHeader

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, "firefox", browserFingerprints[firefox][tls.HelloFirefox_106.Str()], resp)
}

func opera_91(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Opera_91),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultHeader

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, "opera", browserFingerprints[opera][tls.HelloOpera_91.Str()], resp)
}

func compareResponse(t *testing.T, clientName string, expectedValues map[string]string, resp *http.Response) {
	defer resp.Body.Close()

	readBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	tlsApiResponse := shared.TlsApiResponse{}
	if err := json.Unmarshal(readBytes, &tlsApiResponse); err != nil {
		t.Fatal(err)
	}

	for key, expectedValue := range expectedValues {
		switch key {
		case ja3String:
			if tlsApiResponse.TLS.Ja3 != expectedValue {
				t.Errorf("TLS Ja3 mismatch.\nexpected: %s\nactual  : %s\nclient: %s", expectedValue, tlsApiResponse.TLS.Ja3, clientName)
			}
		case ja3Hash:
			if tlsApiResponse.TLS.Ja3Hash != expectedValue {
				t.Errorf("TLS Ja3 hash mismatch.\nexpected: %s\nactual  : %s\nclient: %s", expectedValue, tlsApiResponse.TLS.Ja3Hash, clientName)
			}
		case akamaiFingerprint:
			if tlsApiResponse.HTTP2.AkamaiFingerprint != expectedValue {
				t.Errorf("akamai fingerprint mismatch.\nexpected: %s\nactual  : %s\nclient: %s", expectedValue, tlsApiResponse.HTTP2.AkamaiFingerprint, clientName)
			}
		case akamaiFingerprintHash:
			if tlsApiResponse.HTTP2.AkamaiFingerprintHash != expectedValue {
				t.Errorf("akamai fingerprint hash mismatch.\nexpected: %s\nactual  : %s\nclient: %s", expectedValue, tlsApiResponse.HTTP2.AkamaiFingerprintHash, clientName)
			}
		}
	}
}
