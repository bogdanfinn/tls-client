package tests

import (
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/bogdanfinn/tls-client/profiles"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	tls "github.com/bogdanfinn/utls"
)

func TestClients(t *testing.T) {
	t.Log("testing chrome 124")
	chrome_124(t)
	t.Log("testing chrome 120")
	chrome_120(t)
	time.Sleep(2 * time.Second)
	t.Log("testing chrome 117")
	chrome_117(t)
	time.Sleep(2 * time.Second)
	t.Log("testing firefox 117")
	firefox_117(t)
	time.Sleep(2 * time.Second)
	t.Log("testing chrome 116 with psk")
	chrome116WithPsk(t)
	t.Log("testing chrome 112")
	chrome112(t)
	time.Sleep(2 * time.Second)
	t.Log("testing chrome 111")
	chrome111(t)
	time.Sleep(2 * time.Second)
	t.Log("testing chrome 110")
	chrome110(t)
	time.Sleep(2 * time.Second)
	t.Log("testing chrome 109")
	chrome109(t)
	time.Sleep(2 * time.Second)
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
	t.Log("testing firefox 108")
	firefox_108(t)
	time.Sleep(2 * time.Second)
	t.Log("testing firefox 110")
	firefox_110(t)
	time.Sleep(2 * time.Second)
	t.Log("testing opera 91")
	opera_91(t)
	t.Log("testing safari ios 17")
	safariIos17(t)
}

func TestCustomClients(t *testing.T) {
	t.Log("testing okhttp4 android 13")
	okhttp4Android13(t)
	time.Sleep(2 * time.Second)
	t.Log("testing okhttp4 android 12")
	okhttp4Android12(t)
	time.Sleep(2 * time.Second)
	t.Log("testing okhttp4 android 11")
	okhttp4Android11(t)
	time.Sleep(2 * time.Second)
	t.Log("testing okhttp4 android 10")
	okhttp4Android10(t)
	time.Sleep(2 * time.Second)
	t.Log("testing okhttp4 android 9")
	okhttp4Android9(t)
	time.Sleep(2 * time.Second)
	t.Log("testing okhttp4 android 8")
	okhttp4Android8(t)
	time.Sleep(2 * time.Second)
	t.Log("testing okhttp4 android 7")
	okhttp4Android7(t)
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

var defaultOkHttp4Header = http.Header{
	"accept-encoding": {"gzip"},
	"user-agent":      {"okhttp/4.10.0"},
	http.HeaderOrderKey: {
		"accept-encoding",
		"user-agent",
	},
}

func chrome116WithPsk(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_116_PSK),
		tls_client.WithTimeoutSeconds(120),
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

	compareResponse(t, "chrome", clientFingerprints[chrome][tls.HelloChrome_112.Str()], resp)

	req, err = http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultHeader

	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, "chrome", clientFingerprints[chrome][tls.HelloChrome_112_PSK.Str()], resp)
}

func chrome112(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_112),
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

	compareResponse(t, "chrome", clientFingerprints[chrome][tls.HelloChrome_112.Str()], resp)
}

func chrome111(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_111),
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

	compareResponse(t, "chrome", clientFingerprints[chrome][tls.HelloChrome_111.Str()], resp)
}

func chrome110(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_110),
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

	compareResponse(t, "chrome", clientFingerprints[chrome][tls.HelloChrome_110.Str()], resp)
}

func chrome109(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_109),
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

	compareResponse(t, "chrome", clientFingerprints[chrome][tls.HelloChrome_109.Str()], resp)
}

func chrome108(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_108),
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

	compareResponse(t, "chrome", clientFingerprints[chrome][tls.HelloChrome_108.Str()], resp)
}

func chrome107(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_107),
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

	compareResponse(t, "chrome", clientFingerprints[chrome][tls.HelloChrome_107.Str()], resp)
}

func chrome105(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_105),
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

	compareResponse(t, "chrome", clientFingerprints[chrome][tls.HelloChrome_105.Str()], resp)
}

func chrome104(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_104),
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

	compareResponse(t, "chrome", clientFingerprints[chrome][tls.HelloChrome_104.Str()], resp)
}

func chrome103(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_103),
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

	compareResponse(t, "chrome", clientFingerprints[chrome][tls.HelloChrome_103.Str()], resp)
}

func safari_16_0(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Safari_16_0),
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

	compareResponse(t, "safari", clientFingerprints[safari][tls.HelloSafari_16_0.Str()], resp)
}

func safari_iOS_16_0(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Safari_IOS_16_0),
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

	compareResponse(t, "safari ios", clientFingerprints[safariIos][tls.HelloIOS_16_0.Str()], resp)
}

func firefox_105(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Firefox_105),
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

	compareResponse(t, "firefox", clientFingerprints[firefox][tls.HelloFirefox_105.Str()], resp)
}

func firefox_106(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Firefox_106),
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

	compareResponse(t, "firefox", clientFingerprints[firefox][tls.HelloFirefox_106.Str()], resp)
}

func firefox_108(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Firefox_108),
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

	compareResponse(t, "firefox", clientFingerprints[firefox][tls.HelloFirefox_108.Str()], resp)
}

func chrome_124(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_124),
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

	compareResponse(t, "chrome", clientFingerprints[chrome][profiles.Chrome_124.GetClientHelloStr()], resp)
}

func chrome_120(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_120),
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

	compareResponse(t, "chrome", clientFingerprints[chrome][profiles.Chrome_120.GetClientHelloStr()], resp)
}

func chrome_117(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_117),
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

	compareResponse(t, "chrome", clientFingerprints[chrome][profiles.Chrome_117.GetClientHelloStr()], resp)
}

func firefox_117(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Firefox_117),
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

	compareResponse(t, "firefox", clientFingerprints[firefox][profiles.Firefox_117.GetClientHelloStr()], resp)
}

func firefox_110(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Firefox_110),
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

	compareResponse(t, "firefox", clientFingerprints[firefox][tls.HelloFirefox_110.Str()], resp)
}

func opera_91(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Opera_91),
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

	compareResponse(t, "opera", clientFingerprints[opera][tls.HelloOpera_91.Str()], resp)
}

func safariIos17(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Safari_IOS_17_0),
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

	compareResponse(t, "safari_IOS", clientFingerprints[safariIos][profiles.Safari_IOS_17_0.GetClientHelloStr()], resp)
}

func compareResponse(t *testing.T, clientName string, expectedValues map[string]string, resp *http.Response) {
	defer resp.Body.Close()

	readBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	tlsApiResponse := TlsApiResponse{}
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

func okhttp4Android13(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Okhttp4Android13),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultOkHttp4Header

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, "okhttp4 android 13", clientFingerprints[okhttpAndroid][profiles.Okhttp4Android13.GetClientHelloStr()], resp)
}

func okhttp4Android12(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Okhttp4Android12),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultOkHttp4Header

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, "okhttp4 android 12", clientFingerprints[okhttpAndroid][profiles.Okhttp4Android12.GetClientHelloStr()], resp)
}

func okhttp4Android11(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Okhttp4Android11),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultOkHttp4Header

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, "okhttp4 android 11", clientFingerprints[okhttpAndroid][profiles.Okhttp4Android11.GetClientHelloStr()], resp)
}

func okhttp4Android10(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Okhttp4Android10),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultOkHttp4Header

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, "okhttp4 android 10", clientFingerprints[okhttpAndroid][profiles.Okhttp4Android10.GetClientHelloStr()], resp)
}

func okhttp4Android9(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Okhttp4Android9),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultOkHttp4Header

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, "okhttp4 android 9", clientFingerprints[okhttpAndroid][profiles.Okhttp4Android9.GetClientHelloStr()], resp)
}

func okhttp4Android8(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Okhttp4Android8),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultOkHttp4Header

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, "okhttp4 android 8", clientFingerprints[okhttpAndroid][profiles.Okhttp4Android8.GetClientHelloStr()], resp)
}

func okhttp4Android7(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Okhttp4Android7),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultOkHttp4Header

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	compareResponse(t, "okhttp4 android 7", clientFingerprints[okhttpAndroid][profiles.Okhttp4Android7.GetClientHelloStr()], resp)
}
