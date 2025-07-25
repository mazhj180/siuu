package proxy

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"siuu/tunnel/mux"
)

const (
	DIRECT Type = iota
	REJECT
	HTTPS
	SOCKS
	SHADOW
	TROJAN
	VMESS

	// TCP it can forward TCP traffic.
	TCP Protocol = 1

	// UDP it can forward UDP traffic.
	UDP Protocol = 2

	BOTH Protocol = 3
)

var ErrProxyTypeNotSupported = errors.New("proxy type not supported")
var ErrProtocolNotSupported = errors.New("protocol not supported")
var ErrProxyResp = errors.New("proxy response error")

type Type int

func (t *Type) MarshalJSON() ([]byte, error) {
	var typ string
	switch *t {
	case DIRECT:
		typ = "direct"
	case REJECT:
		typ = "reject"
	case HTTPS:
		typ = "https"
	case SOCKS:
		typ = "socks"
	case SHADOW:
		typ = "shadow"
	case TROJAN:
		typ = "trojan"
	default:
		return nil, fmt.Errorf("%w: %d", ErrProxyTypeNotSupported, t)
	}
	return json.Marshal(typ)
}

func (t *Type) UnmarshalJSON(data []byte) error {
	var typ string
	if err := json.Unmarshal(data, &typ); err != nil {
		return err
	}
	switch typ {
	case "direct":
		*t = DIRECT
	case "reject":
		*t = REJECT
	case "https":
		*t = HTTPS
	case "socks":
		*t = SOCKS
	case "shadow":
		*t = SHADOW
	case "trojan":
		*t = TROJAN
	default:
		return fmt.Errorf("%w: %s", ErrProxyTypeNotSupported, typ)
	}
	return nil
}

func (t Type) String() string {
	switch t {
	case DIRECT:
		return "direct"
	case REJECT:
		return "reject"
	case HTTPS:
		return "https"
	case SOCKS:
		return "socks"
	case SHADOW:
		return "shadow"
	case TROJAN:
		return "trojan"
	default:
		return ""
	}
}

// Protocol : the types of traffic that support forwarding
type Protocol byte

func (p *Protocol) MarshalJSON() ([]byte, error) {
	var proto string
	switch *p {
	case TCP:
		proto = "tcp"
	case UDP:
		proto = "udp"
	case BOTH:
		proto = "both"
	default:
		return nil, fmt.Errorf("%w: %d", ErrProtocolNotSupported, p)
	}
	return json.Marshal(proto)
}

func (p *Protocol) UnmarshalJSON(data []byte) error {
	var proto string
	if err := json.Unmarshal(data, &proto); err != nil {
		return err
	}
	switch proto {
	case "tcp":
		*p = TCP
	case "udp":
		*p = UDP
	default:
		return fmt.Errorf("%w: %s", ErrProtocolNotSupported, proto)
	}
	return nil
}

func (p *Protocol) String() string {
	switch *p {
	case TCP:
		return "tcp"
	case UDP:
		return "udp"
	case BOTH:
		return "both"
	default:
		return "unknown"
	}
}

// Client forwarded data
type Client struct {
	Sid string // session id

	TrafficType Protocol // traffic type tcp or udp
	Conn        net.Conn // data stream or data packet

	Host string // dst host
	Port uint16 // dst port

	IsTLS bool // whether the traffic is tls encrypted.

	// proxies some data that needs to be written before the bidirectional copy.
	// example: the data that has been read out from the connection in http request
	Req *HttpReader
}

// UdpPocket udp data
type UdpPocket struct {
	Addr *net.UDPAddr
	bytes.Buffer
}

// Proxy the abstraction of proxy
type Proxy interface {
	fmt.Stringer

	// Connect connects to the proxy with the given host and port.
	// It returns the connection if success, otherwise returns an error.
	Connect(context.Context, string, uint16) (*Pd, error)

	Type() Type
	Name() string

	GetServer() string
	GetPort() uint16
	GetProtocol() Protocol
}

type BaseProxy struct {
	Server   string
	Port     uint16
	Protocol Protocol
	Mux      mux.Interface
	Pool     *mux.Pool
	Dialer   func(context.Context) (net.Conn, error)
}

func (p *BaseProxy) GetServer() string {
	return p.Server
}

func (p *BaseProxy) GetPort() uint16 {
	return p.Port
}

func (p *BaseProxy) GetProtocol() Protocol {
	return p.Protocol
}

type Pd struct {
	conn net.Conn
}

func (p *Pd) Read(b []byte) (n int, err error) {
	return p.conn.Read(b)
}

func (p *Pd) Write(b []byte) (n int, err error) {
	return p.conn.Write(b)
}

func (p *Pd) Close() error {
	return p.conn.Close()
}

func (p *Pd) CloseWriter() error {
	switch c := p.conn.(type) {
	case *net.TCPConn:
		return c.CloseWrite()
	case *tls.Conn:
		return c.CloseWrite()
	default:
		return nil
	}
}

func NewPd(conn net.Conn) *Pd {
	return &Pd{conn: conn}
}
