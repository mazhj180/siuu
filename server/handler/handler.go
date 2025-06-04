package handler

import (
	"net"
	visitor "siuu/server/resources_visitor"
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
	visitor visitor.Visitor

	session session.Session
	index   int

	remoteAddr net.Addr

	// handshake
	handshake bool
	dst       string

	// route
	routerName string
	prxName    string
	hit        bool
	rule       string

	// forward
	err                error
	up, down           int64
	upTime, downTime   float64
	upSpeed, downSpeed float64
	delay              float64
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

func Run(s session.Session, visitor visitor.Visitor) {

	ctx := &context{
		visitor: visitor,
		session: s,
		index:   -1,
	}

	ctx.next()

	// preventing unreleased locks
	visitor.Unlock()
	visitor.RUnlock()
}
