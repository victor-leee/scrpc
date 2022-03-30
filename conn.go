package scrpc

import (
	"net"
	"time"
)

const (
	ConnTypeSideCar2Local   = "side_car_to_local"
	ConnTypeSideCar2SideCar = "side_car_to_side_car"
)

type ConnOpt func(conn *Conn)

func WithType(connType string) ConnOpt {
	return func(conn *Conn) {
		conn.Type = connType
	}
}

// Conn is a wrapper to net.Conn
type Conn struct {
	NetConn net.Conn
	Type    string
}

func Dial(network, address string, opts ...ConnOpt) (*Conn, error) {
	netConn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}

	conn := &Conn{
		NetConn: netConn,
	}
	for _, opt := range opts {
		opt(conn)
	}

	return conn, nil
}

func (c *Conn) Read(b []byte) (n int, err error) {
	return c.NetConn.Read(b)
}

func (c *Conn) Write(b []byte) (n int, err error) {
	return c.NetConn.Write(b)
}

func (c *Conn) Close() error {
	return c.NetConn.Close()
}

func (c *Conn) LocalAddr() net.Addr {
	return c.NetConn.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.NetConn.RemoteAddr()
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.NetConn.SetDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.NetConn.SetReadDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.NetConn.SetWriteDeadline(t)
}

type Listener struct {
	Listener net.Listener
	Type     string
}

func (l *Listener) Accept() (*Conn, error) {
	netConn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return &Conn{
		NetConn: netConn,
		Type:    l.Type,
	}, nil
}

func (l *Listener) Close() error {
	return l.Listener.Close()
}

func (l *Listener) Addr() net.Addr {
	return l.Listener.Addr()
}
