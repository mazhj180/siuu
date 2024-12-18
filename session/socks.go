package session

import (
	"encoding/binary"
	"evil-gopher/proxy"
	"evil-gopher/routing"
	"evil-gopher/tunnel"
	"fmt"
	"net"
)

type socksSession struct {
	id   string
	conn net.Conn
	prx  proxy.Proxy
	host *routing.TargetHost
}

func OpenSocksSession(conn net.Conn) Session {
	sid := "s-" + genSid()
	host := &routing.TargetHost{}
	return &socksSession{
		id:   sid,
		conn: conn,
		host: host,
	}
}

func (s *socksSession) ID() string {
	return s.id
}

func (s *socksSession) Handshakes() (*routing.TargetHost, error) {

	buf := make([]byte, 262)
	n, err := s.conn.Read(buf)
	if err != nil {
		return nil, err
	}
	if ver, nmethods := buf[0], int(buf[1]); ver != 0x05 || n < nmethods+2 {
		return nil, fmt.Errorf("invalid version")
	}

	if _, err = s.conn.Write([]byte{0x05, 0x00}); err != nil {
		return nil, err
	}

	n, err = s.conn.Read(buf)
	if err != nil {
		return nil, err
	}

	if n < 7 {
		return nil, fmt.Errorf("invalid command")
	}

	ver := buf[0]
	cmd := buf[1]
	rsv := buf[2]
	atyp := buf[3]

	if ver != 0x05 || rsv != 0x00 || cmd != 0x01 {
		_, _ = s.conn.Write([]byte{0x05, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return nil, fmt.Errorf("invalid version")
	}

	idx := 4

	switch atyp {
	case 0x01:
		if n < idx+6 {
			return nil, fmt.Errorf("invalid ipv4 and port")
		}
		s.host.IP = buf[idx : idx+4]
		idx += 4
		s.host.Port = int(binary.BigEndian.Uint16(buf[idx : idx+2]))

	case 0x03:
		if n < idx+1 {
			return nil, fmt.Errorf("invalid domain and port")
		}
		domainLen := int(buf[idx])
		idx += 1
		if n < idx+domainLen+2 {
			return nil, fmt.Errorf("invalid domain and port")
		}
		domain := string(buf[idx : idx+domainLen])
		s.host.Domain = domain
		idx += domainLen
		s.host.Port = int(binary.BigEndian.Uint16(buf[idx : idx+2]))

	case 0x04:
		if n < idx+18 {
			return nil, fmt.Errorf("invalid ipv6 and	port")
		}
		s.host.IP = buf[idx : idx+16]
		idx += 16
		s.host.Port = int(binary.BigEndian.Uint16(buf[idx : idx+2]))

	default:
		_, _ = s.conn.Write([]byte{0x05, 0x08, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return nil, fmt.Errorf("invalid command")
	}

	if _, err = s.conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0}); err != nil {
		return nil, err
	}

	return s.host, nil
}

func (s *socksSession) String() string {
	return s.id
}

func (s *socksSession) GetHost() string {
	if s.host.Domain != "" {
		return s.host.Domain
	}
	return s.host.IP.String()
}

func (s *socksSession) GetPort() int {
	return s.host.Port
}

func (s *socksSession) SetProxy(p proxy.Proxy) {
	s.prx = p
}

func (s *socksSession) GetProxy() proxy.Proxy {
	return s.prx
}

func (s *socksSession) GetConn() net.Conn {
	return s.conn
}

func (s *socksSession) GetProtocol() tunnel.Protocol {
	return tunnel.SOCKS
}
