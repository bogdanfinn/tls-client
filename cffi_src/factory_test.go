package tls_client_cffi_src

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/httptest"
)

// TestBuildResponse_IsByteResponse_NonEmptyBody guards against the regression
// where isByteResponse=true returned an empty base64 payload because
// io.ReadAll was scoped to the !isByteResponse branch.
func TestBuildResponse_IsByteResponse_NonEmptyBody(t *testing.T) {
	want := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0x80, 0x7F}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(want)
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("http.Get: %v", err)
	}

	out, cErr := BuildResponse("", false, resp, nil, RequestInput{IsByteResponse: true})
	if cErr != nil {
		t.Fatalf("BuildResponse: %v", cErr)
	}

	const prefix = "data:"
	if !strings.HasPrefix(out.Body, prefix) {
		t.Fatalf("expected data URI prefix, got %q", out.Body)
	}
	commaIdx := strings.Index(out.Body, ",")
	if commaIdx == -1 {
		t.Fatalf("expected comma separator in data URI, got %q", out.Body)
	}
	b64 := out.Body[commaIdx+1:]
	if b64 == "" {
		t.Fatalf("byte-response payload is empty; body = %q", out.Body)
	}
	got, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		t.Fatalf("base64 decode: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("bytes mismatch: want %x, got %x", want, got)
	}
}

// TestBuildResponse_IsByteResponse_EmptyBody covers the edge case of an empty
// body with isByteResponse=true. The result should be a data URI whose payload
// is an empty base64 string; it should not error.
func TestBuildResponse_IsByteResponse_EmptyBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("http.Get: %v", err)
	}

	out, cErr := BuildResponse("", false, resp, nil, RequestInput{IsByteResponse: true})
	if cErr != nil {
		t.Fatalf("BuildResponse: %v", cErr)
	}

	commaIdx := strings.Index(out.Body, ",")
	if commaIdx == -1 {
		t.Fatalf("expected comma separator in data URI, got %q", out.Body)
	}
	if out.Body[commaIdx+1:] != "" {
		t.Fatalf("expected empty base64 payload, got %q", out.Body[commaIdx+1:])
	}
}

// TestBuildResponse_TextResponse_EmptyBody covers the case #232 originally
// fixed: a zero-length text response should not trip charset.NewReader into
// raising EOF; it should return an empty body cleanly.
func TestBuildResponse_TextResponse_EmptyBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("http.Get: %v", err)
	}

	out, cErr := BuildResponse("", false, resp, nil, RequestInput{IsByteResponse: false})
	if cErr != nil {
		t.Fatalf("BuildResponse: %v", cErr)
	}
	if out.Body != "" {
		t.Fatalf("expected empty body, got %q", out.Body)
	}
}

// TestBuildResponse_TextResponse_NonEmptyBody checks the common text path
// still works: charset detection decodes and returns the body as-is.
func TestBuildResponse_TextResponse_NonEmptyBody(t *testing.T) {
	const want = "hello, world"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(want))
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("http.Get: %v", err)
	}

	out, cErr := BuildResponse("", false, resp, nil, RequestInput{IsByteResponse: false})
	if cErr != nil {
		t.Fatalf("BuildResponse: %v", cErr)
	}
	if out.Body != want {
		t.Fatalf("body mismatch: want %q, got %q", want, out.Body)
	}
}
