package torjan

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"io"
	"net"
	"siuu/logger"
	"siuu/tunnel/proxy"
	"strconv"
)

type TrojanProxy struct {
	Type     proxy.Type
	Name     string
	Server   string
	Port     uint16
	Password string
	Protocol proxy.Protocol
	Sni      string
}

func (t *TrojanProxy) Act(client *proxy.Client) error {
	if t.Protocol == proxy.TCP {
		return t.actOfTcp(client)
	} else if t.Protocol == proxy.UDP {
		return proxy.ErrProtocolNotSupported
	}
	return nil
}

func (t *TrojanProxy) actOfTcp(client *proxy.Client) error {
	conn := client.Conn
	defer conn.Close()

	// tls handshake
	tlsConfig := &tls.Config{
		ServerName: t.Sni,
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			return nil
		},
	}

	agency, err := tls.Dial("tcp", net.JoinHostPort(t.Server, strconv.FormatUint(uint64(t.Port), 10)), tlsConfig)
	if err != nil {
		return err
	}
	defer agency.Close()

	// trojan handshake
	hash := sha256.New224()
	hash.Write([]byte(t.Password))
	pwd := hex.EncodeToString(hash.Sum(nil))
	authMsg := pwd + "\r\n"
	if _, err = agency.Write([]byte(authMsg)); err != nil {
		return err
	}

	// send connect req
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(0x01)

	var atyp byte
	var addrBytes []byte
	var addrLen byte
	var portBytes = make([]byte, 2)

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

	buf.WriteByte(atyp)
	buf.Write(addrBytes)

	binary.BigEndian.PutUint16(portBytes, client.Port)
	buf.Write(portBytes)

	if _, err = agency.Write(append(buf.Bytes(), []byte{0x0d, 0x0a}...)); err != nil {
		return err
	}

	// bidirectional forwarding
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

func (t *TrojanProxy) GetName() string {
	return t.Name
}

func (t *TrojanProxy) GetType() proxy.Type {
	return t.Type
}

func (t *TrojanProxy) GetServer() string {
	return t.Server
}

func (t *TrojanProxy) GetPort() uint16 {
	return t.Port
}

func (t *TrojanProxy) GetProtocol() proxy.Protocol {
	return t.Protocol
}
