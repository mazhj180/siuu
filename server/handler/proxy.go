package handler

import (
	"fmt"
	"siuu/logger"
	"siuu/server/handler/routing"
	"siuu/tunnel"
)

func proxyHandle(ctx *context) {
	s := ctx.session
	err := s.Handshakes()
	if err != nil {
		logger.SError("<%s> ack proxy resp was failed; err: %s", s.ID(), err)
		return
	}
	logger.SDebug("<%s> client: handshakes successfully with siuu server", s.ID())
	host := s.GetHost()
	port := s.GetPort()
	addr := fmt.Sprintf("%s:%d", host, port)
	logger.SDebug("<%s> dst addr was [%s]", s.ID(), addr)

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
