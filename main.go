package main

import (
	"os"
	"os/signal"
	"path"
	"siu/logger"
	"siu/routing"
	"siu/server"
	"siu/server/store"
	"siu/util"
)

var (
	serverPort uint16
	httpPort   uint16
	socksPort  uint16
)

func init() {

	_, _ = os.Stdout.WriteString("initialize configuration ....\n")
	sysViper := util.CreateConfig("conf", "toml")
	sl := sysViper.GetString("log.system.level")
	pl := sysViper.GetString("log.proxy.level")
	logPath := sysViper.GetString("log.path")
	logger.InitSystemLog(path.Dir(logPath)+"/system.log", 10*logger.MB, util.LogLevel(sl))
	logger.InitProxyLog(path.Dir(logPath)+"/proxy.log", 1*logger.MB, util.LogLevel(pl))

	serverPort = sysViper.GetUint16("server.port")
	httpPort = sysViper.GetUint16("server.proxy.http.port")
	socksPort = sysViper.GetUint16("server.proxy.socks.port")

	prxPath := sysViper.GetString("proxy.path")
	prxPath = util.ExpandHomePath(prxPath)
	store.InitProxy(prxPath)
	_, _ = os.Stdout.WriteString("load proxy\n")

	if sysViper.GetBool("router.enable") {
		routePath := sysViper.GetString("router.route.path")
		routePath = util.ExpandHomePath(routePath)

		xdbPath := sysViper.GetString("router.xdb.path")
		xdbPath = util.ExpandHomePath(xdbPath)
		routing.InitRouter(routePath, xdbPath)
		_, _ = os.Stdout.WriteString("load router\n")
	}
	_, _ = os.Stdout.WriteString("siu welcomes you\n")

}

func main() {

	go server.StartServer(serverPort)
	go server.StartHttpProxyServer(httpPort)
	go server.StartSocksProxyServer(socksPort)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	_ = <-sigCh
}
