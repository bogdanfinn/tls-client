package tests

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	stdhttp "net/http"
	"net/http/httptest"
	"os"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/stretchr/testify/require"
)

// Exercise negotiation, certificate verification, response bodies, connection
// reuse, and redialing without depending on an external fingerprint service.
func TestSafari272Requests(t *testing.T) {
	for _, tc := range []struct {
		name       string
		version    uint16
		curve      tls.CurveID
		forceHTTP1 bool
	}{
		{"tls13_h2_mlkem", tls.VersionTLS13, tls.X25519MLKEM768, false},
		{"tls13_h2_x25519", tls.VersionTLS13, tls.X25519, false},
		{"tls12_h2", tls.VersionTLS12, tls.CurveP256, false},
		{"tls13_http1", tls.VersionTLS13, tls.X25519, true},
		{"tls12_http1", tls.VersionTLS12, tls.CurveP256, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			type result struct {
				Protocol   int
				Version    uint16
				RemoteAddr string
			}
			server := httptest.NewUnstartedServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(result{r.ProtoMajor, r.TLS.Version, r.RemoteAddr})
			}))
			server.EnableHTTP2 = true
			server.TLS = &tls.Config{MinVersion: tc.version, MaxVersion: tc.version, CurvePreferences: []tls.CurveID{tc.curve}}
			server.StartTLS()
			defer server.Close()
			roots := x509.NewCertPool()
			roots.AddCert(server.Certificate())
			options := []tls_client.HttpClientOption{
				tls_client.WithClientProfile(profiles.Safari_27_2),
				tls_client.WithDisableHttp3(),
				tls_client.WithTimeoutSeconds(5),
				tls_client.WithTransportOptions(&tls_client.TransportOptions{RootCAs: roots}),
			}
			if tc.forceHTTP1 {
				options = append(options, tls_client.WithForceHttp1())
			}
			client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
			require.NoError(t, err)
			defer client.Close()
			previousAddr := ""
			for i := 0; i < 3; i++ {
				if i == 2 {
					client.CloseIdleConnections()
				}
				req, err := http.NewRequest(http.MethodGet, server.URL, nil)
				require.NoError(t, err)
				resp, err := client.Do(req)
				require.NoError(t, err)
				body, err := io.ReadAll(resp.Body)
				resp.Body.Close()
				require.NoError(t, err)
				require.Equal(t, 200, resp.StatusCode)
				var got result
				require.NoError(t, json.Unmarshal(body, &got))
				protocol := 2
				if tc.forceHTTP1 {
					protocol = 1
				}
				require.Equal(t, protocol, got.Protocol)
				require.Equal(t, protocol, resp.ProtoMajor)
				require.Equal(t, tc.version, got.Version)
				if i == 1 {
					require.Equal(t, previousAddr, got.RemoteAddr, "connection should be reused")
				}
				if i == 2 {
					require.NotEqual(t, previousAddr, got.RemoteAddr, "idle connection should be replaced")
				}
				previousAddr = got.RemoteAddr
			}
		})
	}
}

// Opt in explicitly so ordinary CI does not depend on BrowserLeaks being up.
func TestSafari272LiveFingerprint(t *testing.T) {
	if os.Getenv("TLS_CLIENT_LIVE_SAFARI") != "1" {
		t.Skip("set TLS_CLIENT_LIVE_SAFARI=1 to verify against BrowserLeaks")
	}
	data, err := os.ReadFile("../profiles/testdata/safari_27_2.json")
	require.NoError(t, err)
	var expected map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &expected))
	var userAgent string
	require.NoError(t, json.Unmarshal(expected["user_agent"], &userAgent))
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(),
		tls_client.WithClientProfile(profiles.Safari_27_2),
		tls_client.WithDisableHttp3(), tls_client.WithTimeoutSeconds(20))
	require.NoError(t, err)
	defer client.Close()
	for i := 0; i < 3; i++ {
		t.Run(fmt.Sprintf("connection_%d", i+1), func(t *testing.T) {
			client.CloseIdleConnections()
			req, err := http.NewRequest(http.MethodGet, "https://tls.browserleaks.com/", nil)
			require.NoError(t, err)
			req.Header.Set("user-agent", userAgent)
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, 200, resp.StatusCode)
			require.Equal(t, 2, resp.ProtoMajor)
			var actual map[string]json.RawMessage
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&actual))
			for _, key := range []string{"ja3_text", "ja3_hash", "ja4", "ja4_ro", "akamai_text", "akamai_hash"} {
				require.NotEmpty(t, actual[key], key)
				require.JSONEq(t, string(expected[key]), string(actual[key]), key)
			}
		})
	}
}
