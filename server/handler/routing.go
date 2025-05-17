package handler

import (
	"siuu/logger"
	"siuu/server/config/proxies"
	"siuu/server/handler/routing"
)

func routeHandle(ctx *context) {
	s := ctx.session
	host := s.GetHost()

	r := routing.R()
	if r != nil {
		prx, rule, err := r.Route(host)
		if err != nil {
			logger.SWarn("<%s> [routing] route failed: %s", s.ID(), err)
			s.SetProxy(proxies.GetSelectedProxy())
		} else {
			logger.SDebug("<%s> [routing] matched [%s] to [%s] by [%s::%s]", s.ID(), ctx.dst, prx.GetName(), r.Name(), rule)
			s.SetProxy(prx)
			ctx.hit = true
		}

		ctx.routerName = r.Name()
		ctx.prxName = s.GetProxy().GetName()
		ctx.rule = rule
	}

	ctx.next()
}
