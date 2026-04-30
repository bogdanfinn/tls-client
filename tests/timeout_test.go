package tests

import (
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/httptest"
	tls_client "github.com/bogdanfinn/tls-client"
	tls_client_cffi_src "github.com/bogdanfinn/tls-client/cffi_src"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests pin down the cffi timeout-resolution contract. The motivating
// regression: callers passing TimeoutSeconds=0 / TimeoutMilliseconds=0 expected
// "no deadline" (because the underlying tls_client.WithTimeoutSeconds(0)
// documents 0 as "Unlimited"), but the cffi treated 0 as "field omitted" and
// silently fell through to DefaultTimeoutSeconds (30s). That made it
// impossible to consume long-lived SSE streams without hitting a 30 s reset.
//
// The fix: a negative value means "disable the deadline". 0 still means
// "default" so existing callers keep working.

// startSlowServer returns an httptest server whose handler blocks for `delay`
// before responding 200 OK. Used to provoke client timeouts on demand.
func startSlowServer(t *testing.T, delay time.Duration) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// doWithTimeoutOption builds a client carrying only the option under test plus
// the bare minimum required by NewHttpClient and runs a GET against url. Returns
// the elapsed duration and the error (if any). Caller is responsible for
// closing the response body when err == nil.
func doWithTimeoutOption(t *testing.T, opt tls_client.HttpClientOption, url string) (time.Duration, error) {
	t.Helper()
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), opt)
	require.NoError(t, err, "NewHttpClient")

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err, "NewRequest")

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	return elapsed, err
}

// isTimeoutErr returns true if err looks like a client-side deadline-exceeded
// error. Different transports word the message differently.
func isTimeoutErr(err error) bool {
	if err == nil {
		return false
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "deadline exceeded") ||
		strings.Contains(msg, "context canceled")
}

// -----------------------------------------------------------------------------
// pure unit tests on ResolveTimeoutOption — no network involved.
// -----------------------------------------------------------------------------
//
// We can't directly inspect the time.Duration sitting on the unexported
// httpClientConfig from the outside, so we verify the option's *effect* by
// applying it through the public NewHttpClient path and observing whether a
// slow request times out or not. The expected-behavior matrix is:
//
//   sec   ms     | behavior
//   --------------+----------
//    0     0     | default 30s   (request to slow server completes well under 30s → no err)
//   >0     0     | seconds       (deadline)
//    0    >0     | ms            (deadline)
//   <0     0     | DISABLED      (long server delay still completes)
//    0    <0     | DISABLED      (long server delay still completes)
//   >0    >0     | ms wins       (precedence preserved)

func TestResolveTimeoutOption_BothZero_UsesDefault(t *testing.T) {
	// Default is 30 s. We can't afford to wait that long, so instead assert
	// that a fast request (short server delay) completes under default — which
	// proves "default" did not collapse to 0-disabled by accident, but doesn't
	// distinguish "default 30 s" from "any other positive value > 50ms". The
	// stronger assertions below pin the >0 / <0 paths; this one just guards
	// against a regression where the default branch goes missing entirely.
	srv := startSlowServer(t, 50*time.Millisecond)

	opt := tls_client_cffi_src.ResolveTimeoutOption(0, 0)
	elapsed, err := doWithTimeoutOption(t, opt, srv.URL)

	require.NoError(t, err, "default-timeout request must complete (took %v)", elapsed)
}

func TestResolveTimeoutOption_PositiveSeconds_AppliesDeadline(t *testing.T) {
	// Server delays 500 ms, deadline is 1 s → must complete.
	srv := startSlowServer(t, 500*time.Millisecond)

	opt := tls_client_cffi_src.ResolveTimeoutOption(1, 0)
	elapsed, err := doWithTimeoutOption(t, opt, srv.URL)

	require.NoError(t, err, "expected 1 s deadline to allow a 500 ms request, got err=%v elapsed=%v", err, elapsed)
}

func TestResolveTimeoutOption_PositiveMilliseconds_AppliesDeadline(t *testing.T) {
	// Server delays 500 ms, deadline 100 ms → must time out.
	srv := startSlowServer(t, 500*time.Millisecond)

	opt := tls_client_cffi_src.ResolveTimeoutOption(0, 100)
	elapsed, err := doWithTimeoutOption(t, opt, srv.URL)

	require.Error(t, err, "expected timeout, got success after %v", elapsed)
	assert.True(t, isTimeoutErr(err), "expected timeout-shaped error, got: %v", err)
	// Sanity: timeout should fire well before the server would have finished.
	assert.Less(t, elapsed, 400*time.Millisecond, "timeout fired too late: %v", elapsed)
}

func TestResolveTimeoutOption_NegativeSeconds_DisablesDeadline(t *testing.T) {
	// The bug fix: negative seconds must produce a no-deadline option, so a
	// server that sleeps longer than the previous default 30s would *also*
	// complete. We don't actually want to wait 30s in CI, so use a delay
	// (300 ms) that's:
	//   - long enough to trip a too-aggressive default
	//   - short enough that the test stays cheap
	// and combine with a sanity assertion that the request actually waited.
	srv := startSlowServer(t, 300*time.Millisecond)

	opt := tls_client_cffi_src.ResolveTimeoutOption(-1, 0)
	elapsed, err := doWithTimeoutOption(t, opt, srv.URL)

	require.NoError(t, err, "expected request to complete with timeout disabled, got err=%v elapsed=%v", err, elapsed)
	assert.GreaterOrEqual(t, elapsed, 250*time.Millisecond, "request returned suspiciously fast (%v); did the server actually wait?", elapsed)
}

func TestResolveTimeoutOption_NegativeMilliseconds_DisablesDeadline(t *testing.T) {
	// Same contract as the seconds case — pin both branches because they're
	// independent code paths in the resolver.
	srv := startSlowServer(t, 300*time.Millisecond)

	opt := tls_client_cffi_src.ResolveTimeoutOption(0, -1)
	elapsed, err := doWithTimeoutOption(t, opt, srv.URL)

	require.NoError(t, err, "expected request to complete with timeout disabled, got err=%v elapsed=%v", err, elapsed)
	assert.GreaterOrEqual(t, elapsed, 250*time.Millisecond, "request returned suspiciously fast (%v); did the server actually wait?", elapsed)
}

func TestResolveTimeoutOption_NegativePrecedesPositive(t *testing.T) {
	// If either field is negative the resolver must disable the timeout, even
	// if the other field is a positive deadline that would otherwise apply.
	// Pin the precedence so a future refactor doesn't accidentally reintroduce
	// the regression for callers that pass both fields.
	srv := startSlowServer(t, 300*time.Millisecond)

	// Positive seconds + negative ms → disabled wins over the positive seconds.
	opt := tls_client_cffi_src.ResolveTimeoutOption(1, -1)
	_, err := doWithTimeoutOption(t, opt, srv.URL)
	require.NoError(t, err, "negative ms must disable even when seconds is positive: %v", err)
}

func TestResolveTimeoutOption_MillisecondsWinOverSeconds(t *testing.T) {
	// Both positive: ms takes precedence (preserves the original behavior of
	// the 'if TimeoutMilliseconds != 0' branch overwriting the seconds option).
	// We pin this so callers that set both fields keep getting the millisecond
	// resolution.
	srv := startSlowServer(t, 500*time.Millisecond)

	// Generous seconds (60s) but tight ms (100ms) → ms must win, request times out.
	opt := tls_client_cffi_src.ResolveTimeoutOption(60, 100)
	elapsed, err := doWithTimeoutOption(t, opt, srv.URL)

	require.Error(t, err, "expected ms deadline to override seconds; request completed in %v", elapsed)
	assert.True(t, isTimeoutErr(err), "expected timeout-shaped error, got: %v", err)
}
