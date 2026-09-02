package tls_client

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/http2"
	"github.com/bogdanfinn/tls-client/bandwidth"
	tls "github.com/bogdanfinn/utls"
)

type protocolRacer struct {
	protocolCache   map[string]string
	protocolCacheMu sync.RWMutex

	clientSessionCache  tls.ClientSessionCache
	insecureSkipVerify  bool
	serverNameOverwrite string
	transportOptions    *TransportOptions
	settings            map[http2.SettingID]uint32
	cachedTransports    map[string]http.RoundTripper
	cachedTransportsLck *sync.Mutex
	certificatePinner   CertificatePinner
	badPinHandlerFunc   BadPinHandlerFunc
	bandwidthTracker    bandwidth.BandwidthTracker

	// dropTransport forgets the transport cached for an address so the next
	// attempt builds one from a fresh handshake. It is the round tripper's own
	// eviction, handed over because a raced request returns from RoundTrip
	// before the branch that would otherwise do it.
	dropTransport func(addr string, stale http.RoundTripper)

	// HTTP/3 specific settings
	http3Settings          map[uint64]uint64
	http3SettingsOrder     []uint64
	http3PriorityParam     uint32
	http3PseudoHeaderOrder []string
	http3SendGreaseFrames  bool

	proxyURL string
}

func newProtocolRacer(
	clientSessionCache tls.ClientSessionCache,
	insecureSkipVerify bool,
	serverNameOverwrite string,
	transportOptions *TransportOptions,
	settings map[http2.SettingID]uint32,
	cachedTransports map[string]http.RoundTripper,
	cachedTransportsLck *sync.Mutex,
	dropTransport func(addr string, stale http.RoundTripper),
	certificatePinner CertificatePinner,
	badPinHandlerFunc BadPinHandlerFunc,
	bandwidthTracker bandwidth.BandwidthTracker,
	http3Settings map[uint64]uint64,
	http3SettingsOrder []uint64,
	http3PriorityParam uint32,
	http3PseudoHeaderOrder []string,
	http3SendGreaseFrames bool,
	proxyURL string,
) *protocolRacer {
	return &protocolRacer{
		protocolCache:          make(map[string]string),
		clientSessionCache:     clientSessionCache,
		insecureSkipVerify:     insecureSkipVerify,
		serverNameOverwrite:    serverNameOverwrite,
		transportOptions:       transportOptions,
		settings:               settings,
		cachedTransports:       cachedTransports,
		cachedTransportsLck:    cachedTransportsLck,
		dropTransport:          dropTransport,
		certificatePinner:      certificatePinner,
		badPinHandlerFunc:      badPinHandlerFunc,
		bandwidthTracker:       bandwidthTracker,
		http3Settings:          http3Settings,
		http3SettingsOrder:     http3SettingsOrder,
		http3PriorityParam:     http3PriorityParam,
		http3PseudoHeaderOrder: http3PseudoHeaderOrder,
		http3SendGreaseFrames:  http3SendGreaseFrames,
		proxyURL:               proxyURL,
	}
}

// race races HTTP/3 and HTTP/2 connections and uses whichever responds first.
// Similar to Chrome's "Happy Eyeballs" approach.
func (pr *protocolRacer) race(req *http.Request, addr string, getTransportFunc func(*http.Request, string) error) (*http.Response, error) {
	// Try cached protocol first if available
	if resp, shouldRace := pr.tryUseCachedProtocol(req, addr, getTransportFunc); !shouldRace {
		return resp, nil
	}

	// No cached protocol or it failed - start racing
	return pr.startRace(req, addr, getTransportFunc)
}

func (pr *protocolRacer) tryUseCachedProtocol(req *http.Request, addr string, getTransportFunc func(*http.Request, string) error) (*http.Response, bool) {
	pr.protocolCacheMu.RLock()
	cachedProtocol, found := pr.protocolCache[addr]
	pr.protocolCacheMu.RUnlock()

	if !found {
		return nil, true // No cache, proceed to racing
	}

	transport, err := pr.getOrCreateTransport(cachedProtocol, addr, req, getTransportFunc)
	if err != nil {
		pr.handleCachedProtocolError(err, addr, req)
		return nil, true // Cached protocol failed, proceed to racing
	}

	resp, err := pr.roundTrip(transport, req, addr)
	if err == nil {
		return resp, false // Success!
	}

	pr.clearProtocolCache(addr)
	return nil, true
}

func (pr *protocolRacer) getOrCreateTransport(protocol, addr string, req *http.Request, getTransportFunc func(*http.Request, string) error) (http.RoundTripper, error) {
	transportKey := pr.getTransportKey(protocol, addr)

	pr.cachedTransportsLck.Lock()
	defer pr.cachedTransportsLck.Unlock()

	if transport, exists := pr.cachedTransports[transportKey]; exists {
		return transport, nil
	}

	transport, err := pr.createTransportForProtocol(protocol, addr, req, getTransportFunc)
	if err != nil {
		return nil, err
	}

	pr.cachedTransports[transportKey] = transport
	return transport, nil
}

func (pr *protocolRacer) createTransportForProtocol(protocol, addr string, req *http.Request, getTransportFunc func(*http.Request, string) error) (http.RoundTripper, error) {
	if protocol == "h3" {
		return buildHTTP3Transport(pr.getHTTP3Config())
	}

	// For HTTP/2, use the standard transport creation
	transportKey := pr.getTransportKey(protocol, addr)
	if err := getTransportFunc(req, transportKey); err != nil {
		return nil, err
	}

	return pr.cachedTransports[transportKey], nil
}

func (pr *protocolRacer) startRace(req *http.Request, addr string, getTransportFunc func(*http.Request, string) error) (*http.Response, error) {
	resultCh := make(chan racingResult, 2)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go pr.attemptHTTP3(req, resultCh)
	go pr.attemptHTTP2(req, addr, getTransportFunc, resultCh)

	return pr.waitForRaceWinner(ctx, addr, resultCh, cancel)
}

func (pr *protocolRacer) attemptHTTP3(req *http.Request, resultCh chan<- racingResult) {
	h3Transport, err := buildHTTP3Transport(pr.getHTTP3Config())
	if err != nil {
		resultCh <- racingResult{protocol: "h3", err: fmt.Errorf("failed to build HTTP/3 transport: %w", err)}
		return
	}

	resp, err := h3Transport.RoundTrip(req)
	if err != nil {
		resultCh <- racingResult{protocol: "h3", err: fmt.Errorf("HTTP/3 request failed: %w", err)}
	} else {
		resultCh <- racingResult{protocol: "h3", response: resp}
	}
}

func (pr *protocolRacer) attemptHTTP2(req *http.Request, addr string, getTransportFunc func(*http.Request, string) error, resultCh chan<- racingResult) {
	// Chrome-like 300ms delay before starting HTTP/2
	// https://groups.google.com/a/chromium.org/g/proto-quic/c/igD7dLSct24
	time.Sleep(300 * time.Millisecond)

	pr.cachedTransportsLck.Lock()
	if _, ok := pr.cachedTransports[addr]; !ok {
		if err := getTransportFunc(req, addr); err != nil {
			pr.cachedTransportsLck.Unlock()
			resultCh <- racingResult{protocol: "h2", err: err}
			return
		}
	}
	h2Transport := pr.cachedTransports[addr]
	pr.cachedTransportsLck.Unlock()

	resp, err := pr.roundTrip(h2Transport, req, addr)
	resultCh <- racingResult{protocol: "h2", response: resp, err: err}
}

func (pr *protocolRacer) waitForRaceWinner(ctx context.Context, addr string, resultCh <-chan racingResult, cancel context.CancelFunc) (*http.Response, error) {
	var lastErr error

	for i := 0; i < 2; i++ {
		select {
		case result := <-resultCh:
			if result.err == nil && result.response != nil {
				pr.cacheWinningProtocol(addr, result.protocol)
				cancel()
				return result.response, nil
			}
			lastErr = result.err

		case <-ctx.Done():
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, ctx.Err()
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("http3 racing: both protocols failed to connect")
}

// roundTrip sends the request over transport and, when the dial underneath
// reports that the server has moved to a protocol this transport cannot speak,
// forgets the transport so the next attempt builds one that fits.
//
// Clearing the protocol cache alone does not recover from that: the race it
// falls back to reaches for the same cached transport and fails on the same
// mismatch, for every request from then on.
//
// addr is the dial address rather than the transport's cache key, because only
// the transport stored under that key dials through the round tripper. The
// HTTP/3 transport brings its own dialer and never reports this.
func (pr *protocolRacer) roundTrip(transport http.RoundTripper, req *http.Request, addr string) (*http.Response, error) {
	resp, err := transport.RoundTrip(req)
	if err != nil && errors.Is(err, errProtocolChanged) && pr.dropTransport != nil {
		pr.dropTransport(addr, transport)
	}

	return resp, err
}

func (pr *protocolRacer) getTransportKey(protocol, addr string) string {
	if protocol == "h3" {
		return addr + ":h3"
	}
	return addr
}

func (pr *protocolRacer) clearProtocolCache(addr string) {
	pr.protocolCacheMu.Lock()
	delete(pr.protocolCache, addr)
	pr.protocolCacheMu.Unlock()
}

func (pr *protocolRacer) cacheWinningProtocol(addr, protocol string) {
	pr.protocolCacheMu.Lock()
	pr.protocolCache[addr] = protocol
	pr.protocolCacheMu.Unlock()

	if protocol == "h3" {
		h3Transport, err := buildHTTP3Transport(pr.getHTTP3Config())
		if err != nil {
			// Never cache a nil transport, the next request would panic on it.
			return
		}

		pr.cachedTransportsLck.Lock()
		pr.cachedTransports[addr+":h3"] = h3Transport
		pr.cachedTransportsLck.Unlock()
	}
}

func (pr *protocolRacer) handleCachedProtocolError(err error, addr string, req *http.Request) {
	if errors.Is(err, ErrBadPinDetected) && pr.badPinHandlerFunc != nil {
		pr.badPinHandlerFunc(req)
	}
	pr.clearProtocolCache(addr)
}

func (pr *protocolRacer) getHTTP3Config() *http3Config {
	return &http3Config{
		clientSessionCache:     pr.clientSessionCache,
		insecureSkipVerify:     pr.insecureSkipVerify,
		serverNameOverwrite:    pr.serverNameOverwrite,
		transportOptions:       pr.transportOptions,
		http3Settings:          pr.http3Settings,
		http3SettingsOrder:     pr.http3SettingsOrder,
		http3PriorityParam:     pr.http3PriorityParam,
		http3PseudoHeaderOrder: pr.http3PseudoHeaderOrder,
		http3SendGreaseFrames:  pr.http3SendGreaseFrames,
		proxyURL:               pr.proxyURL,
	}
}

type racingResult struct {
	protocol string
	response *http.Response
	err      error
}
