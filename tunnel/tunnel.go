package tunnel

import (
	"evil-gopher/proxy"
	"net"
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
	GetPort() int
	ID() string
}

type HttpInterface interface {
	IsTLS() bool
}

var (
	inCh   chan Interface
	outCh  chan Interface
	buffer []Interface
	T      Tunnel
)

func init() {
	initInfiniteCh()
	initTunnel()
	go dispatch()
}

type Tunnel interface {
	In(Interface)
	Out() (Interface, bool)
}

type tunnel struct {
}

func initTunnel() {
	T = new(tunnel)
}

func (t *tunnel) In(v Interface) {
	inCh <- v
}

func (t *tunnel) Out() (Interface, bool) {
	v, ok := <-outCh
	return v, ok
}

func initInfiniteCh() {
	buffer = make([]Interface, 0, 10)
	inCh = make(chan Interface, 10)
	outCh = make(chan Interface, 1)

	go func() {
		for {
			if len(buffer) > 0 {
				select {
				case outCh <- buffer[0]:
					buffer = buffer[1:]
				case v := <-inCh:
					buffer = append(buffer, v)
				}
			} else {
				v := <-inCh
				buffer = append(buffer, v)
			}
		}
	}()
}
