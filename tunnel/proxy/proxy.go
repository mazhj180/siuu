package proxy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
)

const (
	DIRECT Type = iota
	REJECT
	HTTPS
	SOCKS
	SHADOW
	TROJAN
	VMESS

	// TCP it cannot forward UDP traffic.
	TCP Protocol = 1

	// UDP it can forward UDP traffic.
	UDP Protocol = 2
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

// Client forwarded data
type Client struct {
	Sid string // session id

	TrafficType Protocol // traffic type tcp or udp
	Conn        net.Conn // data stream or data packet

	Host string // dst host
	Port uint16 // dst port

	IsTLS bool // whether the traffic is tls encrypted.

	// store some data that needs to be written before the bidirectional copy.
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
	// ForwardTcp forwards the traffic of tcp according to the configuration.
	// It is used to handle the first connection and the first packet from the client.
	// It will block until the connection is closed.
	ForwardTcp(*Client) error
	ForwardUdp(*Client) (*UdpPocket, error)

	GetType() Type
	GetName() string
	GetServer() string
	GetPort() uint16
	GetProtocol() Protocol
}
