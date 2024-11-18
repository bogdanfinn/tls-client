package tls_client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/http2"
	"github.com/bogdanfinn/tls-client/bandwidth"
	"github.com/bogdanfinn/tls-client/profiles"
	tls "github.com/bogdanfinn/utls"
	"golang.org/x/net/proxy"
)

const defaultIdleConnectionTimeout = 90 * time.Second

var errProtocolNegotiated = errors.New("protocol negotiated")

type roundTripper struct {
	clientHelloId     tls.ClientHelloID
	certificatePinner CertificatePinner

	dialer proxy.ContextDialer

	bandwidthTracker bandwidth.BandwidthTracker

	clientSessionCache tls.ClientSessionCache

	badPinHandlerFunc BadPinHandlerFunc
	cachedConnections map[string]net.Conn
	cachedTransports  map[string]http.RoundTripper

	headerPriority      *http2.PriorityParam
	settings            map[http2.SettingID]uint32
	transportOptions    *TransportOptions
	serverNameOverwrite string
	priorities          []http2.Priority
	pseudoHeaderOrder   []string
	settingsOrder       []http2.SettingID
	sync.Mutex

	cachedTransportsLck sync.Mutex
	connectionFlow      uint32

	forceHttp1 bool

	insecureSkipVerify          bool
	withRandomTlsExtensionOrder bool
	disableIPV6                 bool
	disableIPV4                 bool
}

func (rt *roundTripper) CloseIdleConnections() {
	rt.cachedTransportsLck.Lock()
	defer rt.cachedTransportsLck.Unlock()

	type closeIdler interface {
		CloseIdleConnections()
	}

	for _, transport := range rt.cachedTransports {
		if tr, ok := transport.(closeIdler); ok {
			tr.CloseIdleConnections()
		}
	}
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	addr := rt.getDialTLSAddr(req)

	rt.cachedTransportsLck.Lock()

	if _, ok := rt.cachedTransports[addr]; !ok {
		if err := rt.getTransport(req, addr); err != nil {
			rt.cachedTransportsLck.Unlock()

			if errors.Is(err, ErrBadPinDetected) && rt.badPinHandlerFunc != nil {
				rt.badPinHandlerFunc(req)
			}

			return nil, err
		}
	}

	t := rt.cachedTransports[addr]
	rt.cachedTransportsLck.Unlock()

	return t.RoundTrip(req)
}

func (rt *roundTripper) getTransport(req *http.Request, addr string) error {
	switch strings.ToLower(req.URL.Scheme) {
	case "http":
		rt.cachedTransports[addr] = rt.buildHttp1Transport()
		return nil
	case "https":
	default:
		return fmt.Errorf("invalid URL scheme: [%v]", req.URL.Scheme)
	}

	_, err := rt.dialTLS(req.Context(), "tcp", addr)
	switch err {
	case errProtocolNegotiated:
	case nil:
		// Should never happen.
		panic("dialTLS returned no error when determining cachedTransports")
	default:
		return err
	}

	return nil
}

func (rt *roundTripper) dialTLS(ctx context.Context, network, addr string) (net.Conn, error) {
	rt.Lock()
	defer rt.Unlock()

	// If we have the connection from when we determined the HTTPS
	// cachedTransports to use, return that.
	if conn := rt.cachedConnections[addr]; conn != nil {
		delete(rt.cachedConnections, addr)

		return conn, nil
	}

	if network == "tcp" && rt.disableIPV6 {
		network = "tcp4"
	}

	if network == "tcp" && rt.disableIPV4 {
		network = "tcp6"
	}

	rawConn, err := rt.dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	var host string
	if host, _, err = net.SplitHostPort(addr); err != nil {
		host = addr
	}

	if rt.serverNameOverwrite != "" {
		host = rt.serverNameOverwrite
	}

	tlsConfig := &tls.Config{ClientSessionCache: rt.clientSessionCache, ServerName: host, InsecureSkipVerify: rt.insecureSkipVerify, OmitEmptyPsk: true}
	if rt.transportOptions != nil {
		tlsConfig.RootCAs = rt.transportOptions.RootCAs
		tlsConfig.KeyLogWriter = rt.transportOptions.KeyLogWriter
	}

	rawConn = rt.bandwidthTracker.TrackConnection(ctx, rawConn)

	conn := tls.UClient(rawConn, tlsConfig, rt.clientHelloId, rt.withRandomTlsExtensionOrder, rt.forceHttp1)
	if err = conn.HandshakeContext(ctx); err != nil {
		_ = conn.Close()

		return nil, err
	}

	err = rt.certificatePinner.Pin(conn, host)

	if err != nil {
		return nil, err
	}

	if rt.cachedTransports[addr] != nil {
		return conn, nil
	}

	// No http.Transport constructed yet, create one based on the results
	// of ALPN if no http1 is enforced.

	switch conn.ConnectionState().NegotiatedProtocol {
	case http2.NextProtoTLS:
		utlsConfig := &tls.Config{ClientSessionCache: rt.clientSessionCache, InsecureSkipVerify: rt.insecureSkipVerify, OmitEmptyPsk: true}
		if rt.transportOptions != nil {
			utlsConfig.RootCAs = rt.transportOptions.RootCAs
		}

		if rt.serverNameOverwrite != "" {
			utlsConfig.ServerName = rt.serverNameOverwrite
		}

		idleConnectionTimeout := defaultIdleConnectionTimeout

		if rt.transportOptions != nil && rt.transportOptions.IdleConnTimeout != nil {
			idleConnectionTimeout = *rt.transportOptions.IdleConnTimeout
		}

		t2 := http2.Transport{
			DialTLS:         rt.dialTLSHTTP2,
			TLSClientConfig: utlsConfig,
			ConnectionFlow:  rt.connectionFlow,
			HeaderPriority:  rt.headerPriority,
			IdleConnTimeout: idleConnectionTimeout,
		}

		if rt.transportOptions != nil {
			t2.DisableCompression = rt.transportOptions.DisableCompression

			t1 := t2.GetT1()
			if t1 != nil {
				t1.DisableKeepAlives = rt.transportOptions.DisableKeepAlives
				t1.DisableCompression = rt.transportOptions.DisableCompression
				t1.MaxIdleConns = rt.transportOptions.MaxIdleConns
				t1.MaxIdleConnsPerHost = rt.transportOptions.MaxIdleConnsPerHost
				t1.MaxConnsPerHost = rt.transportOptions.MaxConnsPerHost
				t1.MaxResponseHeaderBytes = rt.transportOptions.MaxResponseHeaderBytes
				t1.WriteBufferSize = rt.transportOptions.WriteBufferSize
				t1.ReadBufferSize = rt.transportOptions.ReadBufferSize
				t1.IdleConnTimeout = idleConnectionTimeout
			}
		}

		if rt.pseudoHeaderOrder == nil {
			t2.PseudoHeaderOrder = []string{}
		} else {
			t2.PseudoHeaderOrder = rt.pseudoHeaderOrder
		}

		if rt.settings == nil {
			// when we not provide a map of custom http2 settings
			t2.Settings = map[http2.SettingID]uint32{
				http2.SettingMaxConcurrentStreams: 1000,
				http2.SettingMaxFrameSize:         16384,
				http2.SettingInitialWindowSize:    6291456,
				http2.SettingHeaderTableSize:      65536,
			}

			keys := make([]http2.SettingID, len(t2.Settings))

			i := 0
			// attention: the order might be random here for default values!
			for k := range t2.Settings {
				keys[i] = k
				i++
			}

			t2.SettingsOrder = keys
		} else {
			// use custom http2 settings
			t2.Settings = rt.settings
			t2.SettingsOrder = rt.settingsOrder
		}

		t2.Priorities = rt.priorities

		t2.PushHandler = &http2.DefaultPushHandler{}
		rt.cachedTransports[addr] = &t2
	default:
		rt.cachedTransports[addr] = rt.buildHttp1Transport()
	}

	// Stash the connection just established for use servicing the
	// actual request (should be near-immediate).
	rt.cachedConnections[addr] = conn

	return nil, errProtocolNegotiated
}

func (rt *roundTripper) dial(ctx context.Context, network, addr string) (net.Conn, error) {
	if network == "tcp" && rt.disableIPV6 {
		network = "tcp4"
	}
	return rt.dialer.DialContext(ctx, network, addr)
}

func (rt *roundTripper) buildHttp1Transport() *http.Transport {
	utlsConfig := &tls.Config{ClientSessionCache: rt.clientSessionCache, InsecureSkipVerify: rt.insecureSkipVerify, OmitEmptyPsk: true}
	if rt.transportOptions != nil {
		utlsConfig.RootCAs = rt.transportOptions.RootCAs
	}

	if rt.serverNameOverwrite != "" {
		utlsConfig.ServerName = rt.serverNameOverwrite
	}

	idleConnectionTimeout := defaultIdleConnectionTimeout

	if rt.transportOptions != nil && rt.transportOptions.IdleConnTimeout != nil {
		idleConnectionTimeout = *rt.transportOptions.IdleConnTimeout
	}

	t := &http.Transport{DialContext: rt.dial, DialTLSContext: rt.dialTLS, TLSClientConfig: utlsConfig, ConnectionFlow: rt.connectionFlow, IdleConnTimeout: idleConnectionTimeout}

	if rt.transportOptions != nil {
		t.DisableKeepAlives = rt.transportOptions.DisableKeepAlives
		t.DisableCompression = rt.transportOptions.DisableCompression
		t.MaxIdleConns = rt.transportOptions.MaxIdleConns
		t.MaxIdleConnsPerHost = rt.transportOptions.MaxIdleConnsPerHost
		t.MaxConnsPerHost = rt.transportOptions.MaxConnsPerHost
		t.MaxResponseHeaderBytes = rt.transportOptions.MaxResponseHeaderBytes
		t.WriteBufferSize = rt.transportOptions.WriteBufferSize
		t.ReadBufferSize = rt.transportOptions.ReadBufferSize
	}

	return t
}

func (rt *roundTripper) dialTLSHTTP2(network, addr string, _ *tls.Config) (net.Conn, error) {
	return rt.dialTLS(context.Background(), network, addr)
}

func (rt *roundTripper) getDialTLSAddr(req *http.Request) string {
	host, port, err := net.SplitHostPort(req.URL.Host)
	if err == nil {
		return net.JoinHostPort(host, port)
	}

	return net.JoinHostPort(req.URL.Host, "443")
}

func newRoundTripper(clientProfile profiles.ClientProfile, transportOptions *TransportOptions, serverNameOverwrite string, insecureSkipVerify bool, withRandomTlsExtensionOrder bool, forceHttp1 bool, certificatePins map[string][]string, badPinHandlerFunc BadPinHandlerFunc, disableIPV6 bool, disableIPV4 bool, bandwidthTracker bandwidth.BandwidthTracker, dialer ...proxy.ContextDialer) (http.RoundTripper, error) {
	pinner, err := NewCertificatePinner(certificatePins)
	if err != nil {
		return nil, fmt.Errorf("can not instantiate certificate pinner: %w", err)
	}

	var clientSessionCache tls.ClientSessionCache

	withSessionResumption := supportsSessionResumption(clientProfile.GetClientHelloId())

	if withSessionResumption {
		clientSessionCache = tls.NewLRUClientSessionCache(32)
	}

	rt := &roundTripper{
		dialer:                      dialer[0],
		certificatePinner:           pinner,
		badPinHandlerFunc:           badPinHandlerFunc,
		transportOptions:            transportOptions,
		clientSessionCache:          clientSessionCache,
		serverNameOverwrite:         serverNameOverwrite,
		settings:                    clientProfile.GetSettings(),
		settingsOrder:               clientProfile.GetSettingsOrder(),
		priorities:                  clientProfile.GetPriorities(),
		headerPriority:              clientProfile.GetHeaderPriority(),
		pseudoHeaderOrder:           clientProfile.GetPseudoHeaderOrder(),
		insecureSkipVerify:          insecureSkipVerify,
		forceHttp1:                  forceHttp1,
		withRandomTlsExtensionOrder: withRandomTlsExtensionOrder,
		connectionFlow:              clientProfile.GetConnectionFlow(),
		clientHelloId:               clientProfile.GetClientHelloId(),
		cachedTransports:            make(map[string]http.RoundTripper),
		cachedConnections:           make(map[string]net.Conn),
		disableIPV6:                 disableIPV6,
		disableIPV4:                 disableIPV4,
		bandwidthTracker:            bandwidthTracker,
	}

	if len(dialer) > 0 {
		rt.dialer = dialer[0]
	} else {
		rt.dialer = proxy.Direct
	}

	return rt, nil
}

func supportsSessionResumption(id tls.ClientHelloID) bool {
	spec, err := tls.UTLSIdToSpec(id)

	if err != nil {
		spec, err = id.ToSpec()

		if err != nil {
			return false
		}
	}

	for _, ext := range spec.Extensions {
		if _, ok := ext.(*tls.UtlsPreSharedKeyExtension); ok {
			return true
		}
	}

	return false
}
