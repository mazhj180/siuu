package handler

import (
	"siu/session"
	"siu/tunnel/proxy"
)

var (
	handlers []handle
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
