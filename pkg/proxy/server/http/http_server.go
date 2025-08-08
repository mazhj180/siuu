package http

import (
	"bufio"
	"context"
	"fmt"
	"net"
	stdhttp "net/http"
	"siuu/pkg/proxy/server"
	"siuu/pkg/tunnel"
	httputil "siuu/pkg/util/net/http"
	"strconv"
	"strings"
	"sync"
)

type http struct {
	isRunning bool

	listener net.Listener

	port uint16

	activeTunnels sync.Map

	onError     func(*server.Context, error)
	onAcceptd   func(*server.Context)
	onConnected func(*server.Context) tunnel.Tunnel
	onFinished  func(*server.Context)
	contextFunc func() context.Context
}

func New(config *server.Config) server.ProxyServer {
	if config == nil {
		config = server.DefaultConfig()
	}

	if config.Callback.OnError == nil {
		config.Callback.OnError = func(ctx *server.Context, err error) {}
	}

	return &http{
		port:        config.Port,
		onError:     config.Callback.OnError,
		onAcceptd:   config.Callback.OnAcceptd,
		onConnected: config.Callback.OnConnected,
		onFinished:  config.Callback.OnFinished,
		contextFunc: config.ContextFunc,
	}
}

func (h *http) Start() error {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", h.port))
	if err != nil {
		lis, _ = net.Listen("tcp", ":0")
		h.port = uint16(lis.Addr().(*net.TCPAddr).Port)
	}
	h.listener = lis
	h.isRunning = true

	for {
		conn, err := h.listener.Accept()
		if err != nil {
			h.onError(nil, err)
			continue
		}

		go h.process(server.NewContext(h.contextFunc(), conn, "http"))
	}

}

func (h *http) process(ctx *server.Context) {

	if h.onAcceptd != nil {
		h.onAcceptd(ctx)
	}

	conn := ctx.Conn()
	req, err := stdhttp.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		h.onError(ctx, err)
		return
	}
	var domain string
	var port uint64 = 80
	if strings.Contains(req.Host, ":") {
		d, p, _ := net.SplitHostPort(req.Host)
		domain = d
		port, _ = strconv.ParseUint(p, 10, 16)
	} else {
		domain = req.Host
	}

	isTLS := true
	if req.Method == stdhttp.MethodConnect {
		port = 443
		if _, err = ctx.Conn().Write([]byte(fmt.Sprintf("%s 200 Connection Established\r\n\r\n", req.Proto))); err != nil {
			h.onError(ctx, err)
			return
		}
	} else {
		isTLS = false
		conn = &c{
			Reader: httputil.NewHttpReader(req),
			Writer: ctx.Conn(),
		}

	}
	isPrivate := false
	if ip := net.ParseIP(domain); ip != nil && (ip.IsLoopback() || ip.IsPrivate()) {
		isPrivate = true
	}

	ctx.DstHost = domain
	ctx.DstPort = uint16(port)

	ctx.Stage = "connected"

	var t tunnel.Tunnel
	if h.onConnected != nil && isTLS && !isPrivate {
		t = h.onConnected(ctx)
	}

	if t == nil {

		dstConn, err := h.directDial(ctx, domain, uint16(port))
		if err != nil {
			h.onError(ctx, err)
			return
		}

		t, err = tunnel.NewSystemProxyTunnel(nil, conn, dstConn, ctx.SessionId())
		if err != nil {
			h.onError(ctx, err)
			return
		}
	}

	ctx.Stage = "transfer"
	ctx.TunnelStatus = t.GetStatus()

	h.activeTunnels.LoadOrStore(ctx.SessionId(), t)
	if err = t.Start(ctx); err != nil {
		h.onError(ctx, err)
	}
	h.activeTunnels.LoadAndDelete(ctx.SessionId())

	if h.onFinished != nil {
		h.onFinished(ctx)
	}

}

func (h *http) ActiveTunnels() map[string]tunnel.Tunnel {
	tunnels := make(map[string]tunnel.Tunnel)
	h.activeTunnels.Range(func(key, value interface{}) bool {
		tunnels[key.(string)] = value.(tunnel.Tunnel)
		return true
	})
	return tunnels
}

func (h *http) Stop() error {
	h.isRunning = false
	return h.listener.Close()
}

func (h *http) IsRunning() bool {
	return h.isRunning
}

func (h *http) directDial(ctx *server.Context, host string, port uint16) (net.Conn, error) {
	dialer := &net.Dialer{}

	dstConn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, strconv.FormatUint(uint64(port), 10)))
	if err != nil {
		return nil, err
	}

	return dstConn, nil
}
