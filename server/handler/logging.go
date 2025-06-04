package handler

import (
	"siuu/logger"
)

func loggingHandle(ctx *context) {
	s := ctx.session
	addr := s.GetConn().RemoteAddr()
	ctx.remoteAddr = addr
	logger.PDebug("<%s> [access] [cli: %s] is arrival", s.ID(), addr)

	ctx.next()

	if !ctx.handshake {
		return
	}

	sid, pro, host, port, prx := s.ID(), s.GetProtocol(), s.GetHost(), s.GetPort(), s.GetProxy()
	if ctx.err != nil {
		logger.SError("<%s> [ending] [%s] [%s] to [%s:%d] using by [%s]  err: %s",
			sid,
			pro,
			addr,
			host,
			port,
			prx.Name(),
			ctx.err)

		logger.PError("<%s> [ending] connect to [%s:%d] used by [%s] failed", sid, host, port, prx.Name())
	} else {
		logger.PInfo("<%s> [ending] send to [%s:%d] used by [%s] [router: %s::%s]  [up:%d B | %.2f KB/s] [down:%d B | %.2f KB/s] [delay: %d ms]",
			sid,
			host,
			port,
			prx.Name(),
			ctx.routerName,
			ctx.rule,
			ctx.up, ctx.upSpeed/1024, ctx.down, ctx.downSpeed/1024,
			int(ctx.delay*1000),
		)
	}
}
