package handler

import (
	"siuu/logger"
	"siuu/server/config/proxies"
)

func routeHandle(ctx *context) {
	s := ctx.session
	host := s.GetHost()

	r := ctx.router
	if r != nil {
		prx, _, err := r.Route(host, false)
		if err != nil {
			logger.SWarn("<%s> [router] route failed: %s", s.ID(), err)
			s.SetProxy(proxies.GetSelectedProxy())
		} else if p := proxies.GetProxy(prx); p == nil {
			s.SetProxy(proxies.GetSelectedProxy())
		} else {
			s.SetProxy(p)
			ctx.hit = true
		}

		ctx.routerName = r.Name()
		ctx.prxName = s.GetProxy().GetName()
	}

	ctx.next()
}
