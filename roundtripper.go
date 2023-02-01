package tls_client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/http2"
	"golang.org/x/net/proxy"

	utls "github.com/bogdanfinn/utls"
)

var errProtocolNegotiated = errors.New("protocol negotiated")

type roundTripper struct {
	sync.Mutex
	transportOptions    *TransportOptions
	serverNameOverwrite string
	clientHelloId       utls.ClientHelloID
	settings            map[http2.SettingID]uint32
	settingsOrder       []http2.SettingID
	priorities          []http2.Priority
	headerPriority      *http2.PriorityParam
	pseudoHeaderOrder   []string
	connectionFlow      uint32

	insecureSkipVerify          bool
	withRandomTlsExtensionOrder bool

	cachedTransportsLck sync.Mutex
	cachedConnections   map[string]net.Conn
	cachedTransports    map[string]http.RoundTripper

	forceHttp1 bool

	dialer            proxy.ContextDialer
	certificatePinner CertificatePinner
	badPinHandlerFunc BadPinHandlerFunc
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

	_, err := rt.dialTLS(context.Background(), "tcp", addr)
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

	conn := utls.UClient(rawConn, &utls.Config{ServerName: host, InsecureSkipVerify: rt.insecureSkipVerify}, rt.clientHelloId, rt.withRandomTlsExtensionOrder, rt.forceHttp1)
	if err = conn.Handshake(); err != nil {
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
		utlsConfig := &utls.Config{InsecureSkipVerify: rt.insecureSkipVerify}

		if rt.serverNameOverwrite != "" {
			utlsConfig.ServerName = rt.serverNameOverwrite
		}

		t2 := http2.Transport{DialTLS: rt.dialTLSHTTP2, TLSClientConfig: utlsConfig, ConnectionFlow: rt.connectionFlow, HeaderPriority: rt.headerPriority}

		if rt.transportOptions != nil {
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

func (rt *roundTripper) buildHttp1Transport() *http.Transport {
	utlsConfig := &utls.Config{InsecureSkipVerify: rt.insecureSkipVerify}

	if rt.serverNameOverwrite != "" {
		utlsConfig.ServerName = rt.serverNameOverwrite
	}

	t := &http.Transport{DialTLSContext: rt.dialTLS, TLSClientConfig: utlsConfig, ConnectionFlow: rt.connectionFlow}

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

func (rt *roundTripper) dialTLSHTTP2(network, addr string, _ *utls.Config) (net.Conn, error) {
	return rt.dialTLS(context.Background(), network, addr)
}

func (rt *roundTripper) getDialTLSAddr(req *http.Request) string {
	host, port, err := net.SplitHostPort(req.URL.Host)
	if err == nil {
		return net.JoinHostPort(host, port)
	}

	return net.JoinHostPort(req.URL.Host, "443")
}

func newRoundTripper(clientProfile ClientProfile, transportOptions *TransportOptions, serverNameOverwrite string, insecureSkipVerify bool, withRandomTlsExtensionOrder bool, forceHttp1 bool, certificatePins map[string][]string, badPinHandlerFunc BadPinHandlerFunc, dialer ...proxy.ContextDialer) (http.RoundTripper, error) {
	pinner, err := NewCertificatePinner(certificatePins)

	if err != nil {
		return nil, fmt.Errorf("can not instantiate certificate pinner: %w", err)
	}

	rt := &roundTripper{
		dialer:                      dialer[0],
		certificatePinner:           pinner,
		badPinHandlerFunc:           badPinHandlerFunc,
		transportOptions:            transportOptions,
		serverNameOverwrite:         serverNameOverwrite,
		settings:                    clientProfile.settings,
		settingsOrder:               clientProfile.settingsOrder,
		priorities:                  clientProfile.priorities,
		headerPriority:              clientProfile.headerPriority,
		pseudoHeaderOrder:           clientProfile.pseudoHeaderOrder,
		insecureSkipVerify:          insecureSkipVerify,
		forceHttp1:                  forceHttp1,
		withRandomTlsExtensionOrder: withRandomTlsExtensionOrder,
		connectionFlow:              clientProfile.connectionFlow,
		clientHelloId:               clientProfile.clientHelloId,
		cachedTransports:            make(map[string]http.RoundTripper),
		cachedConnections:           make(map[string]net.Conn),
	}

	if len(dialer) > 0 {
		rt.dialer = dialer[0]
	} else {
		rt.dialer = proxy.Direct
	}

	return rt, nil
}
