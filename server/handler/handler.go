package handler

import (
	"net"
	"siuu/server/session"
	"siuu/tunnel/proxy"
)

var (
	handlers        []handle
	privateIPBlocks = []net.IPNet{
		{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
		{IP: net.IPv4(172, 16, 0, 0), Mask: net.CIDRMask(12, 32)},
		{IP: net.IPv4(192, 168, 0, 0), Mask: net.CIDRMask(16, 32)},
		{IP: net.IPv4(127, 0, 0, 0), Mask: net.CIDRMask(8, 32)}, // 本地回环
	}
)

func init() {
	handlers = []handle{
		loggingHandle,
		proxyHandle,
	}
}

type handle func(*context)

type context struct {
	session session.Session
	index   int
}

func (c *context) next() {
	c.index++
	if c.index < len(handlers) {
		handlers[c.index](c)
	}
}

func (c *context) setProxy(p proxy.Proxy) {
	c.session.SetProxy(p)
}

func Run(s session.Session) {
	ctx := &context{
		session: s,
		index:   -1,
	}
	ctx.next()
}
