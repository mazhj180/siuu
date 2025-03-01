package handler

import (
	"siuu/server/session"
	"siuu/tunnel/proxy"
)

var (
	handlers []handle
)

func init() {
	handlers = []handle{
		loggingHandle,
		handshakeHandle,
		routeHandle,
		forwardHandle,
	}
}

type handle func(*context)

type context struct {
	session session.Session
	index   int

	// route
	routerName string
	prxName    string
	hit        bool

	// forward
	err                error
	up, down           int64
	upSpeed, downSpeed float64
}

func (c *context) skip() {
	c.index++
	c.next()
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
