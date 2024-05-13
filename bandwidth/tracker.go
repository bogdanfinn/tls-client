package bandwidth

import (
	"net"
	"sync/atomic"
)

type BandwidthTracker interface {
	GetTotalBandwidth() int64
	GetWriteBytes() int64
	GetReadBytes() int64
	TrackConnection(conn net.Conn) net.Conn
}

type Tracker struct {
	writeBytes atomic.Int64
	readBytes  atomic.Int64
}

func (bt *Tracker) GetWriteBytes() int64 {
	return bt.writeBytes.Load()
}

func (bt *Tracker) GetReadBytes() int64 {
	return bt.readBytes.Load()
}

func (bt *Tracker) GetTotalBandwidth() int64 {
	return bt.readBytes.Load() + bt.writeBytes.Load()
}

func (bt *Tracker) TrackConnection(conn net.Conn) net.Conn {
	return newTrackedConn(conn, bt)
}

func (bt *Tracker) addWriteBytes(n int64) {
	bt.writeBytes.Add(n)
}

func (bt *Tracker) addReadBytes(n int64) {
	bt.readBytes.Add(n)
}

func NewTracker() *Tracker {
	return &Tracker{}
}

var _ BandwidthTracker = (*Tracker)(nil)
