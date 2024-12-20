package handler

import (
	"evil-gopher/logger"
	"evil-gopher/routing"
	"evil-gopher/tunnel"
)

func proxyHandle(ctx *context) {
	s := ctx.session
	err := s.Handshakes()
	if err != nil {
		logger.SError("ack proxy resp was failed; err: %s", err)
		return
	}
	logger.SDebug("<%s> client handshakes with gop server success", s.ID())
	host, ip := s.GetHost(), s.GetPort()
	logger.SDebug("<%s> client dst addr was [%s]", s.ID(), host)

	prx, r, err := routing.Route(host, ip)
	if err != nil {
		logger.SError("<%s> route routing failed; err: %s", s.ID(), err)
	}
	logger.SDebug("<%s> client routing using by [%s] router", s.ID(), r)
	s.SetProxy(prx)

	tunnel.T.In(s)
	ctx.next()
}
