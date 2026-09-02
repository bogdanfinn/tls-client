// Package main is ./example/concurrency with one change: every worker builds
// its own tls-client instead of sharing one. All workers still request the
// same host, so the two programs are directly comparable.
//
// What changes with a client per worker:
//
//   - Each client keeps its own connection pool, so the workers cannot share a
//     connection. The dial counter below shows one dial per worker rather than
//     one for the whole run, which means one TLS handshake per worker too.
//   - Each client keeps its own TLS session cache, so nothing is resumed
//     across workers.
//   - The fixed cost of a client is paid N times. The line after the clients
//     are built reports what that costs on the heap.
//
// Sharing one client is the better default for this workload. Build a client
// per worker when the workers need to differ - a proxy or profile each, say -
// not to get concurrency, which a single client already handles.
//
// Each worker decodes the response and then waits a couple of seconds, so the
// target sees a trickle of traffic rather than a load test.
//
// Run:
//
//	go run ./example/client_per_worker
//	go run ./example/client_per_worker -goroutines 8 -requests 10
//	go run ./example/client_per_worker -memprofile heap.out
//
// Inspect a heap profile with: go tool pprof heap.out
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

const targetURL = "https://de.afew-store.com/products.json"

// productsResponse is the part of the Shopify products.json payload this
// example cares about. Decoding into a struct rather than map[string]any keeps
// the allocation count down, which is what makes the heap profile readable.
type productsResponse struct {
	Products []struct {
		ID       int64    `json:"id"`
		Title    string   `json:"title"`
		Handle   string   `json:"handle"`
		Vendor   string   `json:"vendor"`
		Tags     []string `json:"tags"`
		Variants []struct {
			ID        int64  `json:"id"`
			Title     string `json:"title"`
			SKU       string `json:"sku"`
			Price     string `json:"price"`
			Available bool   `json:"available"`
		} `json:"variants"`
	} `json:"products"`
}

type stats struct {
	requests  atomic.Int64
	failures  atomic.Int64
	products  atomic.Int64
	variants  atomic.Int64
	available atomic.Int64
}

func main() {
	goroutines := flag.Int("goroutines", 8, "number of concurrent workers, each with its own client")
	requests := flag.Int("requests", 50, "requests per worker")
	minDelay := flag.Duration("min-delay", 2*time.Second, "minimum wait between a worker's requests")
	maxDelay := flag.Duration("max-delay", 3*time.Second, "maximum wait between a worker's requests")
	memStatsInterval := flag.Duration("memstats-interval", 5*time.Second, "how often to log heap stats, 0 to disable")
	memProfile := flag.String("memprofile", "", "write a heap profile to this path when done")
	flag.Parse()

	if *maxDelay < *minDelay {
		log.Fatal("-max-delay must not be smaller than -min-delay")
	}

	// Ctrl+C cancels the in-flight request and stops the workers, so a hung
	// request cannot hold up the shutdown for the full client timeout. The
	// connection pool is still torn down cleanly afterwards.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var counters stats

	// Taken before the clients exist, so the heap numbers include what N
	// clients cost rather than only what the requests cost.
	baseline := readMemStats()

	// dials counts every TCP connection opened across all clients. With a
	// client per worker this lands at one dial per worker, because a
	// connection cannot cross a client's pool.
	var dials atomic.Int64

	clients := make([]tls_client.HttpClient, 0, *goroutines)

	for worker := 1; worker <= *goroutines; worker++ {
		client, err := newClient(&dials)
		if err != nil {
			log.Fatalf("building the client for worker %d: %v", worker, err)
		}

		clients = append(clients, client)
	}

	afterClients := readMemStats()
	log.Printf("built %d clients: heap %s -> %s (%s for the clients)\n",
		*goroutines, humanBytes(baseline.HeapAlloc), humanBytes(afterClients.HeapAlloc),
		humanBytes(afterClients.HeapAlloc-baseline.HeapAlloc))

	log.Printf("start: %d workers, %d requests each, %s-%s between requests\n",
		*goroutines, *requests, *minDelay, *maxDelay)

	monitorDone := startMemStatsMonitor(ctx, *memStatsInterval, &counters)

	var wg sync.WaitGroup

	for worker := 1; worker <= *goroutines; worker++ {
		wg.Add(1)

		go func(worker int) {
			defer wg.Done()

			// Stagger the starts so the workers do not all fire at once.
			if !sleepCtx(ctx, time.Duration(worker-1)*250*time.Millisecond) {
				return
			}

			runWorker(ctx, clients[worker-1], worker, *requests, *minDelay, *maxDelay, &counters)
		}(worker)
	}

	wg.Wait()
	stop()
	<-monitorDone

	for _, client := range clients {
		client.CloseIdleConnections()
	}

	log.Printf("done: %d requests, %d failures, %d products, %d variants (%d available)\n",
		counters.requests.Load(), counters.failures.Load(),
		counters.products.Load(), counters.variants.Load(), counters.available.Load())
	log.Printf("TCP dials: %d across %d clients\n", dials.Load(), *goroutines)

	reportMemory(baseline)

	if *memProfile != "" {
		if err := writeHeapProfile(*memProfile); err != nil {
			log.Printf("writing the heap profile failed: %v\n", err)
		} else {
			log.Printf("heap profile written to %s - inspect it with: go tool pprof %s\n", *memProfile, *memProfile)
		}
	}
}

// newClient builds one client. MaxIdleConnsPerHost is 1 here, unlike in
// ./example/concurrency, because this pool only ever serves a single worker.
func newClient(dials *atomic.Int64) (tls_client.HttpClient, error) {
	dialer := &net.Dialer{Timeout: 30 * time.Second}

	return tls_client.NewHttpClient(tls_client.NewNoopLogger(), []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_152),
		// Chrome shuffles its TLS extensions per connection since version 107,
		// so a static order is a fingerprint of its own. Every client draws its
		// own order per handshake.
		tls_client.WithRandomTLSExtensionOrder(),
		tls_client.WithTransportOptions(&tls_client.TransportOptions{
			MaxIdleConnsPerHost: 1,
			IdleConnTimeout:     durationPtr(90 * time.Second),
		}),
		// Only here to count the dials. A real client does not need this.
		tls_client.WithDialContext(func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := dialer.DialContext(ctx, network, addr)
			if err == nil {
				dials.Add(1)
			}

			return conn, err
		}),
	}...)
}

func runWorker(ctx context.Context, client tls_client.HttpClient, worker, requests int, minDelay, maxDelay time.Duration, counters *stats) {
	for i := 1; i <= requests; i++ {
		if ctx.Err() != nil {
			return
		}

		products, variants, available, err := fetchAndDecode(ctx, client)

		// A request cancelled by the shutdown signal is not a failure and is
		// not counted, otherwise every Ctrl+C would show up in the summary.
		if err != nil && errors.Is(err, context.Canceled) {
			log.Printf("worker %d stopping: shutdown requested\n", worker)

			return
		}

		counters.requests.Add(1)

		if err != nil {
			counters.failures.Add(1)
			log.Printf("worker %d request %d/%d failed: %v\n", worker, i, requests, err)
		} else {
			counters.products.Add(int64(products))
			counters.variants.Add(int64(variants))
			counters.available.Add(int64(available))
			log.Printf("worker %d request %d/%d: %d products, %d variants, %d available\n",
				worker, i, requests, products, variants, available)
		}

		if i == requests {
			return
		}

		// Wait before the next request so the endpoint is not put under load.
		if !sleepCtx(ctx, jitter(minDelay, maxDelay)) {
			return
		}
	}
}

// fetchAndDecode performs one request and decodes the body. It streams the
// decode straight off the connection instead of reading the whole payload into
// a []byte first, which keeps the peak heap flat no matter how many workers
// run.
func fetchAndDecode(ctx context.Context, client tls_client.HttpClient) (products, variants, available int, err error) {
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return 0, 0, 0, err
	}

	req = req.WithContext(ctx)

	req.Header = http.Header{
		"accept":          {"application/json"},
		"accept-language": {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"user-agent":      {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-language",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, 0, err
	}

	// Draining the body before closing it lets the pooled connection be
	// reused. Skipping this is the most common reason a concurrent client
	// keeps opening new connections and the heap keeps growing.
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, 0, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var payload productsResponse

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, 0, 0, fmt.Errorf("decoding the response: %w", err)
	}

	for _, product := range payload.Products {
		variants += len(product.Variants)

		for _, variant := range product.Variants {
			if variant.Available {
				available++
			}
		}
	}

	return len(payload.Products), variants, available, nil
}

// jitter picks a wait between min and max so the workers do not fall into
// lockstep and hit the endpoint in bursts.
func jitter(min, max time.Duration) time.Duration {
	if max <= min {
		return min
	}

	return min + rand.N(max-min)
}

// sleepCtx waits for d and reports whether the wait completed. It returns
// false when the context was cancelled, which is the signal to stop working.
func sleepCtx(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return ctx.Err() == nil
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}

func startMemStatsMonitor(ctx context.Context, interval time.Duration, counters *stats) <-chan struct{} {
	done := make(chan struct{})

	if interval <= 0 {
		close(done)

		return done
	}

	go func() {
		defer close(done)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m := readMemStats()
				log.Printf("heap: %s in use, %d objects, %d GCs, %d goroutines, %d requests so far\n",
					humanBytes(m.HeapAlloc), m.HeapObjects, m.NumGC, runtime.NumGoroutine(), counters.requests.Load())
			case <-ctx.Done():
				return
			}
		}
	}()

	return done
}

func reportMemory(baseline runtime.MemStats) {
	// Collect first, so what is reported is live memory rather than garbage
	// the collector has not gotten to yet.
	runtime.GC()

	final := readMemStats()

	log.Printf("heap in use: %s at start, %s at end\n", humanBytes(baseline.HeapAlloc), humanBytes(final.HeapAlloc))
	log.Printf("cumulative allocations: %s in %d objects, %d GC cycles\n",
		humanBytes(final.TotalAlloc-baseline.TotalAlloc),
		final.Mallocs-baseline.Mallocs,
		final.NumGC-baseline.NumGC)
}

func writeHeapProfile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// The heap profile reports what is live at the moment it is taken, so
	// collect first.
	runtime.GC()

	return pprof.WriteHeapProfile(file)
}

func readMemStats() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return m
}

func humanBytes(b uint64) string {
	const unit = 1024

	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	value := float64(b)
	units := []string{"KiB", "MiB", "GiB"}

	for _, suffix := range units {
		value /= unit

		if value < unit {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
	}

	return fmt.Sprintf("%.1f TiB", value)
}

func durationPtr(d time.Duration) *time.Duration {
	return &d
}
