package handler

import (
	"net"
	"siuu/server/session"
	"siuu/tunnel/proxy"
	"siuu/tunnel/routing"
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
	router routing.Router

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

func Run(s session.Session, router routing.Router) {
	ctx := &context{
		router:  router,
		session: s,
		index:   -1,
	}
	ctx.next()
}
