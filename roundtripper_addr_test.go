package tls_client

import (
	"testing"

	http "github.com/bogdanfinn/fhttp"
)

func TestGetDialTLSAddr(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://example.com/", "example.com:443"},
		{"http://example.com/", "example.com:80"},
		{"http://example.com:8080/", "example.com:8080"},
		{"https://[::1]/", "[::1]:443"},
	}

	rt := &roundTripper{}
	for _, tt := range tests {
		req, err := http.NewRequest(http.MethodGet, tt.url, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got := rt.getDialTLSAddr(req); got != tt.want {
			t.Errorf("getDialTLSAddr(%s) = %s, want %s", tt.url, got, tt.want)
		}
	}
}
