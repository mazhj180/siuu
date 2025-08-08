package mux

import (
	"errors"
	"io"
	"net"
	"strings"
)

var multiplexers map[string]func() Interface

func init() {
	multiplexers = make(map[string]func() Interface)
}

func Register(key string, getter func() Interface) {
	if strings.EqualFold(key, "no") {
		panic("mux: Register key must not be 'no'")
	}
	multiplexers[key] = getter
}

func GetMultiplexer(key string) (Interface, error) {
	if v, ok := multiplexers[key]; ok {
		return v(), nil
	}
	return nil, errors.New("undefined multiplexer")
}

type Interface interface {
	Name() string
	Client(conn net.Conn) (Session, error)
	Server() (Session, error)
}

type Session interface {
	io.Closer

	OpenStream() (Stream, error)
	AcceptStream() (Stream, error)
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	IsClosed() bool
	NumStreams() int
}

type Stream interface {
	net.Conn
	CloseWriter() error
	CloseReader() error
}

type FataError struct {
	error
}

func FataErr(err error) *FataError {
	return &FataError{err}
}
