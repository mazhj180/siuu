package server

import (
	"context"
	"fmt"
	"net"
	"siuu/pkg/tunnel"
	"sync/atomic"
)

type ProxyServer interface {
	Start() error
	Stop() error
	IsRunning() bool
	ActiveTunnels() map[string]tunnel.Tunnel
}

const maxSid = 0x400

var counter int32

func genSid() string {
	for {
		cur := atomic.LoadInt32(&counter)
		newVal := (cur + 1) % (maxSid + 1)
		if atomic.CompareAndSwapInt32(&counter, cur, newVal) {
			return fmt.Sprintf("sid-%#X", newVal)
		}
	}
}

type Context struct {
	context.Context
	conn net.Conn

	sessionId string

	DstHost string
	DstPort uint16

	Stage         string
	SelectedRoute string
	MatchedRule   string
	TunnelStatus  *tunnel.Status
}

func NewContext(ctx context.Context, conn net.Conn, prefix string) *Context {

	return &Context{
		Context:   ctx,
		conn:      conn,
		sessionId: fmt.Sprintf("%s-%s", prefix, genSid()),
		Stage:     "acceptd",
	}
}

func (c *Context) Conn() net.Conn {
	return c.conn
}

func (c *Context) SessionId() string {
	return c.sessionId
}

type Config struct {
	Port        uint16
	Callback    *Callback
	ContextFunc func() context.Context
}

func DefaultConfig() *Config {
	return &Config{
		Port: 8080,
		ContextFunc: func() context.Context {
			return context.Background()
		},
	}
}

type Callback struct {
	OnError     func(*Context, error)        // on error
	OnAcceptd   func(*Context)               // on acceptd
	OnConnected func(*Context) tunnel.Tunnel // on connected
	OnFinished  func(*Context)               // on finished
}
