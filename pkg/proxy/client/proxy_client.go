package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"siuu/pkg/proxy/mux"
)

var ErrProxyResp = errors.New("proxy response error")

// ProxyClient is the interface for proxy clients.
// It is used to connect to the proxy server and get the connection.
type ProxyClient interface {
	fmt.Stringer

	Type() string // proxy type : http, socks5, shadowsocks, trojan ....
	Name() string // proxy client name

	// Connect connects to the proxy with the given host and port.
	// It returns the connection if success, otherwise returns an error.
	Connect(ctx context.Context, host string, port uint16) (net.Conn, error)

	ServerHost() string         // server host
	ServerPort() uint16         // server port
	SupportTrafficType() string // support traffic type : tcp, udp, both

}

// BaseClient is the base client for all proxy clients.
type BaseClient struct {
	Server      string
	Port        uint16
	TrafficType string

	Mux    mux.Interface
	Pool   *mux.Pool
	Dialer func(context.Context) (net.Conn, error)
}

func (b *BaseClient) ServerHost() string {
	return b.Server
}

func (b *BaseClient) ServerPort() uint16 {
	return b.Port
}

func (b *BaseClient) SupportTrafficType() string {
	return b.TrafficType
}
