package vmess

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/tls"
	"encoding/binary"
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
	Uuid     string
	AlterId  int
	Cipher   string
	Protocol proxy.Protocol
}

func (v *Proxy) Connect(addr string, port uint16) (net.Conn, error) {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	agency, err := tls.Dial("tcp", net.JoinHostPort(v.Server, strconv.FormatUint(uint64(v.Port), 10)), conf)
	if err != nil {
		return nil, err
	}

	// generate timestamp
	now := time.Now().Unix()
	var ts [8]byte
	binary.BigEndian.PutUint64(ts[:], uint64(now))

	// generate auth id
	var authId [16]byte
	mac := hmac.New(md5.New, []byte(v.Uuid))
	mac.Write(ts[:])
	copy(authId[:], mac.Sum(nil))

	// dst addr
	// todo

	return agency, proxy.ErrProtocolNotSupported
}

func (v *Proxy) GetName() string {
	return v.Name
}

func (v *Proxy) GetType() proxy.Type {
	return v.Type
}

func (v *Proxy) GetServer() string {
	return v.Server
}

func (v *Proxy) GetPort() uint16 {
	return v.Port
}

func (v *Proxy) GetProtocol() proxy.Protocol {
	return v.Protocol
}
