package handler

import (
	"siuu/logger"
)

func loggingHandle(ctx *context) {
	s := ctx.session
	addr := s.GetConn().RemoteAddr()
	logger.PDebug("<%s> client: [%s] arrival", s.ID(), addr)
	logger.SInfo("<%s> client: [%s] arrival ", s.ID(), addr)
	ctx.next()
	logger.SInfo("<%s> client: [%s] is dispatching", s.ID(), addr)
}
