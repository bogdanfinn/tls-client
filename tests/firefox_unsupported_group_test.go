package tests

import (
	"fmt"
	"strings"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/httptest"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	tls "github.com/bogdanfinn/utls"
	"github.com/stretchr/testify/assert"
)

// TestFirefoxHandshakeWithGroupWithoutKeyShare connects with Firefox_110 to a
// server that only accepts P-384. Firefox lists P-384 in supported_groups next
// to the FFDHE groups, but sends key shares for X25519 and P-256 only, so the
// server has to answer with a HelloRetryRequest and the client has to retry
// with a P-384 key share.
func TestFirefoxHandshakeWithGroupWithoutKeyShare(t *testing.T) {
	testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	testServer.TLS = &tls.Config{
		MinVersion:       tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{tls.CurveP384},
	}
	testServer.StartTLS()
	defer testServer.Close()

	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Firefox_110),
		tls_client.WithInsecureSkipVerify(),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/", testServer.URL), strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultHeader

	resp, err := client.Do(req)
	if !assert.NoError(t, err) {
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
