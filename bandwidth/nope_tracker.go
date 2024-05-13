package bandwidth

import (
	"net"
)

type NopeTracker struct {
}

func (bt *NopeTracker) GetWriteBytes() int64 {
	return 0
}

func (bt *NopeTracker) GetReadBytes() int64 {
	return 0
}

func (bt *NopeTracker) GetTotalBandwidth() int64 {
	return 0
}

func (bt *NopeTracker) TrackConnection(conn net.Conn) net.Conn {
	return conn
}

func NewNopeTracker() *NopeTracker {
	return &NopeTracker{}
}

var _ BandwidthTracker = (*NopeTracker)(nil)
