package tests

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	tls "github.com/bogdanfinn/utls"
)

func TestClient_Chrome105(t *testing.T) {
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

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-encoding":           {"gzip"},
		"Accept-Encoding":           {"gzip"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"cache-control":             {"max-age=0"},
		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
		"sec-ch-ua":                 {`" Not A;Brand";v="99", "Chromium";v="101", "Google chrome";v="101"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) chrome/100.0.4896.75 safari/537.36"},
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

	compareResponse(t, browserFingerprints[chrome][tls.HelloChrome_105.Str()], resp)
}

func TestClient_Chrome104(t *testing.T) {
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

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-encoding":           {"gzip"},
		"Accept-Encoding":           {"gzip"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"cache-control":             {"max-age=0"},
		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
		"sec-ch-ua":                 {`" Not A;Brand";v="99", "Chromium";v="101", "Google chrome";v="101"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) chrome/100.0.4896.75 safari/537.36"},
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

	compareResponse(t, browserFingerprints[chrome][tls.HelloChrome_104.Str()], resp)
}

func TestClient_Chrome103(t *testing.T) {
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

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-encoding":           {"gzip"},
		"Accept-Encoding":           {"gzip"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"cache-control":             {"max-age=0"},
		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
		"sec-ch-ua":                 {`" Not A;Brand";v="99", "Chromium";v="101", "Google chrome";v="101"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) chrome/100.0.4896.75 safari/537.36"},
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

	compareResponse(t, browserFingerprints[chrome][tls.HelloChrome_103.Str()], resp)
}

func TestClient_Safari_15_3(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Safari_15_3),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
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
		"sec-ch-ua":                 {`" Not A;Brand";v="99", "Chromium";v="101", "Google chrome";v="101"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) chrome/100.0.4896.75 safari/537.36"},
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

	compareResponse(t, browserFingerprints[safari][tls.HelloSafari_15_3.Str()], resp)
}

func TestClient_Safari_15_6_1(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Safari_15_6_1),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
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
		"sec-ch-ua":                 {`" Not A;Brand";v="99", "Chromium";v="101", "Google chrome";v="101"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) chrome/100.0.4896.75 safari/537.36"},
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

	compareResponse(t, browserFingerprints[safari][tls.HelloSafari_15_6_1.Str()], resp)
}

func TestClient_Safari_16_0(t *testing.T) {
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

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-encoding":           {"gzip"},
		"Accept-Encoding":           {"gzip"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"cache-control":             {"max-age=0"},
		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
		"sec-ch-ua":                 {`" Not A;Brand";v="99", "Chromium";v="101", "Google chrome";v="101"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) chrome/100.0.4896.75 safari/537.36"},
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

	compareResponse(t, browserFingerprints[safari][tls.HelloSafari_16_0.Str()], resp)
}

func TestClient_Safari_Ipad_15_6(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Safari_Ipad_15_6),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
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
		"sec-ch-ua":                 {`" Not A;Brand";v="99", "Chromium";v="101", "Google chrome";v="101"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) chrome/100.0.4896.75 safari/537.36"},
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

	compareResponse(t, browserFingerprints[safariIpadOs][tls.HelloIPad_15_6.Str()], resp)
}

func TestClient_Safari_iOS_15_5(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Safari_IOS_15_5),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
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
		"sec-ch-ua":                 {`" Not A;Brand";v="99", "Chromium";v="101", "Google chrome";v="101"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) chrome/100.0.4896.75 safari/537.36"},
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

	compareResponse(t, browserFingerprints[safariIos][tls.HelloIOS_15_5.Str()], resp)
}

func TestClient_Safari_iOS_16_0(t *testing.T) {
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

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-encoding":           {"gzip"},
		"Accept-Encoding":           {"gzip"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"cache-control":             {"max-age=0"},
		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
		"sec-ch-ua":                 {`" Not A;Brand";v="99", "Chromium";v="101", "Google chrome";v="101"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) chrome/100.0.4896.75 safari/537.36"},
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

	compareResponse(t, browserFingerprints[safariIos][tls.HelloIOS_16_0.Str()], resp)
}

func TestClient_Firefox_102(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Firefox_102),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
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
		"sec-ch-ua":                 {`" Not A;Brand";v="99", "Chromium";v="101", "Google chrome";v="101"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) chrome/100.0.4896.75 safari/537.36"},
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

	compareResponse(t, browserFingerprints[firefox][tls.HelloFirefox_102.Str()], resp)
}

func TestClient_Opera_89(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tls_client.Opera_89),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
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
		"sec-ch-ua":                 {`" Not A;Brand";v="99", "Chromium";v="101", "Google chrome";v="101"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) chrome/100.0.4896.75 safari/537.36"},
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

	compareResponse(t, browserFingerprints[opera][tls.HelloOpera_89.Str()], resp)
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

	if len(expectedValues) != len(browserFingerprints[chrome][tls.HelloChrome_105.Str()]) {
		t.Fatal("not enough expected values specified for client")
	}

	for key, expectedValue := range expectedValues {
		switch key {
		case ja3String:
			if tlsApiResponse.TLS.Ja3 != expectedValue {
				t.Errorf("TLS Ja3 mismatch.\nexpected: %s\nactual  : %s", expectedValue, tlsApiResponse.TLS.Ja3)
			}
		case ja3Hash:
			if tlsApiResponse.TLS.Ja3Hash != expectedValue {
				t.Errorf("TLS Ja3 hash mismatch.\nexpected: %s\nactual  : %s", expectedValue, tlsApiResponse.TLS.Ja3Hash)
			}
		case ja3StringWithPadding:
			if tlsApiResponse.TLS.Ja3Padding != expectedValue {
				t.Errorf("TLS Ja3 (wtith Padding) mismatch.\nexpected: %s\nactual  : %s", expectedValue, tlsApiResponse.TLS.Ja3)
			}
		case ja3HashWithPadding:
			if tlsApiResponse.TLS.Ja3HashPadding != expectedValue {
				t.Errorf("TLS Ja3 hash (wtith Padding) mismatch.\nexpected: %s\nactual  : %s", expectedValue, tlsApiResponse.TLS.Ja3Hash)
			}
		case akamaiFingerprint:
			if tlsApiResponse.HTTP2.AkamaiFingerprint != expectedValue {
				t.Errorf("akamai fingerprint mismatch.\nexpected: %s\nactual  : %s", expectedValue, tlsApiResponse.HTTP2.AkamaiFingerprint)
			}
		case akamaiFingerprintHash:
			if tlsApiResponse.HTTP2.AkamaiFingerprintHash != expectedValue {
				t.Errorf("akamai fingerprint hash mismatch.\nexpected: %s\nactual  : %s", expectedValue, tlsApiResponse.HTTP2.AkamaiFingerprintHash)
			}
		}
	}
}
