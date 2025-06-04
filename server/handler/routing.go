package handler

func routeHandle(ctx *context) {
	s := ctx.session
	host := s.GetHost()

	visitor := ctx.visitor

	visitor.RLock()

	res := visitor.Visit()
	r := res.Router

	if r != nil {
		prx, _, err := r.Route(host, false)

		if prx != nil {
			s.SetProxy(prx)
		}

		if err == nil && prx != nil {
			ctx.hit = true
		}

		ctx.routerName = r.Name()
		ctx.prxName = s.GetProxy().Name()
	}

	visitor.RUnlock()

	ctx.next()
}
