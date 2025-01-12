package constant

import (
	"runtime"
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
