package tls_client

import (
	"io"
	"strings"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/httptest"
)

// Plain http:// goes through roundTripper.dial, which used to skip the tracker.
func TestBandwidthTracker_CountsPlainHttp(t *testing.T) {
	body := strings.Repeat("a", 32*1024)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()

	client, err := NewHttpClient(NewNoopLogger(), WithBandwidthTracker(), WithTimeoutSeconds(10))
	if err != nil {
		t.Fatalf("NewHttpClient: %v", err)
	}
	defer client.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		t.Fatalf("read body: %v", err)
	}
	_ = resp.Body.Close()

	tracker := client.GetBandwidthTracker()
	if got := tracker.GetReadBytes(); got < int64(len(body)) {
		t.Fatalf("read bytes = %d, want at least the body size %d", got, len(body))
	}
	if got := tracker.GetWriteBytes(); got <= 0 {
		t.Fatalf("write bytes = %d, want > 0 (the request line and headers)", got)
	}
}
