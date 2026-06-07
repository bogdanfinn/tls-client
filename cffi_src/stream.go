package tls_client_cffi_src

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"sync"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/google/uuid"
	"golang.org/x/net/html/charset"
)

// backpressureTimeout bounds how long the pump goroutine will block trying to
// hand off a chunk before giving up. It exists so a misbehaving caller (one
// that stops reading without calling cancelStream) cannot keep the underlying
// connection and goroutine alive indefinitely.
const backpressureTimeout = 60 * time.Second

// defaultStreamBlockSize is the per-Read buffer size when StreamOutputBlockSize
// is not specified on the RequestInput.
const defaultStreamBlockSize = 4096

// streamChannelCapacity is the buffered capacity of the chunk channel between
// the pump goroutine and the readStream caller. Higher values smooth out short
// caller stalls; lower values keep memory bounded.
const streamChannelCapacity = 64

// StreamChunk is one item produced by the pump goroutine and consumed by
// readStream / readStreamAll. Exactly one of Data, EOF, or Err is meaningful.
type StreamChunk struct {
	Err  error
	Data []byte
	EOF  bool
}

// StreamState tracks one in-flight streaming response.
//
// Headers/status/cookies are captured at start time so readStreamAll can
// produce a complete Response envelope without the caller having to merge
// the requestStream envelope back in.
type StreamState struct {
	cancel      context.CancelFunc
	body        io.ReadCloser
	Chunks      chan StreamChunk
	Headers     http.Header
	Cookies     []*http.Cookie
	SessionId   string
	Proto       string
	Target      string
	ContentType string
	Status      int
	WithSession bool

	// IsByteResponse mirrors the original RequestInput so readStreamAll
	// formats the body the same way the non-streaming path would have.
	IsByteResponse bool

	closeOnce sync.Once
}

// StartStreamParams bundles the inputs needed to register a streaming
// response. The caller is responsible for having already issued the request
// and obtained the response headers.
type StartStreamParams struct {
	Response       *http.Response
	Cookies        []*http.Cookie
	Cancel         context.CancelFunc
	SessionId      string
	WithSession    bool
	IsByteResponse bool
	BlockSize      int // <= 0 falls back to defaultStreamBlockSize
}

// StartStream wraps the response body with decompression (when applicable),
// constructs a StreamState, registers it under streamId, and spawns the pump
// goroutine. It returns the registered StreamState so the caller can build
// the requestStream envelope.
func StartStream(streamId string, p StartStreamParams) *StreamState {
	resp := p.Response

	body := resp.Body
	if !resp.Uncompressed {
		body = http.DecompressBodyByType(body, resp.Header.Get("Content-Encoding"))
	}

	target := ""
	if resp.Request != nil && resp.Request.URL != nil {
		target = resp.Request.URL.String()
	}

	state := &StreamState{
		cancel:         p.Cancel,
		body:           body,
		Chunks:         make(chan StreamChunk, streamChannelCapacity),
		Headers:        resp.Header,
		Cookies:        p.Cookies,
		SessionId:      p.SessionId,
		Proto:          resp.Proto,
		Target:         target,
		ContentType:    resp.Header.Get("Content-Type"),
		Status:         resp.StatusCode,
		WithSession:    p.WithSession,
		IsByteResponse: p.IsByteResponse,
	}

	RegisterStream(streamId, state)
	go pumpStream(state, p.BlockSize)
	return state
}

// Close is idempotent. It cancels the request context (which closes the
// underlying TCP/TLS connection and unblocks any in-flight body Read).
func (s *StreamState) Close() {
	s.closeOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
	})
}

var (
	streamsLock sync.Mutex
	streams     = make(map[string]*StreamState)
)

// RegisterStream stores a freshly created StreamState under streamId. The
// caller takes responsibility for spawning the pump goroutine.
func RegisterStream(streamId string, state *StreamState) {
	streamsLock.Lock()
	defer streamsLock.Unlock()
	streams[streamId] = state
}

// GetStream returns the StreamState for streamId, or (nil, false) if none.
func GetStream(streamId string) (*StreamState, bool) {
	streamsLock.Lock()
	defer streamsLock.Unlock()
	s, ok := streams[streamId]
	return s, ok
}

// removeStream deletes the registry entry for streamId and returns the
// previously stored StreamState. Callers should also invoke Close on the
// returned state to release the underlying connection.
func removeStream(streamId string) (*StreamState, bool) {
	streamsLock.Lock()
	defer streamsLock.Unlock()
	s, ok := streams[streamId]
	if ok {
		delete(streams, streamId)
	}
	return s, ok
}

// ClearStreamCache cancels and removes every registered stream. It is safe to
// call from destroyAll.
func ClearStreamCache() {
	streamsLock.Lock()
	toClose := make([]*StreamState, 0, len(streams))
	for _, s := range streams {
		toClose = append(toClose, s)
	}
	streams = make(map[string]*StreamState)
	streamsLock.Unlock()

	for _, s := range toClose {
		s.Close()
	}
}

// pumpStream reads from state.body in blockSize chunks and pushes them onto
// state.Chunks. It terminates on EOF, on read error, or when the consumer
// stalls for longer than backpressureTimeout. The chunk channel is closed
// before the goroutine returns.
func pumpStream(state *StreamState, blockSize int) {
	defer func() {
		_ = state.body.Close()
		close(state.Chunks)
	}()

	if blockSize <= 0 {
		blockSize = defaultStreamBlockSize
	}

	for {
		buf := make([]byte, blockSize)
		n, readErr := state.body.Read(buf)

		if n > 0 {
			if !sendChunk(state.Chunks, StreamChunk{Data: buf[:n]}) {
				return
			}
		}

		if readErr == io.EOF {
			sendChunk(state.Chunks, StreamChunk{EOF: true})
			return
		}
		if readErr != nil {
			sendChunk(state.Chunks, StreamChunk{Err: readErr})
			return
		}
	}
}

func sendChunk(ch chan<- StreamChunk, chunk StreamChunk) bool {
	select {
	case ch <- chunk:
		return true
	case <-time.After(backpressureTimeout):
		return false
	}
}

// BuildStreamStartResponse constructs the envelope returned by requestStream.
// The caller is expected to register the StreamState and spawn the pump after
// this returns.
func BuildStreamStartResponse(streamId string, state *StreamState) StreamStartResponse {
	response := Response{
		Id:           uuid.New().String(),
		Status:       state.Status,
		UsedProtocol: state.Proto,
		Body:         "",
		Headers:      state.Headers,
		Target:       state.Target,
		Cookies:      cookiesToMap(state.Cookies),
	}

	if state.WithSession {
		response.SessionId = state.SessionId
	}

	return StreamStartResponse{
		Response: response,
		StreamId: streamId,
	}
}

// ReadStreamChunk pulls the next chunk from the registry, applying timeoutMs
// semantics as documented on ReadStreamInput. On EOF or error it removes the
// stream from the registry and closes its underlying connection.
func ReadStreamChunk(streamId string, timeoutMs int) StreamChunkResponse {
	state, ok := GetStream(streamId)
	if !ok {
		return StreamChunkResponse{
			Id:       uuid.New().String(),
			StreamId: streamId,
			Error:    fmt.Sprintf("unknown streamId: %s", streamId),
		}
	}

	switch {
	case timeoutMs < 0:
		chunk, open := <-state.Chunks
		return finalizeChunk(streamId, state, chunk, open)
	case timeoutMs == 0:
		select {
		case chunk, open := <-state.Chunks:
			return finalizeChunk(streamId, state, chunk, open)
		default:
			return StreamChunkResponse{
				Id:       uuid.New().String(),
				StreamId: streamId,
				Timeout:  true,
			}
		}
	default:
		select {
		case chunk, open := <-state.Chunks:
			return finalizeChunk(streamId, state, chunk, open)
		case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
			return StreamChunkResponse{
				Id:       uuid.New().String(),
				StreamId: streamId,
				Timeout:  true,
			}
		}
	}
}

func finalizeChunk(streamId string, state *StreamState, chunk StreamChunk, open bool) StreamChunkResponse {
	if !open {
		// Channel was closed without a final EOF/Err — pump exited via
		// backpressure timeout. Treat as a non-fatal EOF for the caller.
		_, _ = removeStream(streamId)
		state.Close()
		return StreamChunkResponse{
			Id:       uuid.New().String(),
			StreamId: streamId,
			EOF:      true,
		}
	}

	if chunk.Err != nil {
		_, _ = removeStream(streamId)
		state.Close()
		return StreamChunkResponse{
			Id:       uuid.New().String(),
			StreamId: streamId,
			Error:    chunk.Err.Error(),
		}
	}

	if chunk.EOF {
		_, _ = removeStream(streamId)
		state.Close()
		return StreamChunkResponse{
			Id:       uuid.New().String(),
			StreamId: streamId,
			EOF:      true,
		}
	}

	return StreamChunkResponse{
		Id:       uuid.New().String(),
		StreamId: streamId,
		Chunk:    base64.StdEncoding.EncodeToString(chunk.Data),
	}
}

// ReadStreamAll drains the rest of the body in one shot and returns a full
// Response envelope, formatted the same way BuildResponse would have for a
// non-streaming request (charset decoding for text bodies, base64 + MIME
// prefix for byte responses). After this returns, streamId is invalid.
func ReadStreamAll(streamId string) (Response, *TLSClientError) {
	state, ok := GetStream(streamId)
	if !ok {
		return Response{}, NewTLSClientError(fmt.Errorf("unknown streamId: %s", streamId))
	}

	var buf bytes.Buffer
	var streamErr error

	for chunk := range state.Chunks {
		if chunk.Err != nil {
			streamErr = chunk.Err
			break
		}
		if len(chunk.Data) > 0 {
			buf.Write(chunk.Data)
		}
		if chunk.EOF {
			break
		}
	}

	_, _ = removeStream(streamId)
	state.Close()

	if streamErr != nil {
		return Response{}, NewTLSClientError(streamErr)
	}

	bodyBytes := buf.Bytes()
	finalBody := ""

	if state.IsByteResponse {
		mimeType := http.DetectContentType(bodyBytes)
		finalBody = fmt.Sprintf("data:%s;base64,", mimeType) + base64.StdEncoding.EncodeToString(bodyBytes)
	} else if len(bodyBytes) > 0 {
		// Mirror the charset detection BuildResponse does for non-byte bodies.
		reader, err := charset.NewReader(bytes.NewReader(bodyBytes), state.ContentType)
		if err != nil {
			return Response{}, NewTLSClientError(err)
		}
		decoded, err := io.ReadAll(reader)
		if err != nil {
			return Response{}, NewTLSClientError(err)
		}
		finalBody = string(decoded)
	}

	response := Response{
		Id:           uuid.New().String(),
		Status:       state.Status,
		UsedProtocol: state.Proto,
		Body:         finalBody,
		Headers:      state.Headers,
		Target:       state.Target,
		Cookies:      cookiesToMap(state.Cookies),
	}
	if state.WithSession {
		response.SessionId = state.SessionId
	}
	return response, nil
}

// CancelStream tears down the stream identified by streamId. It is idempotent:
// calling it after a natural EOF, after an error, or with an unknown streamId
// returns success without doing any work.
func CancelStream(streamId string) {
	state, ok := removeStream(streamId)
	if !ok {
		return
	}
	state.Close()
	// Drain any remaining chunks so the pump goroutine, which is blocked on
	// sendChunk, can terminate promptly.
	go func() {
		for range state.Chunks {
		}
	}()
}
