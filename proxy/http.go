package proxy

import (
	"evil-gopher/logger"
	"fmt"
	"io"
	"net"
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

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			logger.SError("<%s> original connection close err :", err)
		}
	}(conn)

	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(h.Server, strconv.FormatUint(uint64(h.Port), 10)))
	if err != nil {
		return err
	}

	agency, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}

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
		defer func(agency *net.TCPConn) {
			if e := agency.Close(); e != nil {
				logger.SError("<%s> http agency connection close err : %s", h.Name, err)
			}
		}(agency)

		if _, e := io.Copy(agency, conn); e != nil {
			logger.SError("<%s> data copy err : %s", h.Name, e)
			return
		}
	}()

	_, err = io.Copy(conn, agency)

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
