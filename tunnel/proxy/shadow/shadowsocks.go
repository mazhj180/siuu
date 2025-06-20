package shadow

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/shadowsocks/go-shadowsocks2/core"
	"net"
	"siuu/tunnel/proxy"
	"strconv"
	"time"
)

type p struct {
	proxy.BaseProxy

	name     string
	cipher   string
	password string
}

func New(base proxy.BaseProxy, name, cipher, password string) proxy.Proxy {
	return &p{
		BaseProxy: base,
		name:      name,
		cipher:    cipher,
		password:  password,
	}
}

func (s *p) Type() proxy.Type {
	return proxy.SHADOW
}

func (s *p) Connect(ctx context.Context, addr string, port uint16) (*proxy.Pd, error) {
	dialer := &net.Dialer{
		Timeout: 30 * time.Second,
	}
	agency, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(s.Server, strconv.FormatUint(uint64(s.Port), 10)))
	if err != nil {
		return nil, err
	}

	cipher, err := core.PickCipher(s.cipher, nil, s.password)
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

func (s *p) Name() string {
	return s.name
}

func (s *p) String() string {
	return fmt.Sprintf(
		`{"Server":"%s","Port":%d,"Protocol":"%s","Name":"%s","Cipher":"%s","Password":"%s","Type":"%s"}`,
		s.Server,
		s.Port,
		s.Protocol.String(),
		s.name,
		s.cipher,
		s.password,
		s.Type(),
	)
}
