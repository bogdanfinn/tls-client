// Package main shows pre-request and post-response hooks: one place to touch
// every request a client sends and every response it receives, without
// wrapping each call site.
//
// Run: go run ./example/hooks
package main

import (
	"fmt"
	"log"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

func main() {
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_150),

		// A pre hook sees the request before it goes out and can modify it.
		tls_client.WithPreHook(func(req *http.Request) error {
			req.Header.Set("x-request-source", "tls-client-example")
			log.Printf("pre hook: %s %s\n", req.Method, req.URL)

			return nil
		}),

		// A post hook sees the request, the response and the transport error
		// if there was one.
		tls_client.WithPostHook(func(ctx *tls_client.PostResponseContext) error {
			if ctx.Error != nil {
				log.Printf("post hook: %s failed: %v\n", ctx.Request.URL, ctx.Error)

				return nil
			}

			log.Printf("post hook: %s returned %d\n", ctx.Request.URL, ctx.Response.StatusCode)

			return nil
		}),
	}...)
	if err != nil {
		log.Fatal(err)
	}

	// Hooks can also be added after the client was built. They run in the
	// order they were registered.
	client.AddPreRequestHook(func(req *http.Request) error {
		// An error normally aborts the remaining hooks and the request.
		// Wrapping ErrContinueHooks logs it and carries on instead.
		return fmt.Errorf("non fatal problem: %w", tls_client.ErrContinueHooks)
	})

	req, err := http.NewRequest(http.MethodGet, "https://tls.peet.ws/api/all", nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	log.Printf("status: %d\n", resp.StatusCode)

	// ResetPreHooks / ResetPostHooks drop every registered hook.
	client.ResetPreHooks()
	client.ResetPostHooks()
	log.Println("hooks removed")
}
