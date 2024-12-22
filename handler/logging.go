package handler

import (
	"evil-gopher/logger"
)

func loggingHandle(ctx *context) {
	s := ctx.session
	addr := s.GetConn().RemoteAddr()
	logger.PDebug("<%s> agent req : [%s] access ", s.ID(), addr)
	logger.SDebug("<%s> agent req : [%s] access ", s.ID(), addr)
	ctx.next()
	logger.SDebug("<%s> req: [%s] is dispatching", s.ID(), addr)
}
