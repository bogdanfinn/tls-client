package tests

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tls_client_cffi_src "github.com/bogdanfinn/tls-client/cffi_src"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests pin down BuildResponse's StreamOutputPath path. Regression
// motivation: the body was being routed through charset.NewReader before the
// stream-to-file write, so binary payloads (images, archives, anything with
// high-byte content) got silently re-encoded — typically windows-1252 → UTF-8,
// inflating the file by ~1.6×. The fix skips charset detection when the
// caller has asked for raw-bytes-to-file behaviour.

// streamOutputInput returns a minimal RequestInput that asks BuildResponse to
// write the body to streamPath. blockSize is forwarded to the file writer
// (1 KiB if zero).
func streamOutputInput(streamPath string, blockSize int) tls_client_cffi_src.RequestInput {
	in := tls_client_cffi_src.RequestInput{
		IsByteResponse:   false,
		StreamOutputPath: &streamPath,
	}
	if blockSize > 0 {
		in.StreamOutputBlockSize = &blockSize
	}
	return in
}

// fakePNG returns a minimal PNG-shaped byte sequence with a generous mix of
// high-byte values (0x80-0xFF). The exact bytes don't have to be a valid PNG
// — the test only cares that running them through a UTF-8 charset transcoder
// would corrupt them, which is true for any sequence with bytes ≥ 0x80
// outside the UTF-8 grammar.
func fakePNG(t *testing.T, size int) []byte {
	t.Helper()
	require.GreaterOrEqual(t, size, 16, "fakePNG: size must include at least the header")
	out := make([]byte, size)
	// PNG signature + a couple of high-byte values that are NOT valid UTF-8
	// continuation sequences.
	header := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
	copy(out, header)
	for i := len(header); i < size; i++ {
		// Cycle through 0x80-0xFF — the high-byte range that windows-1252
		// would "decode" into multi-byte UTF-8 sequences.
		out[i] = byte(0x80 + (i % 0x80))
	}
	return out
}

func TestBuildResponse_StreamOutputPath_PreservesBinaryBytes(t *testing.T) {
	// Main regression test: a PNG-shaped binary body with Content-Type:
	// image/png must be written to disk byte-for-byte, NOT routed through
	// charset.NewReader.
	const size = 8192
	payload := fakePNG(t, size)

	streamPath := filepath.Join(t.TempDir(), "image.png")
	resp := streamMakeResp(io.NopCloser(bytes.NewReader(payload)), "image/png", "", true)

	out, err := tls_client_cffi_src.BuildResponse("", false, resp, nil, streamOutputInput(streamPath, 1024))
	require.Nil(t, err, "BuildResponse returned error: %v", err)
	assert.Equal(t, 200, out.Status)

	written, readErr := os.ReadFile(streamPath)
	require.NoError(t, readErr, "reading written file")
	assert.Equal(t, len(payload), len(written), "file length must match input length (would be ~1.6× longer if charset transcoder was applied)")
	assert.Equal(t, payload, written, "file bytes must match input byte-for-byte")
}

func TestBuildResponse_StreamOutputPath_LargerBinaryFullySaved(t *testing.T) {
	// Larger payload that spans many block-write iterations. The previous
	// charset corruption still showed up here, but a buffered-write bug
	// could also surface here as truncation; the test pins both.
	const size = 256 * 1024 // 256 KiB
	payload := fakePNG(t, size)

	streamPath := filepath.Join(t.TempDir(), "big.bin")
	resp := streamMakeResp(io.NopCloser(bytes.NewReader(payload)), "image/png", "", true)

	out, err := tls_client_cffi_src.BuildResponse("", false, resp, nil, streamOutputInput(streamPath, 4096))
	require.Nil(t, err, "BuildResponse returned error: %v", err)
	assert.Equal(t, 200, out.Status)

	written, readErr := os.ReadFile(streamPath)
	require.NoError(t, readErr)
	assert.Equal(t, len(payload), len(written), "file length must match input length")
	assert.Equal(t, payload, written)
}

func TestBuildResponse_StreamOutputPath_OctetStreamPreserved(t *testing.T) {
	// application/octet-stream is the classic "I have no idea what this is,
	// don't touch it" content type — make sure we still don't touch it.
	payload := []byte{0xFF, 0xFE, 0x00, 0x01, 0x80, 0x90, 0xA0, 0xB0, 0xC3, 0x28}

	streamPath := filepath.Join(t.TempDir(), "blob.bin")
	resp := streamMakeResp(io.NopCloser(bytes.NewReader(payload)), "application/octet-stream", "", true)

	_, err := tls_client_cffi_src.BuildResponse("", false, resp, nil, streamOutputInput(streamPath, 1024))
	require.Nil(t, err)

	written, readErr := os.ReadFile(streamPath)
	require.NoError(t, readErr)
	assert.Equal(t, payload, written)
}

func TestBuildResponse_StreamOutputPath_NoContentTypePreserved(t *testing.T) {
	// Empty Content-Type is the worst case for the buggy code path: charset
	// detection sniffs the body itself and may pick something that breaks
	// binary content. Verify the file write skips charset detection here too.
	payload := []byte{0xCA, 0xFE, 0xBA, 0xBE, 0x00, 0xFF, 0xEE, 0xDD}

	streamPath := filepath.Join(t.TempDir(), "noct.bin")
	resp := streamMakeResp(io.NopCloser(bytes.NewReader(payload)), "", "", true)

	_, err := tls_client_cffi_src.BuildResponse("", false, resp, nil, streamOutputInput(streamPath, 1024))
	require.Nil(t, err)

	written, readErr := os.ReadFile(streamPath)
	require.NoError(t, readErr)
	assert.Equal(t, payload, written)
}

func TestBuildResponse_StreamOutputPath_TextResponseStillReachesFile(t *testing.T) {
	// Sanity check: when the body genuinely is text, the stream-to-file path
	// still works (we want raw bytes — the caller chose stream-to-file
	// precisely because they don't want the charset wrapper). Pin the new
	// behaviour so it's clear: stream-to-file always writes raw bytes,
	// regardless of Content-Type.
	const payload = `{"hello":"world","unicode":"ünïçødé"}`

	streamPath := filepath.Join(t.TempDir(), "text.json")
	resp := streamMakeResp(io.NopCloser(strings.NewReader(payload)), "application/json; charset=utf-8", "", true)

	_, err := tls_client_cffi_src.BuildResponse("", false, resp, nil, streamOutputInput(streamPath, 1024))
	require.Nil(t, err)

	written, readErr := os.ReadFile(streamPath)
	require.NoError(t, readErr)
	assert.Equal(t, []byte(payload), written, "raw UTF-8 bytes should land on disk unchanged")
}

func TestBuildResponse_StreamOutputPath_EmptyBodyProducesEmptyFile(t *testing.T) {
	// Empty body must NOT create or write to the file (the !isByteResponse
	// branch short-circuits via n == 0). Pin the contract so we know exactly
	// what to expect on disk.
	streamPath := filepath.Join(t.TempDir(), "empty.bin")
	resp := streamMakeResp(io.NopCloser(strings.NewReader("")), "image/png", "", true)

	_, err := tls_client_cffi_src.BuildResponse("", false, resp, nil, streamOutputInput(streamPath, 1024))
	require.Nil(t, err)

	// Either the file was not created (empty body short-circuits) or it
	// exists with zero length. Both behaviours are acceptable; what's NOT
	// acceptable is a file containing transcoded garbage.
	info, statErr := os.Stat(streamPath)
	if statErr == nil {
		assert.Equal(t, int64(0), info.Size(), "empty body must not produce non-empty file")
	} else {
		assert.True(t, os.IsNotExist(statErr), "unexpected stat error: %v", statErr)
	}
}

func TestBuildResponse_NoStreamOutputPath_TextResponseStillUsesCharsetReader(t *testing.T) {
	// Counterpart sanity: when StreamOutputPath is NOT set, the charset
	// path still applies. We pin this so the fix doesn't accidentally
	// disable charset detection for the in-memory text path. The test is
	// indirect — we just confirm a UTF-8 JSON body comes back as the
	// expected string.
	const payload = `{"hello":"world"}`
	resp := streamMakeResp(io.NopCloser(strings.NewReader(payload)), "application/json; charset=utf-8", "", true)

	input := tls_client_cffi_src.RequestInput{IsByteResponse: false}

	out, err := tls_client_cffi_src.BuildResponse("", false, resp, nil, input)
	require.Nil(t, err)
	assert.Equal(t, payload, out.Body)
}
