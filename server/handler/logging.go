package handler

import (
	"siuu/logger"
)

func loggingHandle(ctx *context) {
	s := ctx.session
	addr := s.GetConn().RemoteAddr()
	logger.PDebug("<%s> [scope: handshake] client: [%s] arrival", s.ID(), addr)
	logger.SInfo("<%s> [scope: handshake] client: [%s] arrival ", s.ID(), addr)

	ctx.next()

	sid, pro, host, port, prx := s.ID(), s.GetProtocol(), s.GetHost(), s.GetPort(), s.GetProxy()
	if ctx.err != nil {
		logger.SError("<%s> [%s] [%s] to [%s:%d] using by [%s]  err: %s",
			sid,
			pro,
			addr,
			host,
			port,
			prx.GetName(),
			ctx.err)

		logger.PError("<%s> connect to [%s:%d] using by [%s] failed", sid, host, port, prx.GetName())
	} else {
		logger.PInfo("<%s> send to [%s:%d] using by [%s]  [up:%d B | %.2f KB/s] [down:%d B | %.2f KB/s] ",
			sid,
			host,
			port,
			prx.GetName(),
			ctx.up, ctx.upSpeed/1024, ctx.down, ctx.downSpeed/1024)
		logger.SInfo("<%s> client: [%s] is dispatching", s.ID(), addr)
	}
}
