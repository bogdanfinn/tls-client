package tls_client_cffi_src

import (
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

// wsTcpInfo unwraps TLS (NetConn) down to the raw socket and reads TCP_INFO. nil if the
// connection is not a plain TCP socket underneath (e.g. tunnelled through an h2 proxy).
func wsTcpInfo(conn net.Conn) *WsTcpInfo {
	for depth := 0; depth < 8 && conn != nil; depth++ {
		if sc, ok := conn.(syscall.Conn); ok {
			raw, err := sc.SyscallConn()
			if err != nil {
				return nil
			}

			var info *unix.TCPInfo
			var infoErr error
			if err := raw.Control(func(fd uintptr) {
				info, infoErr = unix.GetsockoptTCPInfo(int(fd), unix.IPPROTO_TCP, unix.TCP_INFO)
			}); err != nil || infoErr != nil || info == nil {
				return nil
			}

			return &WsTcpInfo{
				State:               info.State,
				Retransmits:         info.Retransmits,
				Backoff:             info.Backoff,
				RtoMs:               info.Rto / 1000,
				RttMs:               info.Rtt / 1000,
				Unacked:             info.Unacked,
				Lost:                info.Lost,
				TotalRetrans:        info.Total_retrans,
				NotsentBytes:        info.Notsent_bytes,
				MsSinceLastDataSent: info.Last_data_sent,
				MsSinceLastDataRecv: info.Last_data_recv,
				MsSinceLastAckRecv:  info.Last_ack_recv,
			}
		}

		nc, ok := conn.(interface{ NetConn() net.Conn })
		if !ok {
			return nil
		}
		conn = nc.NetConn()
	}

	return nil
}
