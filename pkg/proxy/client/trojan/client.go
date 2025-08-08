package trojan

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"siuu/pkg/proxy/client"
	"siuu/pkg/proxy/mux"
	"strconv"
	"time"
)

var _ client.ProxyClient = &trojan{}

type trojan struct {
	client.BaseClient

	name     string
	password string
	sni      string
}

func New(base client.BaseClient, name, password, sni string) client.ProxyClient {

	dialer := func(ctx context.Context) (net.Conn, error) {
		tlsConfig := &tls.Config{
			ServerName:         sni,
			InsecureSkipVerify: true,
		}
		dialer := &tls.Dialer{
			NetDialer: &net.Dialer{
				Timeout: 30 * time.Second,
			},
			Config: tlsConfig,
		}
		agency, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(base.Server, strconv.FormatUint(uint64(base.Port), 10)))
		if err != nil {
			return nil, err
		}
		return agency, nil
	}

	base.Dialer = dialer
	base.Pool = mux.NewPool(base.Mux, dialer, 5)

	return &trojan{
		BaseClient: base,
		name:       name,
		password:   password,
		sni:        sni,
	}
}

func (t *trojan) Connect(ctx context.Context, host string, port uint16) (net.Conn, error) {
	var agency net.Conn
	var err error

	if t.Mux != nil {
		agency, err = t.Pool.GetStream(ctx)
	} else {
		agency, err = t.Dialer(ctx)
	}

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

	ip := net.ParseIP(host)
	if ip4 := ip.To4(); ip4 != nil {
		atyp = 0x01
		addrBytes = ip4

	} else if ip6 := ip.To16(); ip6 != nil {
		atyp = 0x02
		addrBytes = ip6

	} else {
		atyp = 0x03
		addrLen = byte(len(host))
		addrBytes = append([]byte{addrLen}, []byte(host)...)

	}

	buf.WriteByte(atyp)
	buf.Write(addrBytes)

	binary.BigEndian.PutUint16(portBytes, port)
	buf.Write(portBytes)

	if _, err = agency.Write(append(buf.Bytes(), []byte{0x0d, 0x0a}...)); err != nil {
		return nil, err
	}

	return agency, nil
}

func (t *trojan) Name() string {
	return t.name
}

func (t *trojan) Type() string {
	return "trojan"
}

func (t *trojan) String() string {
	return fmt.Sprintf(
		`{"Server":"%s","Port":%d,"Protocol":"%s","Name":"%s","Type":"%s"}`,
		t.Server,
		t.Port,
		t.TrafficType,
		t.name,
		t.Type(),
	)
}
