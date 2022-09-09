package tests

import (
	"net/url"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/stretchr/testify/assert"
)

func TestClient_SkipExistingCookies(t *testing.T) {
	client, err := tls_client.ProvideDefaultClient(tls_client.NewNoopLogger())
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
		Name:  "test1",
		Value: "test1",
	}
	client.SetCookies(u, []*http.Cookie{cookie})

	assert.Equal(t, 1, len(client.GetCookies(u)))

	cookie2 := &http.Cookie{
		Name:  "test2",
		Value: "test2",
	}
	client.SetCookies(u, []*http.Cookie{cookie2})
	
	assert.Equal(t, 2, len(client.GetCookies(u)))

	cookie3 := &http.Cookie{
		Name:  "test1",
		Value: "test1",
	}
	client.SetCookies(u, []*http.Cookie{cookie3})

	assert.Equal(t, 2, len(client.GetCookies(u)))
}
