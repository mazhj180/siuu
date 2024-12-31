package proxy

import (
	"encoding/binary"
	"encoding/json"
	"github.com/shadowsocks/go-shadowsocks2/core"
	"io"
	"net"
	"siuu/logger"
	"strconv"
)

type ShadowSocksProxy struct {
	Type     Type
	Name     string
	Server   string
	Port     uint16
	Cipher   string
	Password string
	Protocol Protocol
}

func (s *ShadowSocksProxy) String() string {
	jbytes, err := json.Marshal(s)
	if err != nil {
		return ""
	}
	return string(jbytes)
}

func (s *ShadowSocksProxy) Act(client *Client) error {

	if s.Protocol == TCP {
		return s.actOfTcp(client)
	}

	return ErrProtocolNotSupported
}

func (s *ShadowSocksProxy) actOfTcp(client *Client) error {

	conn := client.Conn
	defer conn.Close()

	agency, err := net.Dial("tcp", net.JoinHostPort(s.Server, strconv.FormatUint(uint64(s.Port), 10)))
	if err != nil {
		return err
	}
	defer agency.Close()

	cipher, err := core.PickCipher(s.Cipher, nil, s.Password)
	if err != nil {
		return err
	}
	agency = cipher.StreamConn(agency)

	var atyp byte
	var addrBytes []byte
	var addrLen byte
	portBytes := make([]byte, 2)

	binary.BigEndian.PutUint16(portBytes, uint16(client.Port))
	ip := net.ParseIP(client.Host)

	if ip4 := ip.To4(); ip4 != nil {
		atyp = 0x01
		addrBytes = ip4

	} else if ip6 := ip.To16(); ip6 != nil {
		atyp = 0x02
		addrBytes = ip6

	} else {
		atyp = 0x03
		addrLen = byte(len(client.Host))
		addrBytes = append([]byte{addrLen}, []byte(client.Host)...)

	}

	req := make([]byte, 0, 1+len(addrBytes)+2)
	req = append(req, atyp)
	req = append(req, addrBytes...)
	req = append(req, portBytes...)

	if _, err = agency.Write(req); err != nil {
		return err
	}

	go func() {
		if _, e := io.Copy(agency, conn); e != nil {
			logger.SWarn("<%s> %s", client.Sid, e)
		}
	}()

	if _, err = io.Copy(conn, agency); err != nil {
		logger.SWarn("<%s> %s", client.Sid, err)
	}

	return nil
}

func (s *ShadowSocksProxy) GetName() string {
	return s.Name
}

func (s *ShadowSocksProxy) GetType() Type {
	return s.Type
}

func (s *ShadowSocksProxy) GetServer() string {
	return s.Server
}

func (s *ShadowSocksProxy) GetPort() uint16 {
	return s.Port
}

func (s *ShadowSocksProxy) GetProtocol() Protocol {
	return s.Protocol
}
