package http

import (
	"context"
	"fmt"
	"net"
	"siuu/pkg/proxy/client"
	"strconv"
	"strings"
	"time"
)

var _ client.ProxyClient = &http{}

type http struct {
	client.BaseClient

	name string
}

func New(base client.BaseClient, name string) client.ProxyClient {
	return &http{
		BaseClient: base,
		name:       name,
	}
}

func (h *http) Connect(ctx context.Context, proto, host string, port uint16) (net.Conn, error) {
	switch proto {
	case "tcp":
		return h.connectTcp(ctx, host, port)
	case "udp":
		return h.connectUdp(ctx, host, port)
	default:
		return nil, client.ErrProxyResp
	}
}

func (h *http) connectTcp(ctx context.Context, host string, port uint16) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: 30 * time.Second,
	}
	agency, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(h.Server, strconv.FormatUint(uint64(h.Port), 10)))
	if err != nil {
		return nil, err
	}

	req := fmt.Sprintf("CONNECT %s:%d HTTP/1.1\r\nHost: %s:%d\r\n\r\n", host, port, host, port)
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
		return nil, client.ErrProxyResp
	}

	return agency, nil
}

func (h *http) connectUdp(ctx context.Context, host string, port uint16) (net.Conn, error) {
	return nil, client.ErrProtoNotSupported
}

func (h *http) Name() string {
	return h.name
}

func (h *http) Type() string {
	return "http"
}

func (h *http) String() string {
	return fmt.Sprintf(
		`{"Server":"%s","Port":%d,"Protocol":"%s","Name":"%s","Type":"%s"}`,
		h.Server,
		h.Port,
		h.TrafficType,
		h.name,
		h.Type(),
	)
}
