package tls_client

import (
	"net"
	"sync/atomic"
)

type btConn struct {
	net.Conn
	tracker *bandwidthTracker
}

func (bt *btConn) Read(p []byte) (n int, err error) {
	n, err = bt.Conn.Read(p)
	bt.tracker.addReadBytes(int64(n))
	return n, err
}

func (bt *btConn) Write(p []byte) (n int, err error) {
	n, err = bt.Conn.Write(p)
	bt.tracker.addWriteBytes(int64(n))
	return n, err
}

func newBandwidthTrackedConn(conn net.Conn, tracker *bandwidthTracker) *btConn {
	return &btConn{
		Conn:    conn,
		tracker: tracker,
	}
}

type BandwidthTracker interface {
	GetTotalBandwidth() int64
	GetWriteBytes() int64
	GetReadBytes() int64
}

type bandwidthTracker struct {
	writeBytes atomic.Int64
	readBytes  atomic.Int64
}

func (bt *bandwidthTracker) GetWriteBytes() int64 {
	return bt.writeBytes.Load()
}

func (bt *bandwidthTracker) GetReadBytes() int64 {
	return bt.readBytes.Load()
}

func (bt *bandwidthTracker) GetTotalBandwidth() int64 {
	return bt.readBytes.Load() + bt.writeBytes.Load()
}

func (bt *bandwidthTracker) addWriteBytes(n int64) {
	bt.writeBytes.Add(n)
}

func (bt *bandwidthTracker) addReadBytes(n int64) {
	bt.readBytes.Add(n)
}

func newBandwidthTracker() *bandwidthTracker {
	return &bandwidthTracker{}
}

var _ BandwidthTracker = (*bandwidthTracker)(nil)
