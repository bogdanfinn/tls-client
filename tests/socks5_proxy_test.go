package tests

// Integration tests for using a SOCKS5 proxy through the public API, including the
// combination with protocol racing that needs UDP ASSOCIATE for its HTTP/3 leg.
//
// The proxy comes from the SOCKS_5_PROXY environment variable (see mise.toml, value in
// the uncommitted mise.local.toml) and has to support UDP ASSOCIATE, so these tests skip
// when it is unset.
//
// Run them with:
//
//	go test ./tests -run Socks5Proxy -v
//
// Note that these tests deliberately do not assert that a request was served over
// HTTP/3. With racing enabled the protocol is whichever leg wins, and the HTTP/3 leg is
// the slower one when it has to take an extra proxy hop, so HTTP/2 usually wins. The
// public API has no option to force HTTP/3, so the deterministic HTTP/3 assertion lives
// in the root package, which can build the HTTP/3 transport directly.

import (
	"encoding/json"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

const socks5ProxyEnvVar = "SOCKS_5_PROXY"

// socks5ProxyFromEnv returns the configured proxy URL, or skips the test when there is
// none. A bare host:port is accepted as well, since the scheme is only what selects the
// SOCKS5 code path.
func socks5ProxyFromEnv(t *testing.T) string {
	t.Helper()

	proxyURL := strings.TrimSpace(os.Getenv(socks5ProxyEnvVar))
	if proxyURL == "" {
		t.Skipf("%s is not set, skipping the SOCKS5 proxy test", socks5ProxyEnvVar)
	}

	if !strings.Contains(proxyURL, "://") {
		proxyURL = "socks5://" + proxyURL
	}

	if !strings.HasPrefix(proxyURL, "socks5://") && !strings.HasPrefix(proxyURL, "socks5h://") {
		t.Fatalf("%s has to be a socks5:// or socks5h:// URL, got %q", socks5ProxyEnvVar, proxyURL)
	}

	return proxyURL
}

// proxyHost returns the host of the proxy URL without credentials or port, so that it can
// be logged and resolved without leaking the password into the test output.
func proxyHost(t *testing.T, proxyURL string) string {
	t.Helper()

	parsed, err := url.Parse(proxyURL)
	if err != nil {
		t.Fatalf("failed to parse %s: %v", socks5ProxyEnvVar, err)
	}

	return parsed.Hostname()
}

// fetchEgress reports the address the endpoint saw and the protocol it answered over.
func fetchEgress(t *testing.T, client tls_client.HttpClient) (string, string) {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, peetApiEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultHeader

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request to %s failed: %v", peetApiEndpoint, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read the response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d from %s: %s", resp.StatusCode, peetApiEndpoint, string(body))
	}

	var parsed TlsApiResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("failed to parse the response %q: %v", string(body), err)
	}

	address := parsed.IP
	if host, _, splitErr := net.SplitHostPort(address); splitErr == nil {
		address = host
	}

	return address, parsed.HTTPVersion
}

func newSocks5TestClient(t *testing.T, options ...tls_client.HttpClientOption) tls_client.HttpClient {
	t.Helper()

	base := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_144),
		tls_client.WithTimeoutSeconds(30),
	}

	client, err := tls_client.NewHttpClient(nil, append(base, options...)...)
	if err != nil {
		t.Fatalf("failed to create the client: %v", err)
	}

	return client
}

// TestSocks5Proxy_HappyPath is the plain happy path: a real request through the proxy from
// SOCKS_5_PROXY, asserting that the endpoint saw the proxy's address instead of this
// machine's own one.
//
// It also reports whether the egress address is one the proxy hostname resolves to. That
// is a log line rather than an assertion, because a proxy is free to exit through an
// upstream or a rotating pool, in which case the egress address legitimately differs from
// the address the SOCKS5 connection was made to. What has to hold in every case is that
// the address is not the unproxied one.
func TestSocks5Proxy_HappyPath(t *testing.T) {
	proxyURL := socks5ProxyFromEnv(t)
	host := proxyHost(t, proxyURL)

	directIP, directProtocol := fetchEgress(t, newSocks5TestClient(t))
	t.Logf("without a proxy the request came from %s over %s", directIP, directProtocol)

	proxiedIP, proxiedProtocol := fetchEgress(t, newSocks5TestClient(t, tls_client.WithProxyUrl(proxyURL)))
	t.Logf("through the SOCKS5 proxy %s the request came from %s over %s", host, proxiedIP, proxiedProtocol)

	if net.ParseIP(proxiedIP) == nil {
		t.Fatalf("the endpoint reported %q instead of an address for the proxied request", proxiedIP)
	}

	if proxiedIP == directIP {
		t.Fatalf("the proxied request left from %s, the same address as an unproxied request, so it bypassed the proxy", proxiedIP)
	}

	logEgressAgainstProxyAddresses(t, host, proxiedIP)
}

// logEgressAgainstProxyAddresses resolves the proxy hostname and reports whether the
// egress address is one of its own addresses.
func logEgressAgainstProxyAddresses(t *testing.T, host string, egressIP string) {
	t.Helper()

	proxyIPs, err := net.LookupHost(host)
	if err != nil {
		t.Logf("could not resolve the proxy host %s to compare it with the egress address: %v", host, err)

		return
	}

	for _, proxyIP := range proxyIPs {
		if proxyIP == egressIP {
			t.Logf("the egress address %s is the proxy's own address", egressIP)

			return
		}
	}

	t.Logf("the egress address %s is not one of the proxy's own addresses %v, which is expected for a proxy that exits through an upstream or a rotating pool", egressIP, proxyIPs)
}

// TestSocks5Proxy_ProtocolRacingUsesProxy is the combination that config validation now
// permits: protocol racing together with a SOCKS5 proxy. Whichever leg wins the race, the
// request has to leave through the proxy.
func TestSocks5Proxy_ProtocolRacingUsesProxy(t *testing.T) {
	proxyURL := socks5ProxyFromEnv(t)

	directIP, directProtocol := fetchEgress(t, newSocks5TestClient(t))
	t.Logf("unproxied request came from %s over %s", directIP, directProtocol)

	proxiedClient := newSocks5TestClient(t,
		tls_client.WithProtocolRacing(),
		tls_client.WithProxyUrl(proxyURL),
	)

	proxiedIP, proxiedProtocol := fetchEgress(t, proxiedClient)
	t.Logf("proxied request came from %s over %s", proxiedIP, proxiedProtocol)

	if proxiedIP == "" {
		t.Fatal("the endpoint did not report an address for the proxied request")
	}

	if proxiedIP == directIP {
		t.Fatalf("the proxied request left from %s, the same address as an unproxied request, so it bypassed the proxy", proxiedIP)
	}
}

// TestSocks5Proxy_RacingSurvivesRepeatedRequests covers the racer's protocol cache. The
// winning protocol is cached per host, and for HTTP/3 the racer builds a second transport
// from the cache path, which is a separate call site for the SOCKS5 dialer.
func TestSocks5Proxy_RacingSurvivesRepeatedRequests(t *testing.T) {
	proxyURL := socks5ProxyFromEnv(t)

	client := newSocks5TestClient(t,
		tls_client.WithProtocolRacing(),
		tls_client.WithProxyUrl(proxyURL),
	)

	var firstIP string

	for round := 0; round < 3; round++ {
		ip, protocol := fetchEgress(t, client)
		t.Logf("round %d came from %s over %s", round, ip, protocol)

		if ip == "" {
			t.Fatalf("round %d: the endpoint did not report an address", round)
		}

		if round == 0 {
			firstIP = ip

			continue
		}

		if ip != firstIP {
			t.Logf("round %d left from %s instead of %s, which is expected for a rotating proxy", round, ip, firstIP)
		}
	}
}

// TestSocks5Proxy_SetProxyAtRuntime covers switching a racing client onto the SOCKS5
// proxy after it was built, which is the path that skipped validation before.
func TestSocks5Proxy_SetProxyAtRuntime(t *testing.T) {
	proxyURL := socks5ProxyFromEnv(t)

	client := newSocks5TestClient(t, tls_client.WithProtocolRacing())

	directIP, _ := fetchEgress(t, client)
	t.Logf("before SetProxy the request came from %s", directIP)

	if err := client.SetProxy(proxyURL); err != nil {
		t.Fatalf("SetProxy with a SOCKS5 proxy failed: %v", err)
	}

	proxiedIP, proxiedProtocol := fetchEgress(t, client)
	t.Logf("after SetProxy the request came from %s over %s", proxiedIP, proxiedProtocol)

	if proxiedIP == directIP {
		t.Fatalf("the request still left from %s after SetProxy, so the proxy was not applied", proxiedIP)
	}
}
