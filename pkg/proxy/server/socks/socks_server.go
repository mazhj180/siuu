package socks

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"siuu/pkg/proxy/server"
	"siuu/pkg/tunnel"
	"strconv"
	"sync"
)

type socks struct {
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

	return &socks{
		port:        config.Port,
		onError:     config.Callback.OnError,
		onAcceptd:   config.Callback.OnAcceptd,
		onConnected: config.Callback.OnConnected,
		onFinished:  config.Callback.OnFinished,
		contextFunc: config.ContextFunc,
	}
}

func (s *socks) Start() error {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		lis, _ = net.Listen("tcp", ":0")
		s.port = uint16(lis.Addr().(*net.TCPAddr).Port)
	}
	s.listener = lis
	s.isRunning = true

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.onError(nil, err)
			continue
		}

		go s.process(server.NewContext(s.contextFunc(), conn, "socks"))
	}
}

func (s *socks) process(ctx *server.Context) {

	if s.onAcceptd != nil {
		s.onAcceptd(ctx)
	}

	conn := ctx.Conn()

	buf := make([]byte, 262)
	n, err := conn.Read(buf)
	if err != nil {
		s.onError(ctx, err)
		return
	}
	if ver, nmethods := buf[0], int(buf[1]); ver != 0x05 || n < nmethods+2 {
		s.onError(ctx, fmt.Errorf("invalid version"))
		return
	}

	if _, err = conn.Write([]byte{0x05, 0x00}); err != nil {
		s.onError(ctx, err)
		return
	}

	n, err = conn.Read(buf)
	if err != nil {
		s.onError(ctx, err)
		return
	}

	if n < 7 {
		s.onError(ctx, fmt.Errorf("invalid command"))
		return
	}

	ver := buf[0]
	cmd := buf[1]
	rsv := buf[2]
	atyp := buf[3]

	if ver != 0x05 || rsv != 0x00 || cmd != 0x01 {
		_, _ = conn.Write([]byte{0x05, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		s.onError(ctx, fmt.Errorf("invalid version"))
		return
	}

	idx := 4

	switch atyp {
	case 0x01:
		if n < idx+6 {
			s.onError(ctx, fmt.Errorf("invalid ipv4 and port"))
			return
		}
		ctx.DstHost = net.IP(buf[idx : idx+4]).String()
		idx += 4
		ctx.DstPort = binary.BigEndian.Uint16(buf[idx : idx+2])

	case 0x03:
		if n < idx+1 {
			s.onError(ctx, fmt.Errorf("invalid domain and port"))
			return
		}
		domainLen := int(buf[idx])
		idx += 1
		if n < idx+domainLen+2 {
			s.onError(ctx, fmt.Errorf("invalid domain and port"))
			return
		}
		domain := string(buf[idx : idx+domainLen])
		ctx.DstHost = domain
		idx += domainLen
		ctx.DstPort = binary.BigEndian.Uint16(buf[idx : idx+2])

	case 0x04:
		if n < idx+18 {
			s.onError(ctx, fmt.Errorf("invalid ipv6 and	port"))
			return
		}
		ctx.DstHost = net.IP(buf[idx : idx+16]).String()
		idx += 16
		ctx.DstPort = binary.BigEndian.Uint16(buf[idx : idx+2])

	default:
		_, _ = conn.Write([]byte{0x05, 0x08, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		s.onError(ctx, fmt.Errorf("invalid command"))
		return
	}

	if _, err = conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0}); err != nil {
		s.onError(ctx, err)
		return
	}

	ctx.Stage = "connected"

	var t tunnel.Tunnel
	if s.onConnected != nil {
		t = s.onConnected(ctx)
	}

	if t == nil {
		dialer := &net.Dialer{}

		dstConn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(ctx.DstHost, strconv.FormatUint(uint64(ctx.DstPort), 10)))
		if err != nil {
			s.onError(ctx, err)
			return
		}

		t, err = tunnel.NewSystemProxyTunnel(nil, ctx.Conn(), dstConn, ctx.SessionId())
		if err != nil {
			s.onError(ctx, err)
			return
		}
	}

	ctx.TunnelStatus = t.GetStatus()
	ctx.Stage = "transfer"

	s.activeTunnels.LoadOrStore(ctx.SessionId(), t)
	if err = t.Start(ctx); err != nil {
		s.onError(ctx, err)
	}
	s.activeTunnels.LoadAndDelete(ctx.SessionId())

	if s.onFinished != nil {
		s.onFinished(ctx)
	}
}

func (s *socks) ActiveTunnels() map[string]tunnel.Tunnel {
	tunnels := make(map[string]tunnel.Tunnel)
	s.activeTunnels.Range(func(key, value interface{}) bool {
		tunnels[key.(string)] = value.(tunnel.Tunnel)
		return true
	})
	return tunnels
}

func (s *socks) Stop() error {
	s.isRunning = false
	return s.listener.Close()
}

func (s *socks) IsRunning() bool {
	return s.isRunning
}
