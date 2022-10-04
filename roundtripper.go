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

	clientHelloId     utls.ClientHelloID
	settings          map[http2.SettingID]uint32
	settingsOrder     []http2.SettingID
	priorities        []http2.Priority
	pseudoHeaderOrder []string
	connectionFlow    uint32

	insecureSkipVerify bool

	cachedTransportsLck sync.Mutex
	cachedConnections   map[string]net.Conn
	cachedTransports    map[string]http.RoundTripper

	dialer proxy.ContextDialer
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	addr := rt.getDialTLSAddr(req)

	rt.cachedTransportsLck.Lock()
	defer rt.cachedTransportsLck.Unlock()

	if _, ok := rt.cachedTransports[addr]; !ok {
		if err := rt.getTransport(req, addr); err != nil {
			return nil, err
		}
	}
	return rt.cachedTransports[addr].RoundTrip(req)
}

func (rt *roundTripper) getTransport(req *http.Request, addr string) error {
	switch strings.ToLower(req.URL.Scheme) {
	case "http":
		rt.cachedTransports[addr] = &http.Transport{DialContext: rt.dialer.DialContext, PseudoHeaderOrder: rt.pseudoHeaderOrder, ConnectionFlow: rt.connectionFlow}
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

	conn := utls.UClient(rawConn, &utls.Config{ServerName: host, InsecureSkipVerify: rt.insecureSkipVerify}, rt.clientHelloId)
	if err = conn.Handshake(); err != nil {
		_ = conn.Close()
		return nil, err
	}

	if rt.cachedTransports[addr] != nil {
		return conn, nil
	}

	// No http.Transport constructed yet, create one based on the results
	// of ALPN.
	switch conn.ConnectionState().NegotiatedProtocol {
	case http2.NextProtoTLS:
		t2 := http2.Transport{DialTLS: rt.dialTLSHTTP2, TLSClientConfig: &utls.Config{InsecureSkipVerify: rt.insecureSkipVerify}, ConnectionFlow: rt.connectionFlow}

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
		// Assume the remote peer is speaking HTTP 1.x + TLS.
		rt.cachedTransports[addr] = &http.Transport{DialTLSContext: rt.dialTLS, TLSClientConfig: &utls.Config{InsecureSkipVerify: rt.insecureSkipVerify}, ConnectionFlow: rt.connectionFlow}
	}

	// Stash the connection just established for use servicing the
	// actual request (should be near-immediate).
	rt.cachedConnections[addr] = conn

	return nil, errProtocolNegotiated
}

func (rt *roundTripper) dialTLSHTTP2(network, addr string, _ *utls.Config) (net.Conn, error) {
	return rt.dialTLS(context.Background(), network, addr)
}

func (rt *roundTripper) getDialTLSAddr(req *http.Request) string {
	host, port, err := net.SplitHostPort(req.URL.Host)
	if err == nil {
		return net.JoinHostPort(host, port)
	}
	return net.JoinHostPort(req.URL.Host, "443") // we can assume port is 443 at this point
}

func newRoundTripper(clientProfile ClientProfile, insecureSkipVerify bool, dialer ...proxy.ContextDialer) http.RoundTripper {
	if len(dialer) > 0 {
		return &roundTripper{
			dialer:             dialer[0],
			settings:           clientProfile.settings,
			settingsOrder:      clientProfile.settingsOrder,
			priorities:         clientProfile.priorities,
			pseudoHeaderOrder:  clientProfile.pseudoHeaderOrder,
			insecureSkipVerify: insecureSkipVerify,
			connectionFlow:     clientProfile.connectionFlow,
			clientHelloId:      clientProfile.clientHelloId,
			cachedTransports:   make(map[string]http.RoundTripper),
			cachedConnections:  make(map[string]net.Conn),
		}
	} else {
		return &roundTripper{
			dialer:             proxy.Direct,
			settings:           clientProfile.settings,
			settingsOrder:      clientProfile.settingsOrder,
			priorities:         clientProfile.priorities,
			pseudoHeaderOrder:  clientProfile.pseudoHeaderOrder,
			insecureSkipVerify: insecureSkipVerify,
			connectionFlow:     clientProfile.connectionFlow,
			clientHelloId:      clientProfile.clientHelloId,
			cachedTransports:   make(map[string]http.RoundTripper),
			cachedConnections:  make(map[string]net.Conn),
		}
	}
}
