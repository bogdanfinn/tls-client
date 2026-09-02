package tls_client

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/quic-go-utls/http3"
	"github.com/bogdanfinn/tls-client/profiles"
)

// --- parseSOCKS5ProxyURL tests ---

func TestParseSOCKS5ProxyURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantAddr string
		wantUser string
		wantPass string
		wantAuth bool
		wantErr  bool
	}{
		{
			name:     "socks5 with auth",
			input:    "socks5://user:pass@127.0.0.1:1080",
			wantAddr: "127.0.0.1:1080",
			wantUser: "user",
			wantPass: "pass",
			wantAuth: true,
		},
		{
			name:     "socks5h with auth",
			input:    "socks5h://admin:secret@proxy.example.com:9050",
			wantAddr: "proxy.example.com:9050",
			wantUser: "admin",
			wantPass: "secret",
			wantAuth: true,
		},
		{
			name:     "socks5 without auth",
			input:    "socks5://192.168.1.1:1080",
			wantAddr: "192.168.1.1:1080",
			wantAuth: false,
		},
		{
			name:     "socks5 default port",
			input:    "socks5://10.0.0.1",
			wantAddr: "10.0.0.1:1080",
			wantAuth: false,
		},
		{
			name:     "socks5 with username only",
			input:    "socks5://user@127.0.0.1:1080",
			wantUser: "user",
			wantPass: "",
			wantAddr: "127.0.0.1:1080",
			wantAuth: true,
		},
		{
			name:    "http scheme rejected",
			input:   "http://proxy.example.com:8080",
			wantErr: true,
		},
		{
			name:    "https scheme rejected",
			input:   "https://proxy.example.com:443",
			wantErr: true,
		},
		{
			name:    "socks4 scheme rejected",
			input:   "socks4://proxy.example.com:1080",
			wantErr: true,
		},
		{
			name:    "invalid URL",
			input:   "://not-a-url",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			addr, auth, err := parseSOCKS5ProxyURL(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if addr != tc.wantAddr {
				t.Errorf("addr = %q, want %q", addr, tc.wantAddr)
			}
			if tc.wantAuth {
				if auth == nil {
					t.Fatal("expected auth, got nil")
				}
				if auth.user != tc.wantUser {
					t.Errorf("user = %q, want %q", auth.user, tc.wantUser)
				}
				if auth.password != tc.wantPass {
					t.Errorf("password = %q, want %q", auth.password, tc.wantPass)
				}
			} else {
				if auth != nil {
					t.Errorf("expected no auth, got %+v", auth)
				}
			}
		})
	}
}

// --- UDP header encode/decode roundtrip tests ---

func TestSocks5UDPHeaderRoundtrip_IPv4(t *testing.T) {
	original := &net.UDPAddr{
		IP:   net.IPv4(93, 184, 216, 34),
		Port: 443,
	}

	headerBuf := make([]byte, maxSocks5UDPHeaderSize)
	header := headerBuf[:encodeSocks5UDPHeader(headerBuf, original)]

	// Verify header structure: RSV(2) + FRAG(1) + ATYP(1) + IPv4(4) + Port(2) = 10 bytes
	if len(header) != 10 {
		t.Fatalf("header length = %d, want 10", len(header))
	}
	if header[0] != 0 || header[1] != 0 {
		t.Error("RSV bytes should be 0")
	}
	if header[2] != 0 {
		t.Error("FRAG byte should be 0")
	}
	if header[3] != 0x01 {
		t.Errorf("ATYP = 0x%02x, want 0x01 (IPv4)", header[3])
	}

	// Append some payload
	payload := []byte("hello QUIC")
	packet := append(header, payload...)

	decoded, offset, err := decodeSocks5UDPHeader(packet)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if offset != 10 {
		t.Errorf("data offset = %d, want 10", offset)
	}
	if !decoded.IP.Equal(original.IP) {
		t.Errorf("decoded IP = %v, want %v", decoded.IP, original.IP)
	}
	if decoded.Port != original.Port {
		t.Errorf("decoded port = %d, want %d", decoded.Port, original.Port)
	}
	if string(packet[offset:]) != "hello QUIC" {
		t.Errorf("payload mismatch: %q", string(packet[offset:]))
	}
}

func TestSocks5UDPHeaderRoundtrip_IPv6(t *testing.T) {
	original := &net.UDPAddr{
		IP:   net.ParseIP("2606:4700:4700::1111"),
		Port: 8443,
	}

	headerBuf := make([]byte, maxSocks5UDPHeaderSize)
	header := headerBuf[:encodeSocks5UDPHeader(headerBuf, original)]

	// RSV(2) + FRAG(1) + ATYP(1) + IPv6(16) + Port(2) = 22 bytes
	if len(header) != 22 {
		t.Fatalf("header length = %d, want 22", len(header))
	}
	if header[3] != 0x04 {
		t.Errorf("ATYP = 0x%02x, want 0x04 (IPv6)", header[3])
	}

	decoded, offset, err := decodeSocks5UDPHeader(header)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if offset != 22 {
		t.Errorf("data offset = %d, want 22", offset)
	}
	if !decoded.IP.Equal(original.IP) {
		t.Errorf("decoded IP = %v, want %v", decoded.IP, original.IP)
	}
	if decoded.Port != original.Port {
		t.Errorf("decoded port = %d, want %d", decoded.Port, original.Port)
	}
}

func TestDecodeSocks5UDPHeader_TooShort(t *testing.T) {
	_, _, err := decodeSocks5UDPHeader([]byte{0x00, 0x00})
	if err == nil {
		t.Fatal("expected error for too-short header")
	}
}

func TestDecodeSocks5UDPHeader_UnsupportedATYP(t *testing.T) {
	data := []byte{0x00, 0x00, 0x00, 0x99, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	_, _, err := decodeSocks5UDPHeader(data)
	if err == nil {
		t.Fatal("expected error for unsupported ATYP")
	}
}

func TestDecodeSocks5UDPHeader_IPv4TooShort(t *testing.T) {
	// ATYP=IPv4 but not enough bytes for address + port
	data := []byte{0x00, 0x00, 0x00, 0x01, 0x01, 0x02}
	_, _, err := decodeSocks5UDPHeader(data)
	if err == nil {
		t.Fatal("expected error for truncated IPv4 header")
	}
}

func TestDecodeSocks5UDPHeader_IPv6TooShort(t *testing.T) {
	// ATYP=IPv6 but not enough bytes
	data := []byte{0x00, 0x00, 0x00, 0x04, 0x01, 0x02, 0x03, 0x04}
	_, _, err := decodeSocks5UDPHeader(data)
	if err == nil {
		t.Fatal("expected error for truncated IPv6 header")
	}
}

// --- socks5Handshake tests with mock server ---

func TestSocks5Handshake_NoAuth(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	done := make(chan error, 1)
	go func() {
		done <- socks5Handshake(client, nil)
	}()

	// Read greeting from client
	buf := make([]byte, 3)
	if _, err := io.ReadFull(server, buf); err != nil {
		t.Fatalf("server read greeting: %v", err)
	}
	if buf[0] != 0x05 {
		t.Fatalf("version = 0x%02x, want 0x05", buf[0])
	}
	if buf[1] != 1 { // 1 method (NO AUTH)
		t.Fatalf("nMethods = %d, want 1", buf[1])
	}
	if buf[2] != 0x00 {
		t.Fatalf("method = 0x%02x, want 0x00 (NO AUTH)", buf[2])
	}

	// Reply: select NO AUTH
	if _, err := server.Write([]byte{0x05, 0x00}); err != nil {
		t.Fatalf("server write: %v", err)
	}

	if err := <-done; err != nil {
		t.Fatalf("handshake error: %v", err)
	}
}

func TestSocks5Handshake_UsernamePassword(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	auth := &socks5Auth{user: "testuser", password: "testpass"}
	done := make(chan error, 1)
	go func() {
		done <- socks5Handshake(client, auth)
	}()

	// Read greeting
	buf := make([]byte, 4) // VER + NMETHODS + 2 methods
	if _, err := io.ReadFull(server, buf); err != nil {
		t.Fatalf("server read greeting: %v", err)
	}
	if buf[0] != 0x05 {
		t.Fatalf("version = 0x%02x, want 0x05", buf[0])
	}
	if buf[1] != 2 { // 2 methods: NO AUTH + USERNAME/PASSWORD
		t.Fatalf("nMethods = %d, want 2", buf[1])
	}

	// Reply: select USERNAME/PASSWORD auth (0x02)
	if _, err := server.Write([]byte{0x05, 0x02}); err != nil {
		t.Fatalf("server write method: %v", err)
	}

	// Read auth sub-negotiation
	// VER(1) + ULEN(1) + USER(8) + PLEN(1) + PASS(8) = 19
	authBuf := make([]byte, 19)
	if _, err := io.ReadFull(server, authBuf); err != nil {
		t.Fatalf("server read auth: %v", err)
	}
	if authBuf[0] != 0x01 {
		t.Fatalf("auth version = 0x%02x, want 0x01", authBuf[0])
	}
	if authBuf[1] != 8 {
		t.Fatalf("ulen = %d, want 8", authBuf[1])
	}
	if string(authBuf[2:10]) != "testuser" {
		t.Fatalf("username = %q, want %q", string(authBuf[2:10]), "testuser")
	}
	if authBuf[10] != 8 {
		t.Fatalf("plen = %d, want 8", authBuf[10])
	}
	if string(authBuf[11:19]) != "testpass" {
		t.Fatalf("password = %q, want %q", string(authBuf[11:19]), "testpass")
	}

	// Reply: auth success
	if _, err := server.Write([]byte{0x01, 0x00}); err != nil {
		t.Fatalf("server write auth response: %v", err)
	}

	if err := <-done; err != nil {
		t.Fatalf("handshake error: %v", err)
	}
}

func TestSocks5Handshake_AuthFailure(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	auth := &socks5Auth{user: "bad", password: "creds"}
	done := make(chan error, 1)
	go func() {
		done <- socks5Handshake(client, auth)
	}()

	// Read greeting
	buf := make([]byte, 4)
	io.ReadFull(server, buf)

	// Select USERNAME/PASSWORD
	server.Write([]byte{0x05, 0x02})

	// Read auth sub-negotiation (VER + ULEN + "bad" + PLEN + "creds" = 1+1+3+1+5 = 11)
	authBuf := make([]byte, 11)
	io.ReadFull(server, authBuf)

	// Reply: auth failure
	server.Write([]byte{0x01, 0x01})

	err := <-done
	if err == nil {
		t.Fatal("expected auth failure error")
	}
}

func TestSocks5Handshake_NoAcceptableMethods(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	done := make(chan error, 1)
	go func() {
		done <- socks5Handshake(client, nil)
	}()

	// Read greeting
	buf := make([]byte, 3)
	io.ReadFull(server, buf)

	// Reply: no acceptable methods
	server.Write([]byte{0x05, 0xFF})

	err := <-done
	if err == nil {
		t.Fatal("expected no-acceptable-methods error")
	}
}

// --- socks5UDPAssociate tests with mock server ---

func TestSocks5UDPAssociate_Success(t *testing.T) {
	// Start a mock SOCKS5 server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer listener.Close()

	proxyAddr := listener.Addr().String()
	relayPort := uint16(19999)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Handle handshake: read greeting
		buf := make([]byte, 3) // VER + NMETHODS + 1 method
		io.ReadFull(conn, buf)

		// Reply: NO AUTH
		conn.Write([]byte{0x05, 0x00})

		// Read UDP ASSOCIATE request
		// VER(1) + CMD(1) + RSV(1) + ATYP(1) + IPv4(4) + PORT(2) = 10
		reqBuf := make([]byte, 10)
		io.ReadFull(conn, reqBuf)

		if reqBuf[1] != 0x03 {
			t.Errorf("CMD = 0x%02x, want 0x03 (UDP ASSOCIATE)", reqBuf[1])
		}

		// Reply: success, relay at 127.0.0.1:relayPort
		reply := []byte{
			0x05, 0x00, 0x00, // VER, REP (success), RSV
			0x01,         // ATYP = IPv4
			127, 0, 0, 1, // BND.ADDR
			0x00, 0x00, // BND.PORT (filled below)
		}
		binary.BigEndian.PutUint16(reply[8:10], relayPort)
		conn.Write(reply)

		// Keep the control connection open until the test finishes
		// (closing it would terminate the UDP association)
		buf = make([]byte, 1)
		conn.Read(buf) // blocks until client closes
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	controlConn, relay, err := socks5UDPAssociate(ctx, proxyAddr, nil, nil)
	if err != nil {
		t.Fatalf("socks5UDPAssociate error: %v", err)
	}
	defer controlConn.Close()

	if relay.Port != int(relayPort) {
		t.Errorf("relay port = %d, want %d", relay.Port, relayPort)
	}
	if !relay.IP.Equal(net.IPv4(127, 0, 0, 1)) {
		t.Errorf("relay IP = %v, want 127.0.0.1", relay.IP)
	}
}

func TestSocks5UDPAssociate_ZeroBndAddr(t *testing.T) {
	// When proxy returns 0.0.0.0 as BND.ADDR, we should replace it with the proxy's IP
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer listener.Close()

	proxyAddr := listener.Addr().String()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Handshake
		buf := make([]byte, 3)
		io.ReadFull(conn, buf)
		conn.Write([]byte{0x05, 0x00})

		// Read UDP ASSOCIATE request
		reqBuf := make([]byte, 10)
		io.ReadFull(conn, reqBuf)

		// Reply with 0.0.0.0 as BND.ADDR
		reply := []byte{
			0x05, 0x00, 0x00,
			0x01,
			0, 0, 0, 0, // 0.0.0.0
			0x4E, 0x20, // port 20000
		}
		conn.Write(reply)

		buf = make([]byte, 1)
		conn.Read(buf)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	controlConn, relay, err := socks5UDPAssociate(ctx, proxyAddr, nil, nil)
	if err != nil {
		t.Fatalf("socks5UDPAssociate error: %v", err)
	}
	defer controlConn.Close()

	// Should have replaced 0.0.0.0 with proxy IP (127.0.0.1)
	if relay.IP.IsUnspecified() {
		t.Error("relay IP should not be 0.0.0.0 — should be replaced with proxy IP")
	}
	if !relay.IP.Equal(net.IPv4(127, 0, 0, 1)) {
		t.Errorf("relay IP = %v, want 127.0.0.1", relay.IP)
	}
}

func TestSocks5UDPAssociate_RequestRejected(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer listener.Close()

	proxyAddr := listener.Addr().String()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Handshake
		buf := make([]byte, 3)
		io.ReadFull(conn, buf)
		conn.Write([]byte{0x05, 0x00})

		// Read UDP ASSOCIATE request
		reqBuf := make([]byte, 10)
		io.ReadFull(conn, reqBuf)

		// Reply: command not supported (0x07)
		reply := []byte{
			0x05, 0x07, 0x00,
			0x01,
			0, 0, 0, 0,
			0x00, 0x00,
		}
		conn.Write(reply)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err = socks5UDPAssociate(ctx, proxyAddr, nil, nil)
	if err == nil {
		t.Fatal("expected error for rejected UDP ASSOCIATE request")
	}
}

func TestSocks5UDPAssociate_WithAuth(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer listener.Close()

	proxyAddr := listener.Addr().String()
	auth := &socks5Auth{user: "u", password: "p"}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Read greeting: VER + NMETHODS + 2 methods
		buf := make([]byte, 4)
		io.ReadFull(conn, buf)

		// Select USERNAME/PASSWORD
		conn.Write([]byte{0x05, 0x02})

		// Read auth: VER(1) + ULEN(1) + "u"(1) + PLEN(1) + "p"(1) = 5
		authBuf := make([]byte, 5)
		io.ReadFull(conn, authBuf)

		// Auth success
		conn.Write([]byte{0x01, 0x00})

		// Read UDP ASSOCIATE
		reqBuf := make([]byte, 10)
		io.ReadFull(conn, reqBuf)

		// Reply success
		reply := []byte{
			0x05, 0x00, 0x00,
			0x01,
			127, 0, 0, 1,
			0x27, 0x10, // port 10000
		}
		conn.Write(reply)

		buf = make([]byte, 1)
		conn.Read(buf)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	controlConn, relay, err := socks5UDPAssociate(ctx, proxyAddr, auth, nil)
	if err != nil {
		t.Fatalf("socks5UDPAssociate with auth error: %v", err)
	}
	defer controlConn.Close()

	if relay.Port != 10000 {
		t.Errorf("relay port = %d, want 10000", relay.Port)
	}
}

// --- socks5PacketConn tests ---

func TestSocks5PacketConn_WriteToReadFrom(t *testing.T) {
	// Create a mock "relay" UDP endpoint that echoes back with SOCKS5 headers
	relayConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("failed to create relay: %v", err)
	}
	defer relayConn.Close()

	relayAddr := relayConn.LocalAddr().(*net.UDPAddr)

	// Create a local UDP socket for the client side
	clientUDP, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("failed to create client UDP: %v", err)
	}

	// Use a net.Pipe() for the control connection (just needs to stay open)
	ctrlClient, ctrlServer := net.Pipe()
	defer ctrlServer.Close()

	pconn := &socks5PacketConn{
		udpConn:     clientUDP,
		controlConn: ctrlClient,
		relayAddr:   relayAddr,
	}
	defer pconn.Close()

	targetAddr := &net.UDPAddr{IP: net.IPv4(93, 184, 216, 34), Port: 443}

	// Start relay echo goroutine
	go func() {
		buf := make([]byte, 65535)
		n, senderAddr, err := relayConn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		// Echo back the same packet (with SOCKS5 header intact)
		relayConn.WriteToUDP(buf[:n], senderAddr)
	}()

	// Write through the SOCKS5 packet conn
	payload := []byte("test payload")
	n, err := pconn.WriteTo(payload, targetAddr)
	if err != nil {
		t.Fatalf("WriteTo error: %v", err)
	}
	if n != len(payload) {
		t.Errorf("WriteTo wrote %d bytes, want %d", n, len(payload))
	}

	// Read back
	readBuf := make([]byte, 1024)
	pconn.SetReadDeadline(time.Now().Add(5 * time.Second))
	rn, addr, err := pconn.ReadFrom(readBuf)
	if err != nil {
		t.Fatalf("ReadFrom error: %v", err)
	}

	if string(readBuf[:rn]) != "test payload" {
		t.Errorf("payload = %q, want %q", string(readBuf[:rn]), "test payload")
	}

	udpAddr, ok := addr.(*net.UDPAddr)
	if !ok {
		t.Fatalf("addr type = %T, want *net.UDPAddr", addr)
	}
	if !udpAddr.IP.Equal(targetAddr.IP) {
		t.Errorf("source IP = %v, want %v", udpAddr.IP, targetAddr.IP)
	}
	if udpAddr.Port != targetAddr.Port {
		t.Errorf("source port = %d, want %d", udpAddr.Port, targetAddr.Port)
	}
}

func TestSocks5PacketConn_CloseIdempotent(t *testing.T) {
	clientUDP, err := net.ListenUDP("udp", nil)
	if err != nil {
		t.Fatalf("failed to create UDP: %v", err)
	}

	ctrlClient, ctrlServer := net.Pipe()
	defer ctrlServer.Close()

	pconn := &socks5PacketConn{
		udpConn:     clientUDP,
		controlConn: ctrlClient,
		relayAddr:   &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1080},
	}

	// Close should work the first time
	if err := pconn.Close(); err != nil {
		t.Fatalf("first Close error: %v", err)
	}

	// Second close should be a no-op
	if err := pconn.Close(); err != nil {
		t.Fatalf("second Close should be no-op, got: %v", err)
	}
}

func TestSocks5PacketConn_LocalAddr(t *testing.T) {
	clientUDP, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("failed to create UDP: %v", err)
	}
	defer clientUDP.Close()

	ctrlClient, ctrlServer := net.Pipe()
	defer ctrlClient.Close()
	defer ctrlServer.Close()

	pconn := &socks5PacketConn{
		udpConn:     clientUDP,
		controlConn: ctrlClient,
		relayAddr:   &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1080},
	}

	localAddr := pconn.LocalAddr()
	if localAddr == nil {
		t.Fatal("LocalAddr returned nil")
	}

	udpLocalAddr, ok := localAddr.(*net.UDPAddr)
	if !ok {
		t.Fatalf("LocalAddr type = %T, want *net.UDPAddr", localAddr)
	}
	if udpLocalAddr.Port == 0 {
		t.Error("LocalAddr port should not be 0")
	}
}

func TestBuildHTTP3Transport_ProxySchemes(t *testing.T) {
	testCases := []struct {
		name      string
		proxyURL  string
		expectErr bool
	}{
		{name: "no proxy", proxyURL: "", expectErr: false},
		{name: "socks5", proxyURL: "socks5://127.0.0.1:1080", expectErr: false},
		{name: "socks5h with auth", proxyURL: "socks5h://user:pass@127.0.0.1:1080", expectErr: false},
		{name: "http is rejected", proxyURL: "http://127.0.0.1:8080", expectErr: true},
		{name: "https is rejected", proxyURL: "https://127.0.0.1:443", expectErr: true},
		{name: "socks4 is rejected", proxyURL: "socks4://127.0.0.1:1080", expectErr: true},
		{name: "unparseable is rejected", proxyURL: "http://[::1", expectErr: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			transport, err := buildHTTP3Transport(&http3Config{proxyURL: testCase.proxyURL})

			if testCase.expectErr {
				if err == nil {
					t.Fatalf("expected an error for proxy url %q, got none", testCase.proxyURL)
				}
				if transport != nil {
					t.Fatalf("expected a nil transport alongside the error, got %T", transport)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error for proxy url %q, got: %v", testCase.proxyURL, err)
			}
			if transport == nil {
				t.Fatal("expected a transport, got nil")
			}
		})
	}
}

// newTestPacketConn wires a socks5PacketConn to a relay endpoint on loopback and returns
// both, along with the relay's address.
func newTestPacketConn(t *testing.T) (*socks5PacketConn, *net.UDPConn, *net.UDPAddr) {
	t.Helper()

	relayConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("failed to create relay: %v", err)
	}
	t.Cleanup(func() { relayConn.Close() })

	clientUDP, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("failed to create client UDP: %v", err)
	}

	ctrlClient, ctrlServer := net.Pipe()
	t.Cleanup(func() { ctrlServer.Close() })

	pconn := &socks5PacketConn{
		udpConn:     clientUDP,
		controlConn: ctrlClient,
		relayAddr:   relayConn.LocalAddr().(*net.UDPAddr),
	}
	t.Cleanup(func() { pconn.Close() })

	return pconn, relayConn, clientUDP.LocalAddr().(*net.UDPAddr)
}

// A malformed datagram must not surface as a read error, because quic-go treats a read
// error from the packet conn as fatal and would tear down the whole connection.
func TestSocks5PacketConn_ReadFromSkipsMalformedDatagram(t *testing.T) {
	pconn, relayConn, clientAddr := newTestPacketConn(t)

	// Unsupported ATYP, so the header cannot be decoded.
	if _, err := relayConn.WriteToUDP([]byte{0x00, 0x00, 0x00, 0x09, 0xff}, clientAddr); err != nil {
		t.Fatalf("failed to send malformed datagram: %v", err)
	}

	// A well formed datagram behind it has to be delivered.
	good := append(encodeTestHeader(t, &net.UDPAddr{IP: net.IPv4(1, 2, 3, 4), Port: 443}), []byte("payload")...)
	if _, err := relayConn.WriteToUDP(good, clientAddr); err != nil {
		t.Fatalf("failed to send valid datagram: %v", err)
	}

	buf := make([]byte, 1024)
	pconn.SetReadDeadline(time.Now().Add(5 * time.Second))

	n, _, err := pconn.ReadFrom(buf)
	if err != nil {
		t.Fatalf("ReadFrom returned an error instead of skipping the malformed datagram: %v", err)
	}

	if got := string(buf[:n]); got != "payload" {
		t.Fatalf("payload = %q, want %q", got, "payload")
	}
}

// Datagrams from anywhere other than the relay must be dropped, the local socket is
// unconnected and would otherwise accept injected packets.
func TestSocks5PacketConn_ReadFromDropsForeignSource(t *testing.T) {
	pconn, relayConn, clientAddr := newTestPacketConn(t)

	foreign, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 2), Port: 0})
	if err != nil {
		t.Skipf("cannot bind a second loopback address: %v", err)
	}
	defer foreign.Close()

	injected := append(encodeTestHeader(t, &net.UDPAddr{IP: net.IPv4(1, 2, 3, 4), Port: 443}), []byte("injected")...)
	if _, err := foreign.WriteToUDP(injected, clientAddr); err != nil {
		t.Fatalf("failed to send injected datagram: %v", err)
	}

	legit := append(encodeTestHeader(t, &net.UDPAddr{IP: net.IPv4(1, 2, 3, 4), Port: 443}), []byte("legit")...)
	if _, err := relayConn.WriteToUDP(legit, clientAddr); err != nil {
		t.Fatalf("failed to send relay datagram: %v", err)
	}

	buf := make([]byte, 1024)
	pconn.SetReadDeadline(time.Now().Add(5 * time.Second))

	n, _, err := pconn.ReadFrom(buf)
	if err != nil {
		t.Fatalf("ReadFrom error: %v", err)
	}

	if got := string(buf[:n]); got != "legit" {
		t.Fatalf("payload = %q, want %q (injected datagram was not dropped)", got, "legit")
	}
}

func encodeTestHeader(t *testing.T, addr *net.UDPAddr) []byte {
	t.Helper()

	buf := make([]byte, maxSocks5UDPHeaderSize)

	return buf[:encodeSocks5UDPHeader(buf, addr)]
}

func TestDecodeSocks5UDPHeader_RejectsFragmented(t *testing.T) {
	data := []byte{0x00, 0x00, 0x01, 0x01, 1, 2, 3, 4, 0x01, 0xbb}

	if _, _, err := decodeSocks5UDPHeader(data); err == nil {
		t.Fatal("expected fragmented datagrams to be rejected")
	}
}

func TestDecodeSocks5UDPHeader_RejectsDomainName(t *testing.T) {
	data := []byte{0x00, 0x00, 0x00, 0x03, 0x0b}
	data = append(data, []byte("example.com")...)
	data = append(data, 0x01, 0xbb)

	if _, _, err := decodeSocks5UDPHeader(data); err == nil {
		t.Fatal("expected ATYP 0x03 to be rejected on the receive path")
	}
}

func TestSocks5UsernamePasswordAuth_RejectsOversizedCredentials(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	err := socks5UsernamePasswordAuth(client, &socks5Auth{user: strings.Repeat("a", 256), password: "pass"})
	if err == nil {
		t.Fatal("expected an error for a username longer than 255 bytes")
	}

	err = socks5UsernamePasswordAuth(client, &socks5Auth{user: "user", password: strings.Repeat("b", 256)})
	if err == nil {
		t.Fatal("expected an error for a password longer than 255 bytes")
	}
}

// --- end to end test against a real SOCKS5 proxy ---
//
// This needs a real SOCKS5 proxy that implements UDP ASSOCIATE, which many implementations
// do not. The proxy is configured per machine through the SOCKS_5_PROXY environment
// variable (see mise.toml, with the value in the uncommitted mise.local.toml), so the test
// skips when it is unset:
//
//	go test . -run Socks5E2E -v
//
// It lives here rather than next to the other SOCKS5 proxy tests in ./tests because it
// builds the HTTP/3 transport directly. The public API has no option to force HTTP/3, only
// protocol racing, and the HTTP/3 leg is the slower one when it has to take an extra proxy
// hop, so HTTP/2 usually wins the race and the code under test would never run.

const (
	// socks5ProxyEnvVar names the environment variable holding a
	// socks5://[user:pass@]host:port URL of a proxy that supports UDP ASSOCIATE.
	socks5ProxyEnvVar = "SOCKS_5_PROXY"

	// socks5E2EEndpoint reports back the address a request came from and the protocol it
	// was served over, which is exactly what this test needs to assert on. It is the same
	// endpoint the other HTTP/3 tests use, because it is reachable over QUIC, which
	// tls.peet.ws is not despite advertising h3.
	socks5E2EEndpoint = "https://www.cloudflare.com/cdn-cgi/trace"
)

// socks5E2EResponse is the subset of the endpoint's response this test cares about.
type socks5E2EResponse struct {
	IP          string
	HTTPVersion string
}

// socks5ProxyFromEnv returns the configured proxy URL, or skips the test when there is
// none. A bare host:port is accepted as well, since the scheme is only what selects the
// SOCKS5 code path.
func socks5ProxyFromEnv(t *testing.T) string {
	t.Helper()

	proxyURL := strings.TrimSpace(os.Getenv(socks5ProxyEnvVar))
	if proxyURL == "" {
		t.Skipf("%s is not set, skipping the SOCKS5 end to end test", socks5ProxyEnvVar)
	}

	if !strings.Contains(proxyURL, "://") {
		proxyURL = "socks5://" + proxyURL
	}

	if !strings.HasPrefix(proxyURL, "socks5://") && !strings.HasPrefix(proxyURL, "socks5h://") {
		t.Fatalf("%s has to be a socks5:// or socks5h:// URL, got %q", socks5ProxyEnvVar, proxyURL)
	}

	return proxyURL
}

// newSocks5E2ETransport builds the HTTP/3 transport exactly the way the library does for
// a request, so the test covers the production wiring rather than a hand rolled setup.
func newSocks5E2ETransport(t *testing.T, proxyURL string) *http3.Transport {
	t.Helper()

	clientProfile := profiles.Chrome_144

	transport, err := buildHTTP3Transport(&http3Config{
		http3Settings:          clientProfile.GetHttp3Settings(),
		http3SettingsOrder:     clientProfile.GetHttp3SettingsOrder(),
		http3PriorityParam:     clientProfile.GetHttp3PriorityParam(),
		http3PseudoHeaderOrder: clientProfile.GetHttp3PseudoHeaderOrder(),
		http3SendGreaseFrames:  clientProfile.GetHttp3SendGreaseFrames(),
		proxyURL:               proxyURL,
	})
	if err != nil {
		t.Fatalf("failed to build the HTTP/3 transport: %v", err)
	}

	h3Transport, ok := transport.(*http3.Transport)
	if !ok {
		t.Fatalf("expected a *http3.Transport, got %T", transport)
	}

	return h3Transport
}

// fetchSocks5E2E performs the request with the given round trip function and returns what
// the endpoint saw. It takes a function rather than an http.RoundTripper so that the
// baseline can go through a regular client.
func fetchSocks5E2E(t *testing.T, roundTrip func(*http.Request) (*http.Response, error)) socks5E2EResponse {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, socks5E2EEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header = http.Header{
		"accept":     {"*/*"},
		"user-agent": {"tls-client-socks5-e2e"},
	}

	resp, err := roundTrip(req)
	if err != nil {
		t.Fatalf("request to %s failed: %v", socks5E2EEndpoint, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read the response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d from %s: %s", resp.StatusCode, socks5E2EEndpoint, string(body))
	}

	return parseSocks5E2ETrace(t, string(body))
}

// parseSocks5E2ETrace reads the key=value lines the trace endpoint answers with.
func parseSocks5E2ETrace(t *testing.T, body string) socks5E2EResponse {
	t.Helper()

	var parsed socks5E2EResponse

	for _, line := range strings.Split(body, "\n") {
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		switch key {
		case "ip":
			parsed.IP = value
		case "http":
			parsed.HTTPVersion = value
		}
	}

	if parsed.IP == "" {
		t.Fatalf("the trace response did not contain an ip line: %q", body)
	}

	return parsed
}

// TestSocks5E2E_HTTP3ThroughProxy is the test that actually matters for this feature: it
// proves the QUIC traffic goes through the proxy. Without UDP ASSOCIATE support the TCP
// leg was proxied while the HTTP/3 leg went out directly, so the address the server saw
// over HTTP/3 was the real one.
func TestSocks5E2E_HTTP3ThroughProxy(t *testing.T) {
	proxyURL := socks5ProxyFromEnv(t)

	// Baseline: where do we come from without a proxy?
	// IPv4 only, so that the comparison below is between two addresses of the same family
	// rather than trivially different ones. The proxy relays over IPv4.
	baseline, err := NewHttpClient(nil,
		WithClientProfile(profiles.Chrome_144),
		WithTimeoutSeconds(30),
		WithDisableIPV6(),
	)
	if err != nil {
		t.Fatalf("failed to create the baseline client: %v", err)
	}

	direct := fetchSocks5E2E(t, baseline.Do)
	t.Logf("unproxied request came from %s over %s", direct.IP, direct.HTTPVersion)

	transport := newSocks5E2ETransport(t, proxyURL)
	defer transport.Close()

	proxied := fetchSocks5E2E(t, transport.RoundTrip)
	t.Logf("proxied request came from %s over %s", proxied.IP, proxied.HTTPVersion)

	// The endpoint spells it "http/3", other trace endpoints spell it "h3".
	if version := strings.ToLower(proxied.HTTPVersion); version != "h3" && version != "http/3" {
		t.Fatalf("expected the request to be served over HTTP/3, got %q", proxied.HTTPVersion)
	}

	if proxied.IP == direct.IP {
		t.Fatalf("the HTTP/3 request left from %s, the same address as an unproxied request, so the QUIC traffic bypassed the proxy", proxied.IP)
	}
}

// pickIPAddr has to honour the family of the relay, because an IPv4 only relay cannot
// forward to an IPv6 target and the resolver's own order is not tied to the proxy.
func TestPickIPAddr(t *testing.T) {
	v4 := net.IPAddr{IP: net.IPv4(93, 184, 216, 34)}
	v6 := net.IPAddr{IP: net.ParseIP("2606:2800:220:1:248:1893:25c8:1946")}

	relayV4 := net.IPv4(185, 111, 111, 40)
	relayV6 := net.ParseIP("2606:4700::1")

	tests := []struct {
		name   string
		ips    []net.IPAddr
		prefer net.IP
		want   net.IPAddr
	}{
		{name: "ipv4 relay picks the ipv4 address", ips: []net.IPAddr{v6, v4}, prefer: relayV4, want: v4},
		{name: "ipv6 relay picks the ipv6 address", ips: []net.IPAddr{v4, v6}, prefer: relayV6, want: v6},
		{name: "no preference keeps the resolver order", ips: []net.IPAddr{v6, v4}, prefer: nil, want: v6},
		{name: "falls back when no address matches", ips: []net.IPAddr{v6}, prefer: relayV4, want: v6},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := pickIPAddr(tc.ips, tc.prefer)
			if !got.IP.Equal(tc.want.IP) {
				t.Errorf("pickIPAddr = %v, want %v", got.IP, tc.want.IP)
			}
		})
	}
}

// A literal target address is used as is, no matter which family the relay speaks.
func TestResolveUDPAddr_LiteralAddress(t *testing.T) {
	addr, err := resolveUDPAddr(context.Background(), "93.184.216.34:443", net.ParseIP("2606:4700::1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !addr.IP.Equal(net.IPv4(93, 184, 216, 34)) || addr.Port != 443 {
		t.Errorf("resolveUDPAddr = %v, want 93.184.216.34:443", addr)
	}
}
