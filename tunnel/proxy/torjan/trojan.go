package torjan

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"net"
	"siuu/tunnel/proxy"
	"strconv"
)

type Proxy struct {
	conn net.Conn

	Type     proxy.Type
	Name     string
	Server   string
	Port     uint16
	Password string
	Protocol proxy.Protocol
	Sni      string
}

func (t *Proxy) Connect(addr string, port uint16) (net.Conn, error) {
	tlsConfig := &tls.Config{
		ServerName:         t.Sni,
		InsecureSkipVerify: true,
	}

	agency, err := tls.Dial("tcp", net.JoinHostPort(t.Server, strconv.FormatUint(uint64(t.Port), 10)), tlsConfig)
	if err != nil {
		return nil, err
	}

	// trojan handshake
	hash := sha256.New224()
	hash.Write([]byte(t.Password))
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
	t.conn = agency

	return agency, nil
}

func (t *Proxy) GetName() string {
	return t.Name
}

func (t *Proxy) GetType() proxy.Type {
	return t.Type
}

func (t *Proxy) GetServer() string {
	return t.Server
}

func (t *Proxy) GetPort() uint16 {
	return t.Port
}

func (t *Proxy) GetProtocol() proxy.Protocol {
	return t.Protocol
}
