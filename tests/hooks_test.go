package tests

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreHookModifiesRequestHeaders(t *testing.T) {
	client, err := tls_client.NewHttpClient(
		tls_client.NewNoopLogger(),
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithTimeoutSeconds(10),
		tls_client.WithPreHook(func(req *http.Request) error {
			req.Header.Set("X-Custom-Header", "test-value")
			return nil
		}),
	)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "https://httpbin.org/headers", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	// The header was set by the pre-hook
	assert.Equal(t, "test-value", req.Header.Get("X-Custom-Header"))
}

func TestPreHookErrorAbortsRequest(t *testing.T) {
	expectedErr := errors.New("pre-hook error")

	client, err := tls_client.NewHttpClient(
		tls_client.NewNoopLogger(),
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithTimeoutSeconds(10),
		tls_client.WithPreHook(func(req *http.Request) error {
			return expectedErr
		}),
	)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "https://httpbin.org/get", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	assert.Nil(t, resp)
	assert.Equal(t, expectedErr, err)
}

func TestPostHookReceivesCorrectMetadata(t *testing.T) {
	var capturedCtx *tls_client.PostResponseContext

	client, err := tls_client.NewHttpClient(
		tls_client.NewNoopLogger(),
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithTimeoutSeconds(10),
		tls_client.WithPostHook(func(ctx *tls_client.PostResponseContext) {
			capturedCtx = ctx
		}),
	)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "https://httpbin.org/get", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.NotNil(t, capturedCtx)
	assert.NotNil(t, capturedCtx.Request)
	assert.NotNil(t, capturedCtx.Response)
	assert.Equal(t, 200, capturedCtx.Response.StatusCode)
	assert.Nil(t, capturedCtx.Error)
}

func TestMultiplePreHooksExecuteInOrder(t *testing.T) {
	var order []int
	var mu sync.Mutex

	client, err := tls_client.NewHttpClient(
		tls_client.NewNoopLogger(),
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithTimeoutSeconds(10),
		tls_client.WithPreHook(func(req *http.Request) error {
			mu.Lock()
			order = append(order, 1)
			mu.Unlock()
			return nil
		}),
		tls_client.WithPreHook(func(req *http.Request) error {
			mu.Lock()
			order = append(order, 2)
			mu.Unlock()
			return nil
		}),
		tls_client.WithPreHook(func(req *http.Request) error {
			mu.Lock()
			order = append(order, 3)
			mu.Unlock()
			return nil
		}),
	)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "https://httpbin.org/get", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, []int{1, 2, 3}, order)
}

func TestMultiplePostHooksExecuteInOrder(t *testing.T) {
	var order []int
	var mu sync.Mutex

	client, err := tls_client.NewHttpClient(
		tls_client.NewNoopLogger(),
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithTimeoutSeconds(10),
		tls_client.WithPostHook(func(ctx *tls_client.PostResponseContext) {
			mu.Lock()
			order = append(order, 1)
			mu.Unlock()
		}),
		tls_client.WithPostHook(func(ctx *tls_client.PostResponseContext) {
			mu.Lock()
			order = append(order, 2)
			mu.Unlock()
		}),
		tls_client.WithPostHook(func(ctx *tls_client.PostResponseContext) {
			mu.Lock()
			order = append(order, 3)
			mu.Unlock()
		}),
	)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "https://httpbin.org/get", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, []int{1, 2, 3}, order)
}

func TestPreHookErrorStopsSubsequentHooks(t *testing.T) {
	var executedHooks []int
	var mu sync.Mutex
	expectedErr := errors.New("hook 2 error")

	client, err := tls_client.NewHttpClient(
		tls_client.NewNoopLogger(),
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithTimeoutSeconds(10),
		tls_client.WithPreHook(func(req *http.Request) error {
			mu.Lock()
			executedHooks = append(executedHooks, 1)
			mu.Unlock()
			return nil
		}),
		tls_client.WithPreHook(func(req *http.Request) error {
			mu.Lock()
			executedHooks = append(executedHooks, 2)
			mu.Unlock()
			return expectedErr
		}),
		tls_client.WithPreHook(func(req *http.Request) error {
			mu.Lock()
			executedHooks = append(executedHooks, 3)
			mu.Unlock()
			return nil
		}),
	)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "https://httpbin.org/get", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	assert.Nil(t, resp)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, []int{1, 2}, executedHooks) // Hook 3 should not execute
}

func TestPostHookAlwaysRunsAllHooks(t *testing.T) {
	var executedHooks []int
	var mu sync.Mutex

	client, err := tls_client.NewHttpClient(
		tls_client.NewNoopLogger(),
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithTimeoutSeconds(10),
		tls_client.WithPostHook(func(ctx *tls_client.PostResponseContext) {
			mu.Lock()
			executedHooks = append(executedHooks, 1)
			mu.Unlock()
		}),
		tls_client.WithPostHook(func(ctx *tls_client.PostResponseContext) {
			mu.Lock()
			executedHooks = append(executedHooks, 2)
			mu.Unlock()
			panic("intentional panic")
		}),
		tls_client.WithPostHook(func(ctx *tls_client.PostResponseContext) {
			mu.Lock()
			executedHooks = append(executedHooks, 3)
			mu.Unlock()
		}),
	)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "https://httpbin.org/get", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Hook 3 should not execute because hook 2 panicked
	assert.Equal(t, []int{1, 2}, executedHooks)
}

func TestPostHookReceivesErrorOnRequestFailure(t *testing.T) {
	var capturedCtx *tls_client.PostResponseContext

	client, err := tls_client.NewHttpClient(
		tls_client.NewNoopLogger(),
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithTimeoutSeconds(1),
		tls_client.WithPostHook(func(ctx *tls_client.PostResponseContext) {
			capturedCtx = ctx
		}),
	)
	require.NoError(t, err)

	// Request to a non-existent host to trigger an error
	req, err := http.NewRequest(http.MethodGet, "https://this-host-does-not-exist-12345.invalid/", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	assert.Nil(t, resp)
	assert.Error(t, err)

	require.NotNil(t, capturedCtx)
	assert.NotNil(t, capturedCtx.Error)
	assert.Nil(t, capturedCtx.Response)
}

func TestOnPreRequestRuntimeRegistration(t *testing.T) {
	client, err := tls_client.NewHttpClient(
		tls_client.NewNoopLogger(),
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithTimeoutSeconds(10),
	)
	require.NoError(t, err)

	var hookCalled bool
	client.AddPreRequestHook(func(req *http.Request) error {
		hookCalled = true
		req.Header.Set("X-Runtime-Hook", "added")
		return nil
	})

	req, err := http.NewRequest(http.MethodGet, "https://httpbin.org/get", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.True(t, hookCalled)
	assert.Equal(t, "added", req.Header.Get("X-Runtime-Hook"))
}

func TestOnPostResponseRuntimeRegistration(t *testing.T) {
	client, err := tls_client.NewHttpClient(
		tls_client.NewNoopLogger(),
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithTimeoutSeconds(10),
	)
	require.NoError(t, err)

	var capturedCtx *tls_client.PostResponseContext
	client.AddPostResponseHook(func(ctx *tls_client.PostResponseContext) {
		capturedCtx = ctx
	})

	req, err := http.NewRequest(http.MethodGet, "https://httpbin.org/get", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.NotNil(t, capturedCtx)
	assert.Equal(t, 200, capturedCtx.Response.StatusCode)
}

func TestPostHookNotCalledOnPreHookError(t *testing.T) {
	var postHookCalled bool
	preHookErr := errors.New("pre-hook failed")

	client, err := tls_client.NewHttpClient(
		tls_client.NewNoopLogger(),
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithTimeoutSeconds(10),
		tls_client.WithPreHook(func(req *http.Request) error {
			return preHookErr
		}),
		tls_client.WithPostHook(func(ctx *tls_client.PostResponseContext) {
			postHookCalled = true
		}),
	)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "https://httpbin.org/get", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	assert.Nil(t, resp)
	assert.Equal(t, preHookErr, err)

	// PostHook should NOT be called since no request was made
	assert.False(t, postHookCalled)
}

func TestHooksThreadSafety(t *testing.T) {
	client, err := tls_client.NewHttpClient(
		tls_client.NewNoopLogger(),
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithTimeoutSeconds(10),
	)
	require.NoError(t, err)

	var preHookCount int64
	var postHookCount int64

	// Add hooks from multiple goroutines
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.AddPreRequestHook(func(req *http.Request) error {
				atomic.AddInt64(&preHookCount, 1)
				return nil
			})
			client.AddPostResponseHook(func(ctx *tls_client.PostResponseContext) {
				atomic.AddInt64(&postHookCount, 1)
			})
		}()
	}
	wg.Wait()

	req, err := http.NewRequest(http.MethodGet, "https://httpbin.org/get", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// All 10 hooks should have been registered and called
	assert.Equal(t, int64(10), atomic.LoadInt64(&preHookCount))
	assert.Equal(t, int64(10), atomic.LoadInt64(&postHookCount))
}

func TestCombinedConstructorAndRuntimeHooks(t *testing.T) {
	var order []int
	var mu sync.Mutex

	client, err := tls_client.NewHttpClient(
		tls_client.NewNoopLogger(),
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithTimeoutSeconds(10),
		tls_client.WithPreHook(func(req *http.Request) error {
			mu.Lock()
			order = append(order, 1)
			mu.Unlock()
			return nil
		}),
		tls_client.WithPostHook(func(ctx *tls_client.PostResponseContext) {
			mu.Lock()
			order = append(order, 3)
			mu.Unlock()
		}),
	)
	require.NoError(t, err)

	// Add runtime hooks
	client.AddPreRequestHook(func(req *http.Request) error {
		mu.Lock()
		order = append(order, 2)
		mu.Unlock()
		return nil
	})
	client.AddPostResponseHook(func(ctx *tls_client.PostResponseContext) {
		mu.Lock()
		order = append(order, 4)
		mu.Unlock()
	})

	req, err := http.NewRequest(http.MethodGet, "https://httpbin.org/get", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Constructor hooks should run first, then runtime hooks
	assert.Equal(t, []int{1, 2, 3, 4}, order)
}
