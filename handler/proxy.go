package handler

import (
	"siuu/logger"
	"siuu/routing"
	"siuu/tunnel"
)

func proxyHandle(ctx *context) {
	s := ctx.session
	err := s.Handshakes()
	if err != nil {
		logger.SError("ack proxy resp was failed; err: %s", err)
		return
	}
	logger.SDebug("<%s> client handshakes with gop server success", s.ID())
	host := s.GetHost()
	logger.SDebug("<%s> client dst addr was [%s]", s.ID(), host)

	r := routing.R()
	if r != nil {
		if prx, err := r.Route(host); err != nil {
			logger.SWarn("<%s> route router failed; err: %s", s.ID(), err)
		} else {
			logger.SDebug("<%s> client routing using by [%s] router", s.ID(), r.Name())
			s.SetProxy(prx)
		}
	}

	tunnel.T.In(s)
	ctx.next()
}
