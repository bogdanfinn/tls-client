package tests

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"strings"
	"testing"

	tls_client_cffi_src "github.com/bogdanfinn/tls-client/cffi_src"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests pin down BuildResponse's byte-response branch. Regression
// motivation: a refactor previously moved io.ReadAll inside the !isByteResponse
// arm only, so every IsByteResponse=true request silently returned an empty
// payload (just the "data:<mime>;base64," prefix with no body). The branches
// below cover the read-decode-encode path that the bug bypassed.

// buildByteResponseInput returns a minimal RequestInput with IsByteResponse=true.
// All other knobs are left at zero values — BuildResponse only reads
// IsByteResponse and StreamOutputPath off RequestInput.
func buildByteResponseInput() tls_client_cffi_src.RequestInput {
	return tls_client_cffi_src.RequestInput{
		IsByteResponse: true,
	}
}

// decodeDataURL splits a "data:<mime>;base64,<payload>" body into mime + raw
// bytes. Fails the test if the envelope is malformed.
func decodeDataURL(t *testing.T, body string) (mime string, payload []byte) {
	t.Helper()
	require.True(t, strings.HasPrefix(body, "data:"), "byte response body must start with data: prefix; got %q", body)
	prefix, encoded, found := strings.Cut(body, ";base64,")
	require.True(t, found, "byte response body must contain ;base64, separator; got %q", body)
	mime = strings.TrimPrefix(prefix, "data:")
	got, err := base64.StdEncoding.DecodeString(encoded)
	require.NoError(t, err, "base64 decode of body payload")
	return mime, got
}

func TestBuildResponse_ByteResponse_BinaryPayloadReturnedAsBase64DataUrl(t *testing.T) {
	// Bytes that don't survive a UTF-8 round trip — proves we're not silently
	// running through the charset/text path.
	payload := []byte{0x00, 0x01, 0xFF, 0xFE, 0x80, 0x7F, 0x42, 0x00, 0xC3, 0x28}

	resp := streamMakeResp(io.NopCloser(bytes.NewReader(payload)), "application/octet-stream", "", true)

	out, err := tls_client_cffi_src.BuildResponse("", false, resp, nil, buildByteResponseInput())
	require.Nil(t, err, "BuildResponse returned error: %v", err)

	assert.Equal(t, 200, out.Status)
	mime, decoded := decodeDataURL(t, out.Body)
	assert.NotEmpty(t, mime, "mime type should be detected and present in data URL")
	assert.Equal(t, payload, decoded, "decoded payload must match the original bytes")
}

func TestBuildResponse_ByteResponse_EmptyBody(t *testing.T) {
	// Empty bodies must still produce a well-formed envelope: the bug regressed
	// every byte response into looking like an empty body, which made this
	// case indistinguishable from "real content lost". Pin the contract:
	// empty input → valid data URL with zero-length payload.
	resp := streamMakeResp(io.NopCloser(strings.NewReader("")), "application/octet-stream", "", true)

	out, err := tls_client_cffi_src.BuildResponse("", false, resp, nil, buildByteResponseInput())
	require.Nil(t, err, "BuildResponse returned error: %v", err)

	mime, decoded := decodeDataURL(t, out.Body)
	assert.NotEmpty(t, mime, "mime type should still be set even for empty body")
	assert.Empty(t, decoded, "decoded payload should be empty for an empty body")
}

func TestBuildResponse_ByteResponse_LargerPayloadFullyRead(t *testing.T) {
	// Larger-than-default-buffer payload with non-text bytes. Catches a
	// regression where a partial-read bug would truncate the body without
	// surfacing an error.
	const size = 64 * 1024
	payload := make([]byte, size)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	resp := streamMakeResp(io.NopCloser(bytes.NewReader(payload)), "application/octet-stream", "", true)

	out, err := tls_client_cffi_src.BuildResponse("", false, resp, nil, buildByteResponseInput())
	require.Nil(t, err, "BuildResponse returned error: %v", err)

	_, decoded := decodeDataURL(t, out.Body)
	assert.Equal(t, len(payload), len(decoded), "decoded length mismatch")
	assert.Equal(t, payload, decoded, "decoded payload must match input byte-for-byte")
}

func TestBuildResponse_ByteResponse_GzipDecompression(t *testing.T) {
	// Byte responses still flow through DecompressBodyByType — verify gzip is
	// transparently expanded before base64 encoding the *decompressed* bytes.
	original := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55}

	var compressed bytes.Buffer
	gz := gzip.NewWriter(&compressed)
	_, err := gz.Write(original)
	require.NoError(t, err, "gzip write")
	require.NoError(t, gz.Close(), "gzip close")

	// uncompressed=false signals BuildResponse to apply DecompressBodyByType.
	resp := streamMakeResp(io.NopCloser(bytes.NewReader(compressed.Bytes())), "application/octet-stream", "gzip", false)

	out, buildErr := tls_client_cffi_src.BuildResponse("", false, resp, nil, buildByteResponseInput())
	require.Nil(t, buildErr, "BuildResponse returned error: %v", buildErr)

	_, decoded := decodeDataURL(t, out.Body)
	assert.Equal(t, original, decoded, "byte response body must be the gzip-decompressed bytes, not the gzipped bytes")
}

func TestBuildResponse_TextResponseStillReadsFullBody(t *testing.T) {
	// Sanity check on the !isByteResponse branch — make sure the byte-branch
	// fix didn't accidentally regress the original text path.
	const payload = `{"hello":"world"}`
	resp := streamMakeResp(io.NopCloser(strings.NewReader(payload)), "application/json; charset=utf-8", "", true)

	input := tls_client_cffi_src.RequestInput{IsByteResponse: false}

	out, err := tls_client_cffi_src.BuildResponse("", false, resp, nil, input)
	require.Nil(t, err, "BuildResponse returned error: %v", err)

	assert.Equal(t, payload, out.Body)
	assert.Equal(t, 200, out.Status)
}

func TestBuildResponse_ByteResponse_PreservesHeadersAndSession(t *testing.T) {
	// Headers and session metadata shouldn't be lost on the byte path either.
	payload := []byte{0x01, 0x02, 0x03}
	resp := streamMakeResp(io.NopCloser(bytes.NewReader(payload)), "application/pdf", "", true)
	resp.Header.Set("X-Custom", "kept")

	out, err := tls_client_cffi_src.BuildResponse("session-x", true, resp, nil, buildByteResponseInput())
	require.Nil(t, err)

	assert.Equal(t, []string{"kept"}, out.Headers["X-Custom"], "custom header should survive the byte path")
	assert.Equal(t, "session-x", out.SessionId, "session id should be propagated when withSession=true")
	assert.NotEmpty(t, out.Target, "target should reflect the request URL")
	assert.NotEmpty(t, out.Id, "response should have an id (used for freeMemory tracking)")

	_, decoded := decodeDataURL(t, out.Body)
	assert.Equal(t, payload, decoded)
}
