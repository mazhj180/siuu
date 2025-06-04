package vmess

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/tls"
	"encoding/binary"
	"net"
	"siuu/tunnel/proxy"
	"strconv"
	"time"
)

type p struct {
	proxy.BaseProxy

	name    string
	uuid    string
	alterId int
	cipher  string
}

func New(base proxy.BaseProxy, name, uuid string, alterId int, cipher string) proxy.Proxy {
	return &p{
		BaseProxy: base,
		name:      name,
		uuid:      uuid,
		alterId:   alterId,
		cipher:    cipher,
	}
}

func (v *p) Type() proxy.Type {
	return proxy.VMESS
}

func (v *p) Connect(ctx context.Context, addr string, port uint16) (*proxy.Pd, error) {
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
	mac := hmac.New(md5.New, []byte(v.uuid))
	mac.Write(ts[:])
	copy(authId[:], mac.Sum(nil))

	// dst addr
	// todo

	return proxy.NewPd(agency), proxy.ErrProtocolNotSupported
}

func (v *p) Name() string {
	return v.name
}

func (v *p) String() string {
	return ""
}
