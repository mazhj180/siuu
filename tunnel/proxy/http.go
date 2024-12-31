package proxy

import (
	"fmt"
	"io"
	"net"
	"siuu/logger"
	"strconv"
	"strings"
)

type HttpProxy struct {
	Type     Type
	Name     string
	Server   string
	Port     uint16
	Protocol Protocol
}

func (h *HttpProxy) Act(client *Client) error {
	if h.Protocol == TCP {
		if err := h.actOfTcp(client); err != nil {
			return err
		}
	} else if h.Protocol == UDP {
		if err := h.actOfUdp(client); err != nil {
			return err
		}
	}
	return ErrProtocolNotSupported
}

func (h *HttpProxy) actOfTcp(client *Client) error {

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
			return ErrProxyResp
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

func (h *HttpProxy) actOfUdp(client *Client) error {
	return ErrProtocolNotSupported
}

func (h *HttpProxy) GetName() string {
	return h.Name
}

func (h *HttpProxy) GetType() Type {
	return h.Type
}

func (h *HttpProxy) GetServer() string {
	return h.Server
}

func (h *HttpProxy) GetPort() uint16 {
	return h.Port
}

func (h *HttpProxy) GetProtocol() Protocol {
	return h.Protocol
}
