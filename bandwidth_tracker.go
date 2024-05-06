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
	bt.tracker.AddReadBytes(int64(n))
	return n, err
}

func (bt *btConn) Write(p []byte) (n int, err error) {
	n, err = bt.Conn.Write(p)
	bt.tracker.AddWriteBytes(int64(n))
	return n, err
}

func newBandwidthTrackedConn(conn net.Conn, tracker *bandwidthTracker) *btConn {
	return &btConn{
		Conn:    conn,
		tracker: tracker,
	}
}

type bandwidthTracker struct {
	writeBytes atomic.Int64
	readBytes  atomic.Int64
}

func (bt *bandwidthTracker) AddWriteBytes(n int64) {
	bt.writeBytes.Add(n)
}

func (bt *bandwidthTracker) AddReadBytes(n int64) {
	bt.readBytes.Add(n)
}

func (bt *bandwidthTracker) GetWriteBytes() int64 {
	return bt.writeBytes.Load()
}

func (bt *bandwidthTracker) GetReadBytes() int64 {
	return bt.readBytes.Load()
}

func newBandwidthTracker() *bandwidthTracker {
	return &bandwidthTracker{}
}
