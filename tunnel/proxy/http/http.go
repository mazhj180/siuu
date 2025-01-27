package http

import (
	"fmt"
	"io"
	"net"
	"siuu/logger"
	"siuu/tunnel/proxy"
	"strconv"
	"strings"
)

type Proxy struct {
	Type     proxy.Type
	Name     string
	Server   string
	Port     uint16
	Protocol proxy.Protocol
}

func (h *Proxy) ForwardTcp(client *proxy.Client) error {

	conn := client.Conn
	defer conn.Close()

	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(h.Server, strconv.FormatUint(uint64(h.Port), 10)))
	if err != nil {
		return err
	}

	agency, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}
	defer agency.Close()

	if err = agency.SetKeepAlive(true); err != nil {
		return err
	}

	if client.IsTLS {
		req := fmt.Sprintf("CONNECT %s:%d HTTP/1.1\r\nHost: %s:%d\r\n\r\n", client.Host, client.Port, client.Host, client.Port)
		if _, err = agency.Write([]byte(req)); err != nil {
			return err
		}

		resp := make([]byte, 4096)
		n, err := agency.Read(resp)
		if err != nil {
			return err
		}
		respStr := string(resp[:n])
		if !strings.Contains(respStr, "200") {
			return proxy.ErrProxyResp
		}
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

func (h *Proxy) ForwardUdp(client *proxy.Client) (*proxy.UdpPocket, error) {
	return nil, proxy.ErrProtocolNotSupported
}

func (h *Proxy) GetName() string {
	return h.Name
}

func (h *Proxy) GetType() proxy.Type {
	return h.Type
}

func (h *Proxy) GetServer() string {
	return h.Server
}

func (h *Proxy) GetPort() uint16 {
	return h.Port
}

func (h *Proxy) GetProtocol() proxy.Protocol {
	return h.Protocol
}
