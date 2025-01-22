package proto

import (
	"net"
	"siuu/tunnel/proxy"
)

const (
	HTTP Protocol = iota
	SOCKS
)

type Protocol byte

func (p Protocol) String() string {
	if p == HTTP {
		return "http"
	}
	return "socks"
}

type Interface interface {
	GetProtocol() Protocol
	GetProxy() proxy.Proxy
	GetConn() net.Conn
	GetHost() string
	GetPort() uint16
	ID() string
}

type HttpInterface interface {
	IsTLS() bool
	GetHttpReader() *proxy.HttpReader
}

type TrafficRecorder interface {
	RecordUp(int64, float64)
	RecordDown(int64, float64)
}
