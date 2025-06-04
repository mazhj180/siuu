package torjan

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"siuu/tunnel/proxy"
	"strconv"
	"time"
)

type p struct {
	proxy.BaseProxy

	name     string
	password string
	sni      string
}

func New(base proxy.BaseProxy, name, password, sni string) proxy.Proxy {
	return &p{
		BaseProxy: base,
		name:      name,
		password:  password,
		sni:       sni,
	}
}

func (t *p) Type() proxy.Type {
	return proxy.TROJAN
}

func (t *p) Connect(ctx context.Context, addr string, port uint16) (*proxy.Pd, error) {
	tlsConfig := &tls.Config{
		ServerName:         t.sni,
		InsecureSkipVerify: true,
	}
	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{
			Timeout: 30 * time.Second,
		},
		Config: tlsConfig,
	}
	agency, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(t.Server, strconv.FormatUint(uint64(t.Port), 10)))
	if err != nil {
		return nil, err
	}

	// trojan handshake
	hash := sha256.New224()
	hash.Write([]byte(t.password))
	pwd := hex.EncodeToString(hash.Sum(nil))
	authMsg := pwd + "\r\n"
	if _, err = agency.Write([]byte(authMsg)); err != nil {
		return nil, err
	}

	// send connect req
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(0x01)

	var atyp byte
	var addrBytes []byte
	var addrLen byte
	var portBytes = make([]byte, 2)

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

	buf.WriteByte(atyp)
	buf.Write(addrBytes)

	binary.BigEndian.PutUint16(portBytes, port)
	buf.Write(portBytes)

	if _, err = agency.Write(append(buf.Bytes(), []byte{0x0d, 0x0a}...)); err != nil {
		return nil, err
	}

	return proxy.NewPd(agency), nil
}

func (t *p) Name() string {
	return t.name
}

func (t *p) String() string {
	return fmt.Sprintf(
		`{"Server":"%s","Port":%d,"Protocol":"%s","Name":"%s","Sni":"%s","Password":"%s","Type":"%s"}`,
		t.Server,
		t.Port,
		t.Protocol,
		t.name,
		t.sni,
		t.password,
		t.Type(),
	)
}
