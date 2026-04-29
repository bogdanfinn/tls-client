package tests

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/httptest"
	tls_client "github.com/bogdanfinn/tls-client"
	tls_client_cffi_src "github.com/bogdanfinn/tls-client/cffi_src"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
// helpers
// -----------------------------------------------------------------------------

// streamMakeResp builds a minimal *http.Response usable as input to StartStream.
//
// uncompressed=true tells the stream to skip DecompressBodyByType (i.e. the
// body is already plaintext); set it to false together with a Content-Encoding
// header to exercise gzip decompression.
func streamMakeResp(body io.ReadCloser, contentType, contentEncoding string, uncompressed bool) *http.Response {
	h := http.Header{}
	if contentType != "" {
		h.Set("Content-Type", contentType)
	}
	if contentEncoding != "" {
		h.Set("Content-Encoding", contentEncoding)
	}
	u, _ := url.Parse("https://example.test/streaming")
	return &http.Response{
		Body:         body,
		Header:       h,
		StatusCode:   200,
		Proto:        "HTTP/1.1",
		Request:      &http.Request{URL: u},
		Uncompressed: uncompressed,
	}
}

// streamStart mints a fresh stream id, registers a cleanup that cancels it,
// and returns the id + state.
func streamStart(t *testing.T, params tls_client_cffi_src.StartStreamParams) (string, *tls_client_cffi_src.StreamState) {
	t.Helper()
	streamId := fmt.Sprintf("stream-%s", t.Name())
	state := tls_client_cffi_src.StartStream(streamId, params)
	t.Cleanup(func() { tls_client_cffi_src.CancelStream(streamId) })
	return streamId, state
}

func streamDecodeChunk(t *testing.T, out tls_client_cffi_src.StreamChunkResponse) []byte {
	t.Helper()
	if out.Chunk == "" {
		return nil
	}
	b, err := base64.StdEncoding.DecodeString(out.Chunk)
	require.NoError(t, err, "base64 decode of chunk")
	return b
}

// streamDrain reads chunks until EOF or error, concatenating their bytes. It
// returns the assembled bytes and the final terminating chunk response.
func streamDrain(t *testing.T, streamId string) ([]byte, tls_client_cffi_src.StreamChunkResponse) {
	t.Helper()
	var buf bytes.Buffer
	deadline := time.Now().Add(5 * time.Second)
	for {
		require.False(t, time.Now().After(deadline), "drain deadline exceeded; pump appears stuck")
		out := tls_client_cffi_src.ReadStreamChunk(streamId, 1000)
		if out.Error != "" {
			return buf.Bytes(), out
		}
		if out.Timeout {
			continue
		}
		buf.Write(streamDecodeChunk(t, out))
		if out.EOF {
			return buf.Bytes(), out
		}
	}
}

// -----------------------------------------------------------------------------
// ReadStreamChunk
// -----------------------------------------------------------------------------

func TestStream_DrainsBodyThenEOF(t *testing.T) {
	tls_client_cffi_src.ClearStreamCache()

	body := io.NopCloser(strings.NewReader("hello world"))
	resp := streamMakeResp(body, "text/plain; charset=utf-8", "", true)

	streamId, _ := streamStart(t, tls_client_cffi_src.StartStreamParams{
		Response:  resp,
		Cancel:    func() {},
		BlockSize: 1024,
	})

	got, terminal := streamDrain(t, streamId)
	assert.True(t, terminal.EOF, "expected EOF terminal chunk")
	assert.Equal(t, "hello world", string(got))

	// Subsequent reads on a finished stream surface as Error: unknown streamId.
	out := tls_client_cffi_src.ReadStreamChunk(streamId, 100)
	assert.NotEmpty(t, out.Error, "expected unknown-streamId error after EOF")

	_, ok := tls_client_cffi_src.GetStream(streamId)
	assert.False(t, ok, "stream still in registry after EOF")
}

func TestStream_ChunksByBlockSize(t *testing.T) {
	tls_client_cffi_src.ClearStreamCache()

	const blockSize = 16
	const total = 100
	payload := bytes.Repeat([]byte{'x'}, total)
	body := io.NopCloser(bytes.NewReader(payload))
	resp := streamMakeResp(body, "application/octet-stream", "", true)

	streamId, _ := streamStart(t, tls_client_cffi_src.StartStreamParams{
		Response:  resp,
		Cancel:    func() {},
		BlockSize: blockSize,
	})

	got, terminal := streamDrain(t, streamId)
	assert.True(t, terminal.EOF, "expected EOF")
	assert.Equal(t, payload, got)
}

func TestStream_UnknownStreamId(t *testing.T) {
	tls_client_cffi_src.ClearStreamCache()

	out := tls_client_cffi_src.ReadStreamChunk("does-not-exist", 100)
	assert.NotEmpty(t, out.Error, "expected error for unknown streamId")
	assert.Contains(t, out.Error, "unknown streamId")
}

func TestStream_NonBlockingPollNoData(t *testing.T) {
	tls_client_cffi_src.ClearStreamCache()

	pr, _ := io.Pipe()
	resp := streamMakeResp(io.NopCloser(pr), "text/event-stream", "", true)

	streamId, _ := streamStart(t, tls_client_cffi_src.StartStreamParams{
		Response:  resp,
		Cancel:    func() { _ = pr.Close() },
		BlockSize: 64,
	})

	// Pump is blocked on pipe.Read; channel is empty.
	out := tls_client_cffi_src.ReadStreamChunk(streamId, 0)
	assert.True(t, out.Timeout, "expected Timeout=true on non-blocking poll with no data")
	assert.Empty(t, out.Chunk)
	assert.False(t, out.EOF)
	assert.Empty(t, out.Error)
}

func TestStream_TimeoutThenData(t *testing.T) {
	tls_client_cffi_src.ClearStreamCache()

	pr, pw := io.Pipe()
	resp := streamMakeResp(io.NopCloser(pr), "text/event-stream", "", true)

	streamId, _ := streamStart(t, tls_client_cffi_src.StartStreamParams{
		Response:  resp,
		Cancel:    func() { _ = pr.Close() },
		BlockSize: 64,
	})

	// Initial read with no data → Timeout.
	out := tls_client_cffi_src.ReadStreamChunk(streamId, 100)
	assert.True(t, out.Timeout, "expected Timeout=true")

	// Now produce some data.
	go func() { _, _ = pw.Write([]byte("data: ping\n\n")) }()

	out = tls_client_cffi_src.ReadStreamChunk(streamId, 1000)
	require.Empty(t, out.Error, "read error")
	assert.False(t, out.Timeout, "expected data, got Timeout")
	assert.Equal(t, "data: ping\n\n", string(streamDecodeChunk(t, out)))
}

func TestStream_BlockingNegativeTimeoutWaitsForData(t *testing.T) {
	tls_client_cffi_src.ClearStreamCache()

	pr, pw := io.Pipe()
	resp := streamMakeResp(io.NopCloser(pr), "text/event-stream", "", true)

	streamId, _ := streamStart(t, tls_client_cffi_src.StartStreamParams{
		Response:  resp,
		Cancel:    func() { _ = pr.Close() },
		BlockSize: 64,
	})

	const delay = 200 * time.Millisecond
	const payload = "delayed"
	go func() {
		time.Sleep(delay)
		_, _ = pw.Write([]byte(payload))
	}()

	start := time.Now()
	out := tls_client_cffi_src.ReadStreamChunk(streamId, -1) // block
	elapsed := time.Since(start)

	require.Empty(t, out.Error, "read error")
	assert.Equal(t, payload, string(streamDecodeChunk(t, out)))
	assert.GreaterOrEqual(t, elapsed, delay, "ReadStreamChunk(-1) returned too early; did it actually block?")
}

func TestStream_SurfacesReadError(t *testing.T) {
	tls_client_cffi_src.ClearStreamCache()

	pr, pw := io.Pipe()
	wantErr := errors.New("synthetic body failure")

	resp := streamMakeResp(io.NopCloser(pr), "text/event-stream", "", true)

	streamId, _ := streamStart(t, tls_client_cffi_src.StartStreamParams{
		Response:  resp,
		Cancel:    func() { _ = pr.Close() },
		BlockSize: 64,
	})

	// Producer writes data then errors.
	go func() {
		_, _ = pw.Write([]byte("partial"))
		_ = pw.CloseWithError(wantErr)
	}()

	got, terminal := streamDrain(t, streamId)
	assert.Equal(t, "partial", string(got), "data buffered before error must still be delivered")
	assert.NotEmpty(t, terminal.Error, "expected error terminator")
	assert.Contains(t, terminal.Error, wantErr.Error())

	// Stream is removed after error.
	_, ok := tls_client_cffi_src.GetStream(streamId)
	assert.False(t, ok, "stream still registered after error")
}

// -----------------------------------------------------------------------------
// CancelStream
// -----------------------------------------------------------------------------

func TestStream_CancelRemovesAndReleases(t *testing.T) {
	tls_client_cffi_src.ClearStreamCache()

	pr, pw := io.Pipe()
	cancelCalled := make(chan struct{})
	var cancelOnce sync.Once
	resp := streamMakeResp(io.NopCloser(pr), "text/event-stream", "", true)

	streamId, _ := streamStart(t, tls_client_cffi_src.StartStreamParams{
		Response: resp,
		Cancel: func() {
			cancelOnce.Do(func() {
				close(cancelCalled)
				_ = pr.Close() // simulate ctx cancel propagating to body
			})
		},
		BlockSize: 64,
	})

	go func() { _, _ = pw.Write([]byte("hello")) }()

	out := tls_client_cffi_src.ReadStreamChunk(streamId, 1000)
	require.Empty(t, out.Error, "first read failed")
	assert.Equal(t, "hello", string(streamDecodeChunk(t, out)))

	tls_client_cffi_src.CancelStream(streamId)

	select {
	case <-cancelCalled:
	case <-time.After(time.Second):
		t.Fatal("cancel func was not invoked within 1s")
	}

	_, ok := tls_client_cffi_src.GetStream(streamId)
	assert.False(t, ok, "stream still registered after cancel")

	out = tls_client_cffi_src.ReadStreamChunk(streamId, 100)
	assert.NotEmpty(t, out.Error, "expected error reading cancelled stream")
}

func TestStream_CancelIsIdempotent(t *testing.T) {
	tls_client_cffi_src.ClearStreamCache()

	// Unknown id: must not panic, must not error (it's intentionally a no-op).
	tls_client_cffi_src.CancelStream("never-existed")

	// Cancel a real stream twice.
	body := io.NopCloser(strings.NewReader("data"))
	resp := streamMakeResp(body, "text/plain", "", true)
	streamId, _ := streamStart(t, tls_client_cffi_src.StartStreamParams{
		Response:  resp,
		Cancel:    func() {},
		BlockSize: 64,
	})

	tls_client_cffi_src.CancelStream(streamId)
	tls_client_cffi_src.CancelStream(streamId) // second call must not panic
}

// -----------------------------------------------------------------------------
// ReadStreamAll
// -----------------------------------------------------------------------------

func TestStream_ReadStreamAllTextBody(t *testing.T) {
	tls_client_cffi_src.ClearStreamCache()

	const payload = `{"hello":"world"}`
	body := io.NopCloser(strings.NewReader(payload))
	resp := streamMakeResp(body, "application/json; charset=utf-8", "", true)

	streamId, _ := streamStart(t, tls_client_cffi_src.StartStreamParams{
		Response:  resp,
		Cancel:    func() {},
		BlockSize: 64,
	})

	response, drainErr := tls_client_cffi_src.ReadStreamAll(streamId)
	require.Nil(t, drainErr, "ReadStreamAll failed")

	assert.Equal(t, payload, response.Body)
	assert.Equal(t, 200, response.Status)
	assert.NotEmpty(t, response.Headers["Content-Type"], "Content-Type header should be preserved")

	_, ok := tls_client_cffi_src.GetStream(streamId)
	assert.False(t, ok, "stream should be removed after ReadStreamAll")
}

func TestStream_ReadStreamAllByteResponseBase64(t *testing.T) {
	tls_client_cffi_src.ClearStreamCache()

	// Binary content that does NOT survive a UTF-8 charset round-trip, so
	// the byte path is genuinely exercised.
	payload := []byte{0xff, 0xfe, 0x00, 0x01, 0x80, 0x00, 0x42}
	body := io.NopCloser(bytes.NewReader(payload))
	resp := streamMakeResp(body, "", "", true)

	streamId, _ := streamStart(t, tls_client_cffi_src.StartStreamParams{
		Response:       resp,
		Cancel:         func() {},
		BlockSize:      64,
		IsByteResponse: true,
	})

	response, drainErr := tls_client_cffi_src.ReadStreamAll(streamId)
	require.Nil(t, drainErr)

	prefix, encoded, found := strings.Cut(response.Body, ";base64,")
	require.True(t, found, "byte response body should be a data URL; got %q", response.Body)
	assert.True(t, strings.HasPrefix(prefix, "data:"))

	got, err := base64.StdEncoding.DecodeString(encoded)
	require.NoError(t, err, "base64 decode")
	assert.Equal(t, payload, got)
}

func TestStream_ReadStreamAllAfterPartialReads(t *testing.T) {
	tls_client_cffi_src.ClearStreamCache()

	const payload = "abcdefghij" // 10 bytes
	body := io.NopCloser(strings.NewReader(payload))
	resp := streamMakeResp(body, "text/plain; charset=utf-8", "", true)

	streamId, _ := streamStart(t, tls_client_cffi_src.StartStreamParams{
		Response:  resp,
		Cancel:    func() {},
		BlockSize: 4, // forces multiple chunks
	})

	first := tls_client_cffi_src.ReadStreamChunk(streamId, 1000)
	require.Empty(t, first.Error, "first read failed")
	firstBytes := streamDecodeChunk(t, first)
	assert.NotEmpty(t, firstBytes, "first chunk must contain bytes")
	assert.LessOrEqual(t, len(firstBytes), 4, "first chunk must respect block size")

	rest, err := tls_client_cffi_src.ReadStreamAll(streamId)
	require.Nil(t, err, "ReadStreamAll failed")

	assert.Equal(t, payload, string(firstBytes)+rest.Body)
}

// -----------------------------------------------------------------------------
// gzip decompression
// -----------------------------------------------------------------------------

func TestStream_DecompressesGzipBodies(t *testing.T) {
	tls_client_cffi_src.ClearStreamCache()

	const original = "the quick brown fox jumps over the lazy dog"

	var compressed bytes.Buffer
	gz := gzip.NewWriter(&compressed)
	_, err := gz.Write([]byte(original))
	require.NoError(t, err, "gzip write")
	require.NoError(t, gz.Close(), "gzip close")

	body := io.NopCloser(bytes.NewReader(compressed.Bytes()))
	// uncompressed=false → DecompressBodyByType is applied inside StartStream.
	resp := streamMakeResp(body, "text/plain; charset=utf-8", "gzip", false)

	streamId, _ := streamStart(t, tls_client_cffi_src.StartStreamParams{
		Response:  resp,
		Cancel:    func() {},
		BlockSize: 64,
	})

	got, terminal := streamDrain(t, streamId)
	assert.True(t, terminal.EOF, "expected EOF")
	assert.Equal(t, original, string(got), "body should be transparently decompressed")
}

// -----------------------------------------------------------------------------
// BuildStreamStartResponse / ClearStreamCache
// -----------------------------------------------------------------------------

func TestStream_BuildStreamStartResponseEnvelope(t *testing.T) {
	tls_client_cffi_src.ClearStreamCache()

	body := io.NopCloser(strings.NewReader(""))
	resp := streamMakeResp(body, "text/event-stream", "", true)
	resp.Header.Set("X-Custom", "hello")
	resp.StatusCode = 202
	resp.Proto = "HTTP/2.0"

	streamId, state := streamStart(t, tls_client_cffi_src.StartStreamParams{
		Response:    resp,
		Cancel:      func() {},
		SessionId:   "sess-42",
		WithSession: true,
		BlockSize:   64,
	})

	out := tls_client_cffi_src.BuildStreamStartResponse(streamId, state)

	assert.Equal(t, streamId, out.StreamId)
	assert.Equal(t, 202, out.Status)
	assert.Equal(t, "HTTP/2.0", out.UsedProtocol)
	assert.Equal(t, "sess-42", out.SessionId)
	assert.Equal(t, []string{"hello"}, out.Headers["X-Custom"])
	assert.Empty(t, out.Body, "Body must be empty for stream start envelope")
	assert.NotEmpty(t, out.Target, "Target should be the request URL")
	assert.NotEmpty(t, out.Id, "envelope must have an Id (used for freeMemory)")
}

func TestStream_ClearStreamCacheRemovesAll(t *testing.T) {
	tls_client_cffi_src.ClearStreamCache()

	// Use pipes so the pump goroutines block — i.e. streams stay alive
	// rather than racing to EOF before ClearStreamCache runs.
	pipes := make([]*io.PipeWriter, 0, 3)
	ids := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		pr, pw := io.Pipe()
		pipes = append(pipes, pw)
		resp := streamMakeResp(io.NopCloser(pr), "text/event-stream", "", true)
		streamId := fmt.Sprintf("clear-cache-%d", i)
		tls_client_cffi_src.StartStream(streamId, tls_client_cffi_src.StartStreamParams{
			Response:  resp,
			Cancel:    func() { _ = pr.Close() },
			BlockSize: 64,
		})
		ids = append(ids, streamId)
	}

	for _, id := range ids {
		_, ok := tls_client_cffi_src.GetStream(id)
		require.True(t, ok, "stream %q not registered before ClearStreamCache", id)
	}

	tls_client_cffi_src.ClearStreamCache()

	for _, id := range ids {
		_, ok := tls_client_cffi_src.GetStream(id)
		assert.False(t, ok, "stream %q still registered after ClearStreamCache", id)
	}

	for _, pw := range pipes {
		_ = pw.Close()
	}
}

// -----------------------------------------------------------------------------
// End-to-end: drive a real fhttp client through StartStream + ReadStreamChunk
// against an httptest server that flushes SSE-style events.
// -----------------------------------------------------------------------------

func TestStream_EndToEndSSEServer(t *testing.T) {
	tls_client_cffi_src.ClearStreamCache()

	const eventCount = 3

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("test server must support flushing")
		}
		for i := 0; i < eventCount; i++ {
			fmt.Fprintf(w, "data: event-%d\n\n", i)
			flusher.Flush()
			time.Sleep(20 * time.Millisecond)
		}
	}))
	defer server.Close()

	tlsClient, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(),
		tls_client.WithTimeoutSeconds(10),
	)
	require.NoError(t, err, "NewHttpClient")

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err, "NewRequest")

	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)
	t.Cleanup(cancel)

	resp, err := tlsClient.Do(req)
	require.NoError(t, err, "client.Do")

	streamId := "e2e-sse"
	tls_client_cffi_src.StartStream(streamId, tls_client_cffi_src.StartStreamParams{
		Response:  resp,
		Cancel:    cancel,
		BlockSize: 4096,
	})
	t.Cleanup(func() { tls_client_cffi_src.CancelStream(streamId) })

	var got bytes.Buffer
	deadline := time.Now().Add(5 * time.Second)
	for {
		require.False(t, time.Now().After(deadline), "e2e deadline exceeded")
		out := tls_client_cffi_src.ReadStreamChunk(streamId, 500)
		require.Empty(t, out.Error, "read error: %s", out.Error)
		if out.Chunk != "" {
			got.Write(streamDecodeChunk(t, out))
		}
		if out.EOF {
			break
		}
	}

	body := got.String()
	for i := 0; i < eventCount; i++ {
		assert.Contains(t, body, fmt.Sprintf("data: event-%d\n\n", i),
			"missing SSE event %d in assembled stream:\n%s", i, body)
	}
}
