package main

import (
	"evil-gopher/server"
	"evil-gopher/util"
	"os"
	"os/signal"
)

var (
	httpPort  int
	socksPort int
)

func init() {
	sysViper := util.CreateConfig("conf", "toml")
	httpPort = sysViper.GetInt("server.proxy.http.port")
	socksPort = sysViper.GetInt("server.proxy.socks.port")

}

func main() {

	go server.StartHttpProxyServer(httpPort)
	go server.StartSocksProxyServer(socksPort)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	_ = <-sigCh
}
