package config

import (
	"os"
	"path"
	"siuu/logger"
	"siuu/routing"
	"siuu/server/config/constant"
	"siuu/server/store"
	"siuu/util"
)

func InitConfig(p1, p2, p3 *uint16) {
	v := util.CreateConfig("conf", "toml")

	// server
	*p1 = v.GetUint16(constant.ServerPort)
	*p2 = v.GetUint16(constant.ProxyHttpPort)
	*p3 = v.GetUint16(constant.ProxySocksPort)

	// logger
	logPath := v.GetString(constant.LogDirPath)
	logPath = util.ExpandHomePath(logPath)
	logger.InitSystemLog(path.Dir(logPath)+"/system.log", 10*logger.MB, logger.LogLevel(v.GetString(constant.SystemLogLevel)))
	logger.InitProxyLog(path.Dir(logPath)+"/proxy.log", 1*logger.MB, logger.LogLevel(v.GetString(constant.ProxyLogLevel)))
	_, _ = os.Stdout.WriteString("init config" + logPath + "\n")

	// proxy
	prxPath := v.GetString(constant.ProxiesConfigPath)
	prxPath = util.ExpandHomePath(prxPath)
	store.InitProxy(prxPath)

	// router
	if v.GetBool(constant.RouterEnabled) {
		routePath := v.GetString(constant.RouteConfigPath)
		routePath = util.ExpandHomePath(routePath)
		xdbPath := v.GetString(constant.RouteXdbPath)
		xdbPath = util.ExpandHomePath(xdbPath)
		routing.InitRouter(routePath, xdbPath)
	}
}
