package bandwidth

import (
	"net"
)

type BTConn struct {
	net.Conn
	tracker *Tracker
}

func (bt *BTConn) Read(p []byte) (n int, err error) {
	n, err = bt.Conn.Read(p)
	bt.tracker.addReadBytes(int64(n))
	return n, err
}

func (bt *BTConn) Write(p []byte) (n int, err error) {
	n, err = bt.Conn.Write(p)
	bt.tracker.addWriteBytes(int64(n))
	return n, err
}

func newTrackedConn(conn net.Conn, tracker *Tracker) *BTConn {
	return &BTConn{
		Conn:    conn,
		tracker: tracker,
	}
}
