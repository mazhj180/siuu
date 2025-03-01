package http

import (
	"context"
	"fmt"
	"net"
	"siuu/tunnel/proxy"
	"strconv"
	"strings"
	"time"
)

type Proxy struct {
	Type     proxy.Type
	Name     string
	Server   string
	Port     uint16
	Protocol proxy.Protocol
}

func (h *Proxy) Connect(ctx context.Context, addr string, port uint16) (*proxy.Pd, error) {

	dialer := &net.Dialer{
		Timeout: 30 * time.Second,
	}
	agency, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(h.Server, strconv.FormatUint(uint64(h.Port), 10)))
	if err != nil {
		return nil, err
	}

	req := fmt.Sprintf("CONNECT %s:%d HTTP/1.1\r\nHost: %s:%d\r\n\r\n", addr, port, addr, port)
	if _, err = agency.Write([]byte(req)); err != nil {
		return nil, err
	}

	resp := make([]byte, 4096)
	n, err := agency.Read(resp)
	if err != nil {
		return nil, err
	}

	respStr := string(resp[:n])
	if !strings.Contains(respStr, "200") {
		return nil, proxy.ErrProxyResp
	}

	return proxy.NewPd(agency), nil
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
