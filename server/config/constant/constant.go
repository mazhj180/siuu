package constant

import (
	_ "embed"
	"runtime"
	"siuu/util"
)

var (
	//go:embed conf.toml
	Conf []byte

	//go:embed proxies.toml
	Proxies []byte
)

type Model byte

const (
	NORMAL Model = iota
	TUN
)

type Key = string

const (
	ServerPort     Key = "server.port"
	ProxyHttpPort      = "server.http.port"
	ProxySocksPort     = "server.socks.port"
	EnablePProf        = "server.pprof.enable"

	SystemLogLevel   = "log.level.system"
	ProxyLogLevel    = "log.level.proxy"
	LogDirPath       = "log.path"
	ProxyModel       = "proxy.model"
	RuleEnabled      = "rule.enable"
	RuleRoutePath    = "rule.route.path"
	RuleRouteXdbPath = "rule.route.xdb"
	RuleProxyPath    = "rule.proxy.path"
)

var (
	RootPath = util.AppRootPath
	Platform PlatformKind

	Signature map[string]string
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
