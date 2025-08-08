package http

import (
	"errors"
	"io"
	"net"
	"time"
)

type c struct {
	io.Reader
	io.Writer
}

func (c *c) LocalAddr() net.Addr {
	return testAddr{}
}

func (c *c) RemoteAddr() net.Addr {
	return testAddr{}
}

func (c *c) SetDeadline(_ time.Time) error {
	return errors.New("it is not a real conn, it is just for testing ")
}

func (c *c) SetReadDeadline(_ time.Time) error {
	return errors.New("it is not a real conn, it is just for testing ")
}

func (c *c) SetWriteDeadline(_ time.Time) error {
	return errors.New("it is not a real conn, it is just for testing ")
}

func (c *c) Close() error {
	return nil
}

type testAddr struct{}

func (testAddr) Network() string {
	return "testing"
}

func (testAddr) String() string {
	return "there is not real connection, it's just for testing"
}
