package handler

import (
	"siuu/logger"
)

func loggingHandle(ctx *context) {
	s := ctx.session
	addr := s.GetConn().RemoteAddr()
	logger.PDebug("<%s> agent req : [%s] access ", s.ID(), addr)
	logger.SInfo("<%s> agent req : [%s] access ", s.ID(), addr)
	ctx.next()
	logger.SInfo("<%s> req: [%s] is dispatching", s.ID(), addr)
}
