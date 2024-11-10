package tests

import (
	"fmt"
	"testing"

	tls_client "github.com/antibotaio/tls-client"
	http "github.com/bogdanfinn/fhttp"
)

var headers = http.Header{
	http.PHeaderOrderKey:        {":method", ":scheme", ":authority", ":path"},
	"cache-control":             {"max-age=0"},
	"sec-ch-ua":                 {`"Chromium";v="130", "Google Chrome";v="130", "Not?A_Brand";v="99"`},
	"sec-ch-ua-mobile":          {"?0"},
	"sec-ch-ua-platform":        {"Windows"},
	"upgrade-insecure-requests": {"1"},
	"user-agent":                {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"},
	"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
	"sec-fetch-site":            {"none"},
	"sec-fetch-mode":            {"navigate"},
	"sec-fetch-user":            {"?1"},
	"sec-fetch-dest":            {"document"},
	"accept-encoding":           {"gzip, deflate, br, zstd"},
	"accept-language":           {"pl-PL,pl;q=0.9"},
	"priority":                  {"u=0, i"},
	http.HeaderOrderKey:         {"cache-control", "sec-ch-ua", "sec-ch-ua-mobile", "sec-ch-ua-platform", "upgrade-insecure-requests", "user-agent", "accept", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-user", "sec-fetch-dest", "accept-encoding", "accept-language", "priority"},
}

func TestAntibotAIO(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithAntibotAIO("x-api-key", []string{"footlocker.pl"}),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewLogger(), options...)
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("GET", "https://footlocker.pl", nil)
	req.Header = headers.Clone()

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(resp.StatusCode)
}
