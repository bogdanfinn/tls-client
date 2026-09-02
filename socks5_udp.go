package tls_client

// This file implements SOCKS5 UDP ASSOCIATE (RFC 1928, CMD=0x03) to tunnel
// QUIC/HTTP3 traffic through a SOCKS5 proxy.
//
// Important: Not every SOCKS5 proxy supports UDP ASSOCIATE — many implementations
// only support the TCP CONNECT command. The proxy must explicitly implement
// UDP ASSOCIATE for HTTP/3 to work through it.

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/url"
	"sync"
	"time"

	quic "github.com/bogdanfinn/quic-go-utls"
	tls "github.com/bogdanfinn/utls"
)

const (
	// maxSocks5UDPPacketSize is the largest datagram we are willing to handle.
	maxSocks5UDPPacketSize = 65535

	// maxSocks5UDPHeaderSize is RSV[2] + FRAG[1] + ATYP[1] + IPv6 address[16] + port[2].
	maxSocks5UDPHeaderSize = 22

	// socks5ControlTimeout bounds the SOCKS5 negotiation on the TCP control connection.
	// The negotiation does blocking reads which do not observe the context, so without a
	// deadline a silent proxy would block the QUIC dial forever.
	socks5ControlTimeout = 30 * time.Second
)

// socks5BufferPool holds scratch buffers for encapsulating and decapsulating datagrams.
// WriteTo and ReadFrom sit on the QUIC hot path, where a buffer allocation per packet
// would add avoidable GC pressure.
var socks5BufferPool = sync.Pool{
	New: func() any {
		buf := make([]byte, maxSocks5UDPPacketSize)

		return &buf
	},
}

// socks5ControlDeadline returns the deadline to use for the SOCKS5 negotiation, honouring
// an existing context deadline if there is one.
func socks5ControlDeadline(ctx context.Context) time.Time {
	if deadline, ok := ctx.Deadline(); ok {
		return deadline
	}

	return time.Now().Add(socks5ControlTimeout)
}

// socks5Auth holds optional username/password credentials for SOCKS5 authentication.
type socks5Auth struct {
	user     string
	password string
}

// parseSOCKS5ProxyURL parses a socks5://user:pass@host:port URL into its components.
func parseSOCKS5ProxyURL(proxyURL string) (proxyAddr string, auth *socks5Auth, err error) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse SOCKS5 proxy URL: %w", err)
	}

	if u.Scheme != "socks5" && u.Scheme != "socks5h" {
		return "", nil, fmt.Errorf("unsupported proxy scheme: %s", u.Scheme)
	}

	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "1080"
	}
	proxyAddr = net.JoinHostPort(host, port)

	if u.User != nil {
		password, _ := u.User.Password()
		auth = &socks5Auth{
			user:     u.User.Username(),
			password: password,
		}
	}

	return proxyAddr, auth, nil
}

// socks5Handshake performs the SOCKS5 version/method negotiation and authentication
// on the given TCP connection per RFC 1928.
func socks5Handshake(conn net.Conn, auth *socks5Auth) error {
	// Determine methods to offer
	var methods []byte
	if auth != nil {
		methods = []byte{0x00, 0x02} // NO AUTH + USERNAME/PASSWORD
	} else {
		methods = []byte{0x00} // NO AUTH only
	}

	// Send version identifier/method selection message
	// +----+----------+----------+
	// |VER | NMETHODS | METHODS  |
	// +----+----------+----------+
	// | 1  |    1     | 1 to 255 |
	// +----+----------+----------+
	msg := make([]byte, 2+len(methods))
	msg[0] = 0x05 // SOCKS version 5
	msg[1] = byte(len(methods))
	copy(msg[2:], methods)

	if _, err := conn.Write(msg); err != nil {
		return fmt.Errorf("socks5 handshake: failed to send greeting: %w", err)
	}

	// Read server's method selection
	// +----+--------+
	// |VER | METHOD |
	// +----+--------+
	// | 1  |   1    |
	// +----+--------+
	resp := make([]byte, 2)
	if _, err := io.ReadFull(conn, resp); err != nil {
		return fmt.Errorf("socks5 handshake: failed to read method selection: %w", err)
	}

	if resp[0] != 0x05 {
		return fmt.Errorf("socks5 handshake: unexpected version %d", resp[0])
	}

	switch resp[1] {
	case 0x00:
		// No authentication required
		return nil
	case 0x02:
		// Username/password authentication (RFC 1929)
		if auth == nil {
			return fmt.Errorf("socks5 handshake: server requires auth but no credentials provided")
		}
		return socks5UsernamePasswordAuth(conn, auth)
	case 0xFF:
		return fmt.Errorf("socks5 handshake: no acceptable authentication methods")
	default:
		return fmt.Errorf("socks5 handshake: unsupported auth method 0x%02x", resp[1])
	}
}

// socks5UsernamePasswordAuth performs RFC 1929 username/password sub-negotiation.
func socks5UsernamePasswordAuth(conn net.Conn, auth *socks5Auth) error {
	// +----+------+----------+------+----------+
	// |VER | ULEN |  UNAME   | PLEN |  PASSWD  |
	// +----+------+----------+------+----------+
	// | 1  |  1   | 1 to 255 |  1   | 1 to 255 |
	// +----+------+----------+------+----------+
	//
	// Both lengths are encoded in a single byte, so reject oversized credentials instead
	// of silently truncating the length prefix and sending a malformed packet.
	if len(auth.user) > 255 {
		return fmt.Errorf("socks5 auth: username of %d bytes exceeds the maximum of 255", len(auth.user))
	}

	if len(auth.password) > 255 {
		return fmt.Errorf("socks5 auth: password of %d bytes exceeds the maximum of 255", len(auth.password))
	}

	msg := make([]byte, 3+len(auth.user)+len(auth.password))
	msg[0] = 0x01 // sub-negotiation version
	msg[1] = byte(len(auth.user))
	copy(msg[2:], auth.user)
	msg[2+len(auth.user)] = byte(len(auth.password))
	copy(msg[3+len(auth.user):], auth.password)

	if _, err := conn.Write(msg); err != nil {
		return fmt.Errorf("socks5 auth: failed to send credentials: %w", err)
	}

	// +----+--------+
	// |VER | STATUS |
	// +----+--------+
	// | 1  |   1    |
	// +----+--------+
	resp := make([]byte, 2)
	if _, err := io.ReadFull(conn, resp); err != nil {
		return fmt.Errorf("socks5 auth: failed to read response: %w", err)
	}

	if resp[1] != 0x00 {
		return fmt.Errorf("socks5 auth: authentication failed (status 0x%02x)", resp[1])
	}

	return nil
}

// socks5UDPAssociate performs the SOCKS5 UDP ASSOCIATE command (CMD=0x03) per RFC 1928 Section 4.
// It returns the TCP control connection (which must remain open) and the proxy's UDP relay address.
func socks5UDPAssociate(ctx context.Context, proxyAddr string, auth *socks5Auth, clientUDPAddr *net.UDPAddr) (controlConn net.Conn, relayAddr *net.UDPAddr, err error) {
	// TCP dial to proxy
	var d net.Dialer
	tcpConn, err := d.DialContext(ctx, "tcp", proxyAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("socks5 udp associate: failed to connect to proxy: %w", err)
	}

	defer func() {
		if err != nil {
			tcpConn.Close()
		}
	}()

	if err = tcpConn.SetDeadline(socks5ControlDeadline(ctx)); err != nil {
		return nil, nil, fmt.Errorf("socks5 udp associate: failed to set negotiation deadline: %w", err)
	}

	// Authenticate
	if err = socks5Handshake(tcpConn, auth); err != nil {
		return nil, nil, err
	}

	// Build UDP ASSOCIATE request
	// +----+-----+-------+------+----------+----------+
	// |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	// +----+-----+-------+------+----------+----------+
	// | 1  |  1  | X'00' |  1   | Variable |    2     |
	// +----+-----+-------+------+----------+----------+
	//
	// DST.ADDR and DST.PORT = client's UDP address, or 0.0.0.0:0 if unknown
	var dstAddr net.IP
	var dstPort int
	if clientUDPAddr != nil && !clientUDPAddr.IP.IsUnspecified() {
		dstAddr = clientUDPAddr.IP
		dstPort = clientUDPAddr.Port
	} else {
		dstAddr = net.IPv4zero
		dstPort = 0
	}

	req := []byte{
		0x05, // VER
		0x03, // CMD = UDP ASSOCIATE
		0x00, // RSV
	}

	if ip4 := dstAddr.To4(); ip4 != nil {
		req = append(req, 0x01) // ATYP = IPv4
		req = append(req, ip4...)
	} else {
		req = append(req, 0x04) // ATYP = IPv6
		req = append(req, dstAddr.To16()...)
	}

	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(dstPort))
	req = append(req, portBytes...)

	if _, err = tcpConn.Write(req); err != nil {
		return nil, nil, fmt.Errorf("socks5 udp associate: failed to send request: %w", err)
	}

	// Read reply
	// +----+-----+-------+------+----------+----------+
	// |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
	// +----+-----+-------+------+----------+----------+
	// | 1  |  1  | X'00' |  1   | Variable |    2     |
	// +----+-----+-------+------+----------+----------+
	header := make([]byte, 4)
	if _, err = io.ReadFull(tcpConn, header); err != nil {
		return nil, nil, fmt.Errorf("socks5 udp associate: failed to read reply header: %w", err)
	}

	if header[0] != 0x05 {
		return nil, nil, fmt.Errorf("socks5 udp associate: unexpected version %d", header[0])
	}

	if header[1] != 0x00 {
		return nil, nil, fmt.Errorf("socks5 udp associate: request failed with reply code 0x%02x", header[1])
	}

	// Parse BND.ADDR based on ATYP
	var bndIP net.IP
	switch header[3] {
	case 0x01: // IPv4
		addr := make([]byte, 4)
		if _, err = io.ReadFull(tcpConn, addr); err != nil {
			return nil, nil, fmt.Errorf("socks5 udp associate: failed to read BND.ADDR: %w", err)
		}
		bndIP = net.IP(addr)
	case 0x04: // IPv6
		addr := make([]byte, 16)
		if _, err = io.ReadFull(tcpConn, addr); err != nil {
			return nil, nil, fmt.Errorf("socks5 udp associate: failed to read BND.ADDR: %w", err)
		}
		bndIP = net.IP(addr)
	case 0x03: // Domain name
		lenBuf := make([]byte, 1)
		if _, err = io.ReadFull(tcpConn, lenBuf); err != nil {
			return nil, nil, fmt.Errorf("socks5 udp associate: failed to read domain length: %w", err)
		}
		domain := make([]byte, lenBuf[0])
		if _, err = io.ReadFull(tcpConn, domain); err != nil {
			return nil, nil, fmt.Errorf("socks5 udp associate: failed to read domain: %w", err)
		}
		ips, resolveErr := net.ResolveIPAddr("ip", string(domain))
		if resolveErr != nil {
			return nil, nil, fmt.Errorf("socks5 udp associate: failed to resolve BND.ADDR domain %q: %w", string(domain), resolveErr)
		}
		bndIP = ips.IP
	default:
		return nil, nil, fmt.Errorf("socks5 udp associate: unsupported ATYP 0x%02x", header[3])
	}

	// Read BND.PORT
	portBuf := make([]byte, 2)
	if _, err = io.ReadFull(tcpConn, portBuf); err != nil {
		return nil, nil, fmt.Errorf("socks5 udp associate: failed to read BND.PORT: %w", err)
	}
	bndPort := binary.BigEndian.Uint16(portBuf)

	// If BND.ADDR is 0.0.0.0 or ::, use the proxy's IP instead
	if bndIP.IsUnspecified() {
		proxyHost, _, _ := net.SplitHostPort(proxyAddr)
		bndIP = net.ParseIP(proxyHost)
		if bndIP == nil {
			// proxyHost might be a hostname, resolve it
			ips, resolveErr := net.ResolveIPAddr("ip", proxyHost)
			if resolveErr != nil {
				return nil, nil, fmt.Errorf("socks5 udp associate: failed to resolve proxy host %q: %w", proxyHost, resolveErr)
			}
			bndIP = ips.IP
		}
	}

	// The control connection has to stay open for the lifetime of the association, so
	// drop the negotiation deadline again.
	if err = tcpConn.SetDeadline(time.Time{}); err != nil {
		return nil, nil, fmt.Errorf("socks5 udp associate: failed to clear negotiation deadline: %w", err)
	}

	relayAddr = &net.UDPAddr{
		IP:   bndIP,
		Port: int(bndPort),
	}

	return tcpConn, relayAddr, nil
}

// socks5PacketConn wraps a UDP connection through a SOCKS5 UDP ASSOCIATE relay.
// It implements net.PacketConn so it can be used with quic.Transport.
//
// Do not add SyscallConn, ReadMsgUDP and WriteMsgUDP passthroughs to this type. That
// would make it satisfy quic.OOBCapablePacketConn, which makes quic-go read and write
// through those methods directly instead of through ReadFrom and WriteTo. The SOCKS5
// encapsulation would be bypassed, which silently breaks the tunnel and leaks the real
// IP address, the exact thing this type exists to prevent.
type socks5PacketConn struct {
	udpConn     *net.UDPConn // local UDP socket
	controlConn net.Conn     // TCP control connection (must stay open)
	relayAddr   *net.UDPAddr // proxy's UDP relay endpoint
	closeMu     sync.Mutex
	closed      bool
}

// WriteTo encodes the SOCKS5 UDP header and sends the packet through the relay.
func (c *socks5PacketConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	udpAddr, ok := addr.(*net.UDPAddr)
	if !ok {
		return 0, fmt.Errorf("socks5PacketConn: expected *net.UDPAddr, got %T", addr)
	}

	if len(p) > maxSocks5UDPPacketSize-maxSocks5UDPHeaderSize {
		return 0, fmt.Errorf("socks5PacketConn: payload of %d bytes is too large to encapsulate", len(p))
	}

	bufPtr := socks5BufferPool.Get().(*[]byte)
	defer socks5BufferPool.Put(bufPtr)
	buf := *bufPtr

	headerLen := encodeSocks5UDPHeader(buf, udpAddr)
	payloadLen := copy(buf[headerLen:], p)

	if _, err := c.udpConn.WriteToUDP(buf[:headerLen+payloadLen], c.relayAddr); err != nil {
		return 0, err
	}

	// Report the payload bytes accepted. The header is an implementation detail of the
	// tunnel and must not be visible to the caller.
	return payloadLen, nil
}

// ReadFrom reads a packet from the relay, strips the SOCKS5 UDP header, and returns the payload.
//
// Datagrams that did not come from the relay, or that carry a header we cannot decode,
// are dropped and the next datagram is read. Returning an error instead would be fatal:
// quic-go tears down the whole transport when the packet conn reports a read error, so a
// single stray datagram would kill the connection.
func (c *socks5PacketConn) ReadFrom(p []byte) (int, net.Addr, error) {
	bufPtr := socks5BufferPool.Get().(*[]byte)
	defer socks5BufferPool.Put(bufPtr)
	buf := *bufPtr

	for {
		n, srcAddr, err := c.udpConn.ReadFromUDP(buf)
		if err != nil {
			return 0, nil, err
		}

		// The local socket is unconnected, so without this check anyone able to reach it
		// could inject datagrams into the QUIC connection. Only the address is compared,
		// not the port: some proxies relay from a different port than the one they
		// announced in BND.PORT.
		if srcAddr != nil && !srcAddr.IP.Equal(c.relayAddr.IP) {
			continue
		}

		udpAddr, dataOffset, decodeErr := decodeSocks5UDPHeader(buf[:n])
		if decodeErr != nil {
			continue
		}

		return copy(p, buf[dataOffset:n]), udpAddr, nil
	}
}

// Close closes both the UDP socket and the TCP control connection,
// which terminates the UDP association per RFC 1928.
func (c *socks5PacketConn) Close() error {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

	udpErr := c.udpConn.Close()
	ctrlErr := c.controlConn.Close()

	if udpErr != nil {
		return udpErr
	}
	return ctrlErr
}

// LocalAddr returns the local address of the underlying UDP socket.
func (c *socks5PacketConn) LocalAddr() net.Addr {
	return c.udpConn.LocalAddr()
}

func (c *socks5PacketConn) SetDeadline(t time.Time) error {
	return c.udpConn.SetDeadline(t)
}

func (c *socks5PacketConn) SetReadDeadline(t time.Time) error {
	return c.udpConn.SetReadDeadline(t)
}

func (c *socks5PacketConn) SetWriteDeadline(t time.Time) error {
	return c.udpConn.SetWriteDeadline(t)
}

// encodeSocks5UDPHeader builds a SOCKS5 UDP request header for the given destination address.
// +------+------+------+----------+----------+
// | RSV  | FRAG | ATYP | DST.ADDR | DST.PORT |
// +------+------+------+----------+----------+
// |  2   |  1   |  1   | Variable |    2     |
// +------+------+------+----------+----------+
//
// The header is written into dst, which must be at least maxSocks5UDPHeaderSize bytes
// long, and the number of bytes written is returned.
func encodeSocks5UDPHeader(dst []byte, addr *net.UDPAddr) int {
	// RSV (2 bytes) + FRAG (1 byte, 0 = standalone)
	dst[0], dst[1], dst[2] = 0x00, 0x00, 0x00

	offset := 4

	if ip4 := addr.IP.To4(); ip4 != nil {
		dst[3] = 0x01 // ATYP = IPv4
		offset += copy(dst[offset:], ip4)
	} else {
		dst[3] = 0x04 // ATYP = IPv6
		offset += copy(dst[offset:], addr.IP.To16())
	}

	binary.BigEndian.PutUint16(dst[offset:], uint16(addr.Port))

	return offset + 2
}

// decodeSocks5UDPHeader parses a SOCKS5 UDP reply header and returns the source address
// and the offset where the payload data starts.
func decodeSocks5UDPHeader(data []byte) (addr *net.UDPAddr, dataOffset int, err error) {
	if len(data) < 4 {
		return nil, 0, fmt.Errorf("socks5 udp header too short: %d bytes", len(data))
	}

	// RSV[2] + FRAG[1] + ATYP[1]
	//
	// Reassembling fragments is not supported, and a relay has no reason to fragment
	// QUIC datagrams, so anything with FRAG set is rejected.
	if data[2] != 0x00 {
		return nil, 0, fmt.Errorf("socks5 udp header: fragmented datagrams are not supported (FRAG=0x%02x)", data[2])
	}

	atyp := data[3]
	offset := 4

	var ip net.IP
	switch atyp {
	case 0x01: // IPv4
		if len(data) < offset+4+2 {
			return nil, 0, fmt.Errorf("socks5 udp header too short for IPv4")
		}
		// Copied, because data is a scratch buffer that is reused for the next datagram.
		ip = append(net.IP(nil), data[offset:offset+4]...)
		offset += 4
	case 0x04: // IPv6
		if len(data) < offset+16+2 {
			return nil, 0, fmt.Errorf("socks5 udp header too short for IPv6")
		}
		ip = append(net.IP(nil), data[offset:offset+16]...)
		offset += 16
	case 0x03: // Domain name
		// A relay reports the source of a datagram as an address, not as a name.
		// Resolving a name here would mean a blocking DNS lookup on the QUIC receive
		// path, driven by data we do not control, so it is refused instead.
		return nil, 0, fmt.Errorf("socks5 udp header: unsupported ATYP 0x03 (domain name) in a relayed datagram")
	default:
		return nil, 0, fmt.Errorf("unsupported ATYP 0x%02x", atyp)
	}

	if len(data) < offset+2 {
		return nil, 0, fmt.Errorf("socks5 udp header too short for port")
	}
	port := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	return &net.UDPAddr{IP: ip, Port: int(port)}, offset, nil
}

// newSOCKS5QUICDialer returns a function matching the http3.Transport.Dial signature
// that tunnels QUIC traffic through a SOCKS5 proxy's UDP ASSOCIATE relay.
func newSOCKS5QUICDialer(proxyURL string) func(ctx context.Context, addr string, tlsCfg *tls.Config, cfg *quic.Config) (*quic.Conn, error) {
	return func(ctx context.Context, addr string, tlsCfg *tls.Config, cfg *quic.Config) (*quic.Conn, error) {
		proxyAddr, auth, err := parseSOCKS5ProxyURL(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("socks5 quic dialer: %w", err)
		}

		// Create a local UDP socket
		localUDP, err := net.ListenUDP("udp", nil)
		if err != nil {
			return nil, fmt.Errorf("socks5 quic dialer: failed to create UDP socket: %w", err)
		}

		// Establish UDP ASSOCIATE session
		controlConn, relayAddr, err := socks5UDPAssociate(ctx, proxyAddr, auth, localUDP.LocalAddr().(*net.UDPAddr))
		if err != nil {
			localUDP.Close()
			return nil, fmt.Errorf("socks5 quic dialer: %w", err)
		}

		// Wrap in socks5PacketConn
		packetConn := &socks5PacketConn{
			udpConn:     localUDP,
			controlConn: controlConn,
			relayAddr:   relayAddr,
		}

		// Create QUIC transport with our packet conn
		quicTransport := &quic.Transport{
			Conn: packetConn,
		}

		// Resolve the target address.
		//
		// Note that this lookup happens locally and is therefore not covered by the
		// proxy. UDP ASSOCIATE could carry the name to the relay in the per packet
		// header (ATYP=0x03), but quic-go needs a concrete *net.UDPAddr for its path
		// handling, so the name has to be resolved on this side. Only the QUIC traffic
		// itself is tunnelled, not this DNS query.
		targetAddr, err := resolveUDPAddr(ctx, addr, relayAddr.IP)
		if err != nil {
			packetConn.Close()
			return nil, fmt.Errorf("socks5 quic dialer: failed to resolve target %q: %w", addr, err)
		}

		// Dial QUIC through the SOCKS5 relay
		conn, err := quicTransport.DialEarly(ctx, targetAddr, tlsCfg, cfg)
		if err != nil {
			// DialEarly has already started the transport's read goroutine, so the
			// transport needs to be closed as well, not just the packet conn.
			_ = quicTransport.Close()
			packetConn.Close()

			return nil, fmt.Errorf("socks5 quic dialer: QUIC dial failed: %w", err)
		}

		// A quic.Transport only closes a Conn that it created itself, and nothing outside
		// this function holds a reference to the transport or to the packet conn, so tie
		// the lifetime of both to the connection. Without this, every HTTP/3 connection
		// through the proxy would leak a UDP socket, the TCP control connection and the
		// transport's read goroutine.
		go func() {
			<-conn.Context().Done()

			_ = quicTransport.Close()
			_ = packetConn.Close()
		}()

		return conn, nil
	}
}

// resolveUDPAddr resolves a host:port pair into a *net.UDPAddr. Unlike
// net.ResolveUDPAddr it honours the context.
//
// When prefer is set, an address of the same family is picked over the resolver's own
// order. The relay of a SOCKS5 proxy that only speaks IPv4 cannot forward to an IPv6
// target, and on a dual stack host the resolver commonly returns the IPv6 address first,
// so taking the first one blindly leaves the QUIC handshake without a reply.
func resolveUDPAddr(ctx context.Context, addr string, prefer net.IP) (*net.UDPAddr, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	portNumber, err := net.DefaultResolver.LookupPort(ctx, "udp", port)
	if err != nil {
		return nil, err
	}

	if ip := net.ParseIP(host); ip != nil {
		return &net.UDPAddr{IP: ip, Port: portNumber}, nil
	}

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no addresses found for %q", host)
	}

	chosen := pickIPAddr(ips, prefer)

	return &net.UDPAddr{IP: chosen.IP, Port: portNumber, Zone: chosen.Zone}, nil
}

// pickIPAddr returns the first address of the same family as prefer, falling back to the
// resolver's own first choice when prefer is nil or no address matches.
func pickIPAddr(ips []net.IPAddr, prefer net.IP) net.IPAddr {
	if prefer == nil {
		return ips[0]
	}

	preferIPv4 := prefer.To4() != nil

	for _, candidate := range ips {
		if (candidate.IP.To4() != nil) == preferIPv4 {
			return candidate
		}
	}

	return ips[0]
}
