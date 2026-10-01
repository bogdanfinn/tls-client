package tls_client

import (
	"testing"

	http "github.com/bogdanfinn/fhttp"
)

// A plain-http URL and an https URL for the same host must get different
// transport cache keys. With one key, a redirect chain that passes through
// http:// (https://a -> http://b -> https://b) cached an HTTP/1 transport for
// b:443 and the final https request failed with "malformed HTTP response".
func TestGetDialTLSAddrDefaultsPortByScheme(t *testing.T) {
	rt := &roundTripper{}
	cases := map[string]string{
		"https://example.com/":      "example.com:443",
		"http://example.com/":       "example.com:80",
		"HTTP://example.com/":       "example.com:80",
		"http://example.com:8080/":  "example.com:8080",
		"https://example.com:8443/": "example.com:8443",
		"http://[::1]/":             "[::1]:80",
	}
	for rawURL, want := range cases {
		req, err := http.NewRequest(http.MethodGet, rawURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got := rt.getDialTLSAddr(req); got != want {
			t.Errorf("getDialTLSAddr(%q) = %q, want %q", rawURL, got, want)
		}
	}
}
