package socks

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"siuu/tunnel/proxy"
	"strconv"
)

var ErrSocksVerNotSupported = errors.New("socks version not supported")
var ErrSocksAuthentication = errors.New("socks auth was fail or not support the way ")

type Proxy struct {
	Type     proxy.Type
	Name     string
	Server   string
	Port     uint16
	Username string
	Password string
	Protocol proxy.Protocol
}

func (s *Proxy) Connect(addr string, port uint16) (net.Conn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(s.Server, strconv.FormatUint(uint64(s.Port), 10)))
	if err != nil {
		return nil, err
	}

	agency, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	if err = agency.SetKeepAlive(true); err != nil {
		return nil, err
	}

	// VER=0x05, NMETHODS=1, METHODS=0x02
	if _, err = agency.Write([]byte{0x05, 0x01, 0x02}); err != nil {
		return nil, err
	}

	buf := make([]byte, 2)
	if _, err = io.ReadFull(agency, buf); err != nil {
		return nil, err
	}

	if buf[0] != 0x05 {
		return nil, ErrSocksVerNotSupported
	}

	if buf[1] != 0x02 {
		return nil, ErrSocksAuthentication
	}

	// Username/password authentication
	// Username Password Authentication Sub-Negotiation Protocol.
	// VER=0x01, ULEN=xx, USERNAME, PLEN=xx, PASSWORD
	uLen, pLen := byte(len(s.Username)), byte(len(s.Password))
	authMsg := make([]byte, uLen+pLen+3)
	authMsg[0] = 0x01 // ver

	authMsg[1] = uLen
	copy(authMsg[2:2+len(s.Username)], s.Username)

	authMsg[2+len(s.Username)] = pLen
	copy(authMsg[3+len(s.Username):], s.Password)

	if _, err = agency.Write(authMsg); err != nil {
		return nil, err
	}

	// Read authentication results
	// ver=0x01, status=0x00 success
	resp := make([]byte, 2)
	if _, err = io.ReadFull(agency, resp); err != nil {
		return nil, err
	}
	if resp[0] != 0x01 || resp[1] != 0x00 {
		return nil, ErrSocksAuthentication
	}

	// After successful authentication, send a CONNECT request to the upstream
	// formatting: VER=0x05, CMD=0x01, RSV=0x00, ATYP=?, DST.ADDR, DST.PORT
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
	connectReq := append([]byte{0x05, 0x01, 0x00, atyp}, append(addrBytes, portBytes...)...)
	if _, err = agency.Write(connectReq); err != nil {
		return nil, err
	}

	// Read upstream response
	// formatting: VER, REP, RSV, ATYP, BND.ADDR, BND.PORT
	resp = make([]byte, 4)
	if _, err = io.ReadFull(agency, resp); err != nil {
		return nil, err
	}
	if resp[0] != 0x05 {
		return nil, ErrSocksVerNotSupported
	}

	rep := resp[1]
	atyp = resp[3]

	// Read the server's binding information and determine if the response is valid.
	switch atyp {
	case 0x01:
		addrLen = 4

	case 0x03:
		domainLenByte := make([]byte, 1)
		if _, err = io.ReadFull(agency, domainLenByte); err != nil {
			return nil, err
		}
		addrLen = domainLenByte[0]

	case 0x04:
		addrLen = 16

	default:
		return nil, proxy.ErrProxyResp
	}

	bnd := make([]byte, addrLen+2)
	if _, err = io.ReadFull(agency, bnd); err != nil {
		return nil, err
	}

	// According to rep, judge whether the upstream connection is successful
	if rep != 0x00 {
		return nil, proxy.ErrProxyResp
	}

	return agency, nil
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
