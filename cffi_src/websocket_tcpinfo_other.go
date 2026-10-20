//go:build !linux

package tls_client_cffi_src

import "net"

// TCP_INFO is only read on linux.
func wsTcpInfo(net.Conn) *WsTcpInfo { return nil }
