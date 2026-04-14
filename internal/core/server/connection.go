package server

import "net"

// ConnMode represents the operating mode of a client connection.
type ConnMode int

const (
	ModeNormal     ConnMode = iota
	ModeSubscriber          // locked to SUBSCRIBE/UNSUBSCRIBE/PING/QUIT only
)

// Conn holds all per-connection state.
type Conn struct {
	net     net.Conn
	dbIndex int
	authed  bool
	mode    ConnMode
	// writeCh is not stored here — the broker owns the delivery channel per conn.
	// The write goroutine is started by SetMode and drains broker.GetConnChan(conn).
}

func newConn(nc net.Conn) *Conn {
	return &Conn{
		net:     nc,
		dbIndex: 0,
		authed:  false,
		mode:    ModeNormal,
	}
}
