package config

import (
	"os"
	"path"
	"runtime"
	"siuu/logger"
	"siuu/routing"
	"siuu/server/store"
	"siuu/util"
)

type Key = string

const (
	ServerPort        Key = "server.port"
	ProxyHttpPort         = "server.http.port"
	ProxySocksPort        = "server.socks.port"
	SystemLogLevel        = "log.level.system"
	ProxyLogLevel         = "log.level.proxy"
	LogDirPath            = "log.path"
	RouterEnabled         = "router.enable"
	RouteConfigPath       = "router.path.table"
	RouteXdbPath          = "router.path.xdb"
	ProxiesConfigPath     = "proxy.path"
)

var (
	RootPath = util.AppRootPath
	Platform PlatformKind
)

type PlatformKind byte

const (
	Windows PlatformKind = iota
	Linux
	Darwin
)

func init() {
	switch runtime.GOOS {
	case "windows":
		Platform = Windows
	case "linux":
		Platform = Linux
	case "darwin":
		Platform = Darwin
	}
}

func InitConfig(p1, p2, p3 *uint16) {
	v := util.CreateConfig("conf", "toml")

	// server
	*p1 = v.GetUint16(ServerPort)
	*p2 = v.GetUint16(ProxyHttpPort)
	*p3 = v.GetUint16(ProxySocksPort)

	// logger
	logPath := v.GetString(LogDirPath)
	logPath = util.ExpandHomePath(logPath)
	logger.InitSystemLog(path.Dir(logPath)+"/system.log", 10*logger.MB, logger.LogLevel(v.GetString(SystemLogLevel)))
	logger.InitProxyLog(path.Dir(logPath)+"/proxy.log", 1*logger.MB, logger.LogLevel(v.GetString(ProxyLogLevel)))
	_, _ = os.Stdout.WriteString("init config" + logPath + "\n")

	// proxy
	prxPath := v.GetString(ProxiesConfigPath)
	prxPath = util.ExpandHomePath(prxPath)
	store.InitProxy(prxPath)

	// router
	if v.GetBool(RouterEnabled) {
		routePath := v.GetString(RouteConfigPath)
		routePath = util.ExpandHomePath(routePath)
		xdbPath := v.GetString(RouteXdbPath)
		xdbPath = util.ExpandHomePath(xdbPath)
		routing.InitRouter(routePath, xdbPath)
	}
}

func Set(key Key, value any) error {
	v := util.CreateConfig("conf", "toml")
	v.Set(key, value)

	if err := v.WriteConfig(); err != nil {
		return err
	}
	return nil
}

func Get[T ~int64 | ~string | ~bool | ~[]int64 | ~[]string](key Key) T {
	var zero T
	v := util.CreateConfig("conf", "toml")
	if v.Get(key) == nil {
		return zero
	}
	return v.Get(key).(T)
}

func GetAll() map[Key]any {
	v := util.CreateConfig("conf", "toml")
	return v.AllSettings()
}
