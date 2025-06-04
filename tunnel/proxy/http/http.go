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

type p struct {
	proxy.BaseProxy

	name string
}

func New(base proxy.BaseProxy, name string) proxy.Proxy {
	return &p{
		BaseProxy: base,
		name:      name,
	}
}

func (h *p) Type() proxy.Type {
	return proxy.HTTPS
}

func (h *p) Connect(ctx context.Context, addr string, port uint16) (*proxy.Pd, error) {

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

func (h *p) Name() string {
	return h.name
}

func (h *p) String() string {
	return fmt.Sprintf(
		`{"Server":"%s","Port":%d,"Protocol":"%s","Name":"%s","Type":"%s"}`,
		h.Server,
		h.Port,
		h.Protocol,
		h.name,
		h.Type(),
	)
}
