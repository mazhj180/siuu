package handler

import (
	"evil-gopher/logger"
	"fmt"
	"path"
	"runtime"
)

func init() {
	_, filename, _, _ := runtime.Caller(0)
	logFile := path.Dir(filename+"/../../") + "/log/system.log"
	fmt.Printf(logFile)
	logger.InitSystemLog(logFile, 10*logger.MB, logger.InfoLevel)
	logFile = path.Dir(filename+"/../../") + "/log/proxy.log"
	logger.InitProxyLog(logFile, 10*logger.MB, logger.InfoLevel)
}

func loggingHandle(ctx *context) {
	s := ctx.session
	addr := s.GetConn().RemoteAddr()
	logger.PDebug("<%s> agent req : [%s] access ", s.ID(), addr)
	logger.SDebug("<%s> agent req : [%s] access ", s.ID(), addr)
	ctx.next()
	logger.SDebug("<%s> req: [%s] is dispatching", s.ID(), addr)
}
