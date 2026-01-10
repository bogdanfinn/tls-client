package tls_client

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/http2"
	"github.com/bogdanfinn/quic-go-utls/http3"
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
}

func newProtocolRacer(
	clientSessionCache tls.ClientSessionCache,
	insecureSkipVerify bool,
	serverNameOverwrite string,
	transportOptions *TransportOptions,
	settings map[http2.SettingID]uint32,
	cachedTransports map[string]http.RoundTripper,
	cachedTransportsLck *sync.Mutex,
	certificatePinner CertificatePinner,
	badPinHandlerFunc BadPinHandlerFunc,
	bandwidthTracker bandwidth.BandwidthTracker,
) *protocolRacer {
	return &protocolRacer{
		protocolCache:       make(map[string]string),
		clientSessionCache:  clientSessionCache,
		insecureSkipVerify:  insecureSkipVerify,
		serverNameOverwrite: serverNameOverwrite,
		transportOptions:    transportOptions,
		settings:            settings,
		cachedTransports:    cachedTransports,
		cachedTransportsLck: cachedTransportsLck,
		certificatePinner:   certificatePinner,
		badPinHandlerFunc:   badPinHandlerFunc,
		bandwidthTracker:    bandwidthTracker,
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

	resp, err := transport.RoundTrip(req)
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
		return pr.buildHTTP3Transport()
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
	h3Transport, err := pr.buildHTTP3Transport()
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

	resp, err := h2Transport.RoundTrip(req)
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
		pr.cachedTransportsLck.Lock()
		h3Transport, _ := pr.buildHTTP3Transport()
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

type racingResult struct {
	protocol string
	response *http.Response
	err      error
}

func (pr *protocolRacer) buildHTTP3Transport() (http.RoundTripper, error) {
	utlsConfig := &tls.Config{
		ClientSessionCache: pr.clientSessionCache,
		InsecureSkipVerify: pr.insecureSkipVerify,
		OmitEmptyPsk:       true,
	}
	if pr.transportOptions != nil {
		utlsConfig.RootCAs = pr.transportOptions.RootCAs
	}

	if pr.serverNameOverwrite != "" {
		utlsConfig.ServerName = pr.serverNameOverwrite
	}

	t3 := &http3.Transport{
		TLSClientConfig: utlsConfig,
	}

	// Note: Do NOT set AdditionalSettings unless absolutely necessary.
	// HTTP/3 uses different setting IDs than HTTP/2:
	// - HTTP/3: 0x6 (MAX_FIELD_SECTION_SIZE), 0x8 (EXTENDED_CONNECT), 0x33 (DATAGRAM)
	// - HTTP/2: 0x1 (HEADER_TABLE_SIZE), 0x3 (MAX_CONCURRENT_STREAMS), etc.
	// Sending HTTP/2 settings to HTTP/3 will cause H3_SETTINGS_ERROR.
	// The quic-go-utls library has good defaults - use them.

	if pr.transportOptions != nil {
		t3.DisableCompression = pr.transportOptions.DisableCompression

		maxResponseHeaderBytes, convErr := Int64ToInt(pr.transportOptions.MaxResponseHeaderBytes)
		if convErr != nil {
			return nil, fmt.Errorf("error converting MaxResponseHeaderBytes to int: %w", convErr)
		}

		t3.MaxResponseHeaderBytes = maxResponseHeaderBytes
	}

	return t3, nil
}
