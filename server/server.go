package server

import (
	"evil-gopher/handler"
	"evil-gopher/logger"
	"evil-gopher/session"
	"fmt"
	"net"
)

func StartHttpProxyServer(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(fmt.Sprintf("http proxy server start error: %s", err))
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			logger.SError("<%s> http conn accept err : %s", conn.RemoteAddr().String(), err)
			continue
		}
		sess := session.OpenHttpSession(conn)
		go handler.Run(sess)
	}
}

func StartSocksProxyServer(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(fmt.Errorf("socks proxy server start error: %s", err))
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			logger.SError("<%s> socks conn accept err : %s", conn.RemoteAddr().String(), err)
			continue
		}
		sess := session.OpenSocksSession(conn)
		go handler.Run(sess)
	}
}
