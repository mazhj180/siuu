package proxy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"evil-gopher/logger"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type TrojanProxy struct {
	Type     Type
	Name     string
	Server   string
	Port     uint16
	Password string
	Protocol Protocol
	Sni      string
}

func (t *TrojanProxy) Act(client *Client) error {
	if t.Protocol == TCP {
		return t.actOfTcp(client)
	} else if t.Protocol == UDP {
		return ErrProtocolNotSupported
	}
	return nil
}

func (t *TrojanProxy) actOfTcp(client *Client) error {
	conn := client.Conn

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			logger.SError("<%s> original connection close err :", err)
		}
	}(conn)

	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(t.Server, strconv.FormatUint(uint64(t.Port), 10)))
	if err != nil {
		return err
	}

	agency, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}

	// tls handshake
	tlsConfig := &tls.Config{
		ServerName: t.Sni,
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			return nil
		},
	}
	tlsConn := tls.Client(agency, tlsConfig)

	defer func(tlsConn *tls.Conn) {
		_ = tlsConn.Close()
	}(tlsConn)
	err = tlsConn.SetDeadline(time.Now().Add(30 * time.Second))
	if err != nil {
		return err
	}

	if err = tlsConn.Handshake(); err != nil {
		return err
	}

	// trojan handshake
	authMsg := t.Password + "\r\n"
	if _, err = tlsConn.Write([]byte(authMsg)); err != nil {
		return err
	}

	r := bufio.NewReader(tlsConn)
	resp, err := r.ReadString('\n')
	if err != nil {
		return err
	}

	if !strings.Contains(resp, "successful") {
		return ErrProxyResp
	}

	// send connect req
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(0x01)

	var atyp byte
	var addrBytes []byte
	var addrLen byte
	var portBytes []byte

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

	if _, err = tlsConn.Write(buf.Bytes()); err != nil {
		return err
	}

	// analyze resp
	reply := make([]byte, 10)
	if _, err = io.ReadFull(tlsConn, reply); err != nil {
		return err
	}

	rep := reply[1]
	if rep != 0x01 {
		return ErrProxyResp
	}

	// bidirectional forwarding
	go func() {
		defer func(tlsConn net.Conn) {
			if e := tlsConn.Close(); e != nil {
				logger.SError("<%s> agency connection close err :", err)
			}
		}(tlsConn)

		if _, e := io.Copy(tlsConn, conn); e != nil {
			logger.SError("<%s> data copy err :", e)
			return
		}
	}()

	_, _ = io.Copy(conn, agency)

	return nil

}

func (t *TrojanProxy) GetName() string {
	return t.Name
}

func (t *TrojanProxy) GetType() Type {
	return t.Type
}

func (t *TrojanProxy) GetServer() string {
	return t.Server
}

func (t *TrojanProxy) GetPort() uint16 {
	return t.Port
}

func (t *TrojanProxy) GetProtocol() Protocol {
	return t.Protocol
}
