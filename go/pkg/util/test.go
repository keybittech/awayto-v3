package util

import (
	"io"
	"net"
	"time"
)

type NullConn struct{}

func (n NullConn) Read(b []byte) (int, error)         { return 0, io.EOF }
func (n NullConn) Write(b []byte) (int, error)        { return len(b), nil }
func (n NullConn) Close() error                       { return nil }
func (n NullConn) LocalAddr() net.Addr                { return nil }
func (n NullConn) RemoteAddr() net.Addr               { return nil }
func (n NullConn) SetDeadline(t time.Time) error      { return nil }
func (n NullConn) SetReadDeadline(t time.Time) error  { return nil }
func (n NullConn) SetWriteDeadline(t time.Time) error { return nil }

// NewNullConn returns a new no-op connection
func NewNullConn() net.Conn {
	return NullConn{}
}
