package shadow

import (
	"context"
	"encoding/binary"
	"github.com/shadowsocks/go-shadowsocks2/core"
	"net"
	"siuu/tunnel/proxy"
	"strconv"
	"time"
)

type Proxy struct {
	Type     proxy.Type
	Name     string
	Server   string
	Port     uint16
	Cipher   string
	Password string
	Protocol proxy.Protocol
}

func (s *Proxy) Connect(ctx context.Context, addr string, port uint16) (*proxy.Pd, error) {
	dialer := &net.Dialer{
		Timeout: 30 * time.Second,
	}
	agency, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(s.Server, strconv.FormatUint(uint64(s.Port), 10)))
	if err != nil {
		return nil, err
	}

	cipher, err := core.PickCipher(s.Cipher, nil, s.Password)
	if err != nil {
		return nil, err
	}
	agency = cipher.StreamConn(agency).(*net.TCPConn)

	var atyp byte
	var addrBytes []byte
	var addrLen byte
	portBytes := make([]byte, 2)

	binary.BigEndian.PutUint16(portBytes, port)
	ip := net.ParseIP(addr)

	if ip4 := ip.To4(); ip4 != nil {
		atyp = 0x01
		addrBytes = ip4

	} else if ip6 := ip.To16(); ip6 != nil {
		atyp = 0x02
		addrBytes = ip6

	} else {
		atyp = 0x03
		addrLen = byte(len(addr))
		addrBytes = append([]byte{addrLen}, []byte(addr)...)

	}

	req := make([]byte, 0, 1+len(addrBytes)+2)
	req = append(req, atyp)
	req = append(req, addrBytes...)
	req = append(req, portBytes...)

	if _, err = agency.Write(req); err != nil {
		return nil, err
	}
	return proxy.NewPd(agency), nil
}

func (s *Proxy) GetName() string {
	return s.Name
}

func (s *Proxy) GetType() proxy.Type {
	return s.Type
}

func (s *Proxy) GetServer() string {
	return s.Server
}

func (s *Proxy) GetPort() uint16 {
	return s.Port
}

func (s *Proxy) GetProtocol() proxy.Protocol {
	return s.Protocol
}
