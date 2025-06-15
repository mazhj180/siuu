package test

import (
	"path"
	"runtime"
	"siuu/tunnel/logger"
	"testing"
)

func TestLogger(t *testing.T) {
	//logger.InitLog("", 1024*1024, logger.InfoLevel)
	_, filename, _, _ := runtime.Caller(0)
	logFile := path.Dir(filename+"/../../../") + "/log/system.log"
	logger.InitSystemLog(logFile, logger.GB, logger.InfoLevel)
	logFile = path.Dir(filename+"/../../../") + "/log/proxy.log"
	logger.InitProxyLog(logFile, 1024*10, logger.InfoLevel)
	for i := 0; i < 1000; i++ {
		logger.SInfo("this is info log %d", i)
		logger.PInfo("this is warnaada log %d", i)
	}
	logger.Close()
}

func BenchmarkLogger(b *testing.B) {

}
