// Package main verifies that a client reuses one TCP connection for
// subsequent requests to the same host, instead of dialing and handshaking
// again every time. That is the behaviour questioned in
// https://github.com/bogdanfinn/tls-client/issues/112.
//
// The check does not rely on timings. It counts the actual TCP dials through a
// wrapping dialer and cross-checks that against httptrace, which reports per
// request whether the connection came out of the idle pool. The local port of
// every connection is recorded too, so "same connection" is shown rather than
// inferred.
//
// The connection is reused, including with the exact options from the issue.
// Against the default target every scenario here dials once for all requests,
// and the requests after the first drop from roughly 400ms to 100ms, which is
// the TLS handshake no longer happening.
//
// Why the issue reports otherwise: its target is https://tls.peet.ws, and that
// host does not keep the connection alive. It sends a stray CRLF pair on the
// idle connection after a response, so the client has to discard it. Go's own
// net/http behaves identically against that host and logs "Unsolicited
// response received on idle HTTP channel", so it is the server, not the
// client. Pass -url https://tls.peet.ws/api/tls to see it.
//
// Run:
//
//	go run ./example/connection_reuse
//	go run ./example/connection_reuse -scenario h1 -requests 5
//	go run ./example/connection_reuse -url https://tls.peet.ws/api/tls
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/httptrace"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

type scenario struct {
	name        string
	description string
	// options gets the counting dialer appended, so a scenario only lists
	// what it wants to prove.
	options []tls_client.HttpClientOption
	// closeBody false reproduces the code sample in the issue, which reads
	// the body but never closes it.
	closeBody bool
}

type requestResult struct {
	index     int
	proto     string
	status    int
	reused    bool
	wasIdle   bool
	idleTime  time.Duration
	localAddr string
	duration  time.Duration
	err       error
}

func scenarios() []scenario {
	return []scenario{
		{
			name:        "h2",
			description: "default client, ALPN negotiates HTTP/2, body drained and closed",
			options: []tls_client.HttpClientOption{
				tls_client.WithClientProfile(profiles.Chrome_152),
			},
			closeBody: true,
		},
		{
			name:        "h1",
			description: "same, but forced to HTTP/1.1",
			options: []tls_client.HttpClientOption{
				tls_client.WithClientProfile(profiles.Chrome_152),
				tls_client.WithForceHttp1(),
			},
			closeBody: true,
		},
		{
			name:        "issue112",
			description: "the exact options from issue #112, body drained and closed",
			options: []tls_client.HttpClientOption{
				tls_client.WithRandomTLSExtensionOrder(),
				tls_client.WithTransportOptions(&tls_client.TransportOptions{
					MaxConnsPerHost:    1,
					DisableCompression: true,
				}),
			},
			closeBody: true,
		},
		{
			name:        "issue112-unclosed-body",
			description: "the options AND the body handling from issue #112: the body is read but never closed",
			options: []tls_client.HttpClientOption{
				tls_client.WithRandomTLSExtensionOrder(),
				tls_client.WithTransportOptions(&tls_client.TransportOptions{
					MaxConnsPerHost:    1,
					DisableCompression: true,
				}),
			},
			closeBody: false,
		},
	}
}

func main() {
	target := flag.String("url", "https://httpbin.org/get", "url to request repeatedly")
	requests := flag.Int("requests", 3, "sequential requests per scenario")
	only := flag.String("scenario", "all", "scenario to run: all, or one of h2, h1, issue112, issue112-unclosed-body")
	flag.Parse()

	if *requests < 2 {
		log.Fatal("-requests has to be at least 2, otherwise there is nothing that could be reused")
	}

	all := scenarios()
	selected := make([]scenario, 0, len(all))

	for _, s := range all {
		if *only == "all" || *only == s.name {
			selected = append(selected, s)
		}
	}

	if len(selected) == 0 {
		log.Fatalf("unknown scenario %q", *only)
	}

	log.Printf("verifying connection reuse against %s, %d requests per scenario\n", *target, *requests)

	failed := false

	for _, s := range selected {
		if !runScenario(s, *target, *requests) {
			failed = true
		}
	}

	if failed {
		os.Exit(1)
	}
}

func runScenario(s scenario, target string, requests int) bool {
	fmt.Printf("\n=== %s ===\n%s\n\n", s.name, s.description)

	// dials counts every TCP connection the client actually opens. This is
	// the number that settles the question: one dial for many requests means
	// the connection was reused.
	var dials atomic.Int64

	dialer := &net.Dialer{Timeout: 30 * time.Second}
	options := append([]tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithDialContext(func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := dialer.DialContext(ctx, network, addr)
			if err == nil {
				dials.Add(1)
			}

			return conn, err
		}),
	}, s.options...)

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		fmt.Printf("  building the client failed: %v\n", err)

		return false
	}

	results := make([]requestResult, 0, requests)

	for i := 1; i <= requests; i++ {
		results = append(results, doRequest(client, target, i, s.closeBody))
	}

	client.CloseIdleConnections()

	return report(s, results, dials.Load(), requests)
}

func doRequest(client tls_client.HttpClient, target string, index int, closeBody bool) requestResult {
	result := requestResult{index: index}

	// httptrace reports, per request, whether the transport took the
	// connection out of its idle pool rather than opening a new one.
	trace := &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {
			result.reused = info.Reused
			result.wasIdle = info.WasIdle
			result.idleTime = info.IdleTime

			if info.Conn != nil {
				result.localAddr = info.Conn.LocalAddr().String()
			}
		},
	}

	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		result.err = err

		return result
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	start := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		result.err = err
		result.duration = time.Since(start)

		return result
	}

	// Reading to EOF is what hands the connection back to the pool. The
	// unclosed-body scenario deliberately skips the Close that should follow.
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		result.err = err
	}

	if closeBody {
		_ = resp.Body.Close()
	}

	result.duration = time.Since(start)
	result.proto = resp.Proto
	result.status = resp.StatusCode

	return result
}

func report(s scenario, results []requestResult, dials int64, requests int) bool {
	fmt.Printf("  %-4s %-9s %-7s %-7s %-8s %-22s %s\n", "req", "proto", "status", "reused", "wasIdle", "local address", "duration")

	reused := 0
	errors := 0
	addresses := map[string]int{}

	for _, r := range results {
		if r.err != nil {
			errors++
			fmt.Printf("  %-4d failed: %v\n", r.index, r.err)

			continue
		}

		if r.reused {
			reused++
		}

		if r.localAddr != "" {
			addresses[r.localAddr]++
		}

		fmt.Printf("  %-4d %-9s %-7d %-7t %-8t %-22s %s\n",
			r.index, r.proto, r.status, r.reused, r.wasIdle, orDash(r.localAddr), r.duration.Round(time.Millisecond))
	}

	fmt.Printf("\n  TCP dials: %d | connections reused: %d/%d | distinct local addresses: %d\n",
		dials, reused, requests-1, len(addresses))

	if len(addresses) > 0 {
		fmt.Printf("  local addresses: %s\n", strings.Join(sortedKeys(addresses), ", "))
	}

	if errors > 0 {
		fmt.Printf("  VERDICT: inconclusive, %d of %d requests failed\n", errors, requests)

		return false
	}

	// One dial and every request after the first taken from the pool is what
	// reuse looks like.
	switch {
	case dials == 1 && reused == requests-1 && len(addresses) == 1:
		fmt.Printf("  VERDICT: connection is reused - %d requests over a single TCP connection\n", requests)

		return true
	case dials >= int64(requests):
		fmt.Printf("  VERDICT: no reuse - %d dials for %d requests\n", dials, requests)
		fmt.Printf("           A server can force this by not keeping the connection alive, which is\n")
		fmt.Printf("           what tls.peet.ws does. Before suspecting the client, try another host\n")
		fmt.Printf("           or check the same URL with net/http.\n")

		return false
	default:
		fmt.Printf("  VERDICT: partial reuse - %d dials and %d reused connections for %d requests\n", dials, reused, requests)

		return false
	}
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}

	return s
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}
