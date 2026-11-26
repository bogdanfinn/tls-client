package tls_client_cffi_src

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/httptest"
	"github.com/google/uuid"
)

const bandwidthTestBodySize = 64 * 1024

func newBandwidthTestServers(t *testing.T) map[string]*httptest.Server {
	t.Helper()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, strings.Repeat("b", bandwidthTestBodySize))
	})

	servers := map[string]*httptest.Server{
		"http":  httptest.NewServer(handler),
		"https": httptest.NewTLSServer(handler),
	}
	t.Cleanup(func() {
		for _, srv := range servers {
			srv.Close()
		}
	})

	return servers
}

// doBandwidthRequest runs a request the same way the exported cffi request function does.
func doBandwidthRequest(t *testing.T, input RequestInput) Response {
	t.Helper()

	client, sessionId, withSession, cErr := CreateClient(input)
	if cErr != nil {
		t.Fatalf("CreateClient: %v", cErr)
	}
	if !withSession {
		defer client.Close()
	}

	req, cErr := BuildRequest(input)
	if cErr != nil {
		t.Fatalf("BuildRequest: %v", cErr)
	}

	measurement := StartBandwidthMeasurement(client, input)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}

	out, cErr := BuildResponse(sessionId, withSession, resp, nil, input)
	if cErr != nil {
		t.Fatalf("BuildResponse: %v", cErr)
	}
	out.Bandwidth = measurement.Usage()

	return out
}

func bandwidthInput(url string, sessionId *string, track bool) RequestInput {
	return RequestInput{
		RequestMethod:        http.MethodGet,
		RequestUrl:           url,
		TLSClientIdentifier:  "chrome_133",
		InsecureSkipVerify:   true,
		TimeoutSeconds:       10,
		SessionId:            sessionId,
		WithBandwidthTracker: track,
	}
}

func assertRequestUsage(t *testing.T, usage *BandwidthUsage) {
	t.Helper()

	if usage == nil {
		t.Fatal("expected a bandwidth object in the response")
	}
	if usage.DownloadedBytes < bandwidthTestBodySize {
		t.Fatalf("downloadedBytes = %d, want at least the body size %d", usage.DownloadedBytes, bandwidthTestBodySize)
	}
	if usage.UploadedBytes <= 0 {
		t.Fatalf("uploadedBytes = %d, want > 0", usage.UploadedBytes)
	}
	if usage.TotalBytes != usage.DownloadedBytes+usage.UploadedBytes {
		t.Fatalf("totalBytes = %d, want downloaded+uploaded = %d", usage.TotalBytes, usage.DownloadedBytes+usage.UploadedBytes)
	}
	if usage.SessionTotalBytes != usage.SessionDownloadedBytes+usage.SessionUploadedBytes {
		t.Fatalf("sessionTotalBytes = %d, want sessionDownloaded+sessionUploaded", usage.SessionTotalBytes)
	}
}

func TestBandwidth_NotRequested_OmittedFromResponse(t *testing.T) {
	srv := newBandwidthTestServers(t)["http"]

	out := doBandwidthRequest(t, bandwidthInput(srv.URL, nil, false))
	if out.Bandwidth != nil {
		t.Fatalf("expected no bandwidth object, got %+v", out.Bandwidth)
	}

	raw, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(raw), `"bandwidth"`) {
		t.Fatalf("expected the bandwidth key to be omitted, got %s", raw)
	}
}

func TestBandwidth_WithoutSession(t *testing.T) {
	for scheme, srv := range newBandwidthTestServers(t) {
		t.Run(scheme, func(t *testing.T) {
			usage := doBandwidthRequest(t, bandwidthInput(srv.URL, nil, true)).Bandwidth
			assertRequestUsage(t, usage)

			// A client without a session lives for one request only.
			if usage.SessionDownloadedBytes != usage.DownloadedBytes || usage.SessionUploadedBytes != usage.UploadedBytes {
				t.Fatalf("session counters %+v should equal the request counters without a session", usage)
			}
		})
	}
}

func TestBandwidth_SessionAccumulates(t *testing.T) {
	for scheme, srv := range newBandwidthTestServers(t) {
		t.Run(scheme, func(t *testing.T) {
			sessionId := uuid.New().String()
			t.Cleanup(func() { RemoveSession(sessionId) })

			first := doBandwidthRequest(t, bandwidthInput(srv.URL, &sessionId, true)).Bandwidth
			assertRequestUsage(t, first)

			second := doBandwidthRequest(t, bandwidthInput(srv.URL, &sessionId, true)).Bandwidth
			assertRequestUsage(t, second)

			if second.SessionTotalBytes != first.SessionTotalBytes+second.TotalBytes {
				t.Fatalf("sessionTotalBytes = %d, want first session total %d + second request %d",
					second.SessionTotalBytes, first.SessionTotalBytes, second.TotalBytes)
			}
		})
	}
}

func TestBandwidth_SessionCreatedWithoutTracker_Omitted(t *testing.T) {
	srv := newBandwidthTestServers(t)["http"]

	sessionId := uuid.New().String()
	t.Cleanup(func() { RemoveSession(sessionId) })

	_ = doBandwidthRequest(t, bandwidthInput(srv.URL, &sessionId, false))

	// The tracker is chosen when the session is created; asking for it later must not report zeros.
	out := doBandwidthRequest(t, bandwidthInput(srv.URL, &sessionId, true))
	if out.Bandwidth != nil {
		t.Fatalf("expected no bandwidth object for a session created without the tracker, got %+v", out.Bandwidth)
	}
}
