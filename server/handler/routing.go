package handler

import (
	"siuu/logger"
	"siuu/server/handler/routing"
	"siuu/server/store"
)

func routeHandle(ctx *context) {
	s := ctx.session
	host := s.GetHost()
	r := routing.R()
	if r != nil {
		if prx, err := r.Route(host); err != nil {
			logger.SWarn("<%s> route router failed; err: %s", s.ID(), err)
			s.SetProxy(store.GetSelectedProxy())
		} else {
			logger.SDebug("<%s> client routing using by [%s] router", s.ID(), r.Name())
			s.SetProxy(prx)
		}
		ctx.routerName = r.Name()
	}
	ctx.next()
}
