package tester

import (
	"errors"
	"io"
	"net"
	"time"
)

var NotRealConnError = errors.New("it is not a real conn, it is just for testing ")

type TestConn struct {
	io.Reader
	io.Writer
}

func (t *TestConn) LocalAddr() net.Addr {
	return TestAddr{}
}

func (t *TestConn) RemoteAddr() net.Addr {
	return TestAddr{}
}

func (t *TestConn) SetDeadline(_ time.Time) error {
	return NotRealConnError
}

func (t *TestConn) SetReadDeadline(_ time.Time) error {
	return NotRealConnError
}

func (t *TestConn) SetWriteDeadline(_ time.Time) error {
	return NotRealConnError
}

func (t *TestConn) Close() error {
	return nil
}

type TestAddr struct{}

func (TestAddr) Network() string {
	return "testing"
}

func (TestAddr) String() string {
	return "there is not real connection, it's just for testing"
}
