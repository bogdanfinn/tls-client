package tests

import (
	"io"
	"net/url"
	"testing"

	"github.com/bogdanfinn/tls-client/profiles"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/stretchr/testify/assert"
)

func TestClient_SkipExistingCookiesOnClientSetCookies(t *testing.T) {
	jarOptions := []tls_client.CookieJarOption{
		tls_client.WithSkipExisting(),
	}

	jar := tls_client.NewCookieJar(jarOptions...)

	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithCookieJar(jar),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}
	u := &url.URL{
		Scheme: "http",
		Host:   "testhost.de",
		Path:   "/test",
	}

	assert.Equal(t, 0, len(client.GetCookies(u)))

	cookie := &http.Cookie{
		Name:   "test1",
		Value:  "test1",
		MaxAge: 1,
	}

	client.SetCookies(u, []*http.Cookie{cookie})

	assert.Equal(t, 1, len(client.GetCookies(u)))

	cookie2 := &http.Cookie{
		Name:   "test2",
		Value:  "test2",
		MaxAge: 1,
	}
	client.SetCookies(u, []*http.Cookie{cookie2})

	assert.Equal(t, 2, len(client.GetCookies(u)))

	cookie3 := &http.Cookie{
		Name:   "test1",
		Value:  "test1",
		MaxAge: 1,
	}
	client.SetCookies(u, []*http.Cookie{cookie3})

	assert.Equal(t, 2, len(client.GetCookies(u)))
}

func TestClient_SkipExistingCookiesOnSetCookiesResponse(t *testing.T) {
	jarOptions := []tls_client.CookieJarOption{
		tls_client.WithSkipExisting(),
	}

	jar := tls_client.NewCookieJar(jarOptions...)

	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithCookieJar(jar),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	urlString := "https://eu.kith.com/"
	req, err := http.NewRequest(http.MethodGet, urlString, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = http.Header{
		"accept":          {"*/*"},
		"accept-encoding": {"gzip"},
		"accept-language": {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"user-agent":      {"Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Mobile Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-encoding",
			"accept-language",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	u, _ := url.Parse(urlString)

	cookiesAfterFirstRequest := client.GetCookies(u)

	assert.Equal(t, 7, len(cookiesAfterFirstRequest))

	cookie3 := &http.Cookie{
		Name:   cookiesAfterFirstRequest[0].Name,
		Value:  cookiesAfterFirstRequest[0].Value,
		Domain: cookiesAfterFirstRequest[0].Domain,
		MaxAge: cookiesAfterFirstRequest[0].MaxAge,
	}
	client.SetCookies(u, []*http.Cookie{cookie3})

	assert.Equal(t, 7, len(client.GetCookies(u)))

	req, err = http.NewRequest(http.MethodGet, "https://eu.kith.com/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = http.Header{
		"accept":          {"*/*"},
		"accept-encoding": {"gzip"},
		"accept-language": {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"user-agent":      {"Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Mobile Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-encoding",
			"accept-language",
			"user-agent",
		},
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	cookiesAfterSecondRequest := client.GetCookies(u)

	assert.Equal(t, 7, len(cookiesAfterSecondRequest))
}

func TestClient_ExcludeExpiredCookiesFromRequest(t *testing.T) {
	jar := tls_client.NewCookieJar()

	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithCookieJar(jar),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}
	u := &url.URL{
		Scheme: "http",
		Host:   "testhost.de",
		Path:   "/test",
	}

	assert.Equal(t, 0, len(client.GetCookies(u)))

	cookieAlive := &http.Cookie{
		Name:   "test1",
		Value:  "test1",
		MaxAge: 1,
	}

	cookieExpired := &http.Cookie{
		Name:   "test2",
		Value:  "test2",
		MaxAge: -1,
	}

	client.SetCookies(u, []*http.Cookie{cookieAlive, cookieExpired})

	assert.Equal(t, 1, len(client.GetCookies(u)))

	cookieExpireExisting := &http.Cookie{
		Name:   "test1",
		Value:  "test1",
		MaxAge: -1,
	}
	client.SetCookies(u, []*http.Cookie{cookieExpireExisting})

	assert.Equal(t, 0, len(client.GetCookies(u)))
}
