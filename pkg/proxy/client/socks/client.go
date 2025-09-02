package socks

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"siuu/pkg/proxy/client"
	"strconv"
	"time"
)

var _ client.ProxyClient = &socks{}

var ErrSocksVerNotSupported = errors.New("socks version not supported")
var ErrSocksAuthentication = errors.New("socks auth was fail or not support the way ")

type socks struct {
	client.BaseClient

	name     string
	username string
	password string
}

func New(base client.BaseClient, name, username, password string) client.ProxyClient {
	return &socks{
		BaseClient: base,
		name:       name,
		username:   username,
		password:   password,
	}
}

func (s *socks) Type() string {
	return "socks"
}

func (s *socks) Connect(ctx context.Context, proto, addr string, port uint16) (net.Conn, error) {
	switch proto {
	case "tcp":
		return s.connectTcp(ctx, addr, port)
	case "udp":
		return s.connectUdp(ctx, addr, port)
	default:
		return nil, client.ErrProxyResp
	}
}

func (s *socks) connectTcp(ctx context.Context, addr string, port uint16) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: 30 * time.Second,
	}
	agency, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(s.Server, strconv.FormatUint(uint64(s.Port), 10)))
	if err != nil {
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
	uLen, pLen := byte(len(s.username)), byte(len(s.password))
	authMsg := make([]byte, uLen+pLen+3)
	authMsg[0] = 0x01 // ver

	authMsg[1] = uLen
	copy(authMsg[2:2+len(s.username)], s.username)

	authMsg[2+len(s.username)] = pLen
	copy(authMsg[3+len(s.username):], s.password)

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
		return nil, client.ErrProxyResp
	}

	bnd := make([]byte, addrLen+2)
	if _, err = io.ReadFull(agency, bnd); err != nil {
		return nil, err
	}

	// According to rep, judge whether the upstream connection is successful
	if rep != 0x00 {
		return nil, client.ErrProxyResp
	}

	return agency, nil
}

func (s *socks) connectUdp(ctx context.Context, addr string, port uint16) (net.Conn, error) {
	return nil, client.ErrProxyResp
}

func (s *socks) Name() string {
	return s.name
}

func (s *socks) String() string {
	return fmt.Sprintf(
		`{"Server":"%s","Port":%d,"Protocol":"%s","Name":"%s","Username":"%s","Password":"%s","Type":"%s"}`,
		s.Server,
		s.Port,
		s.TrafficType,
		s.name,
		s.username,
		s.password,
		s.Type(),
	)
}
