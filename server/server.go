package server

import (
	"evil-gopher/handler"
	"evil-gopher/logger"
	"evil-gopher/session"
	"fmt"
	"net"
	"net/http"
)

func StartServer(port uint16) {
	mux := http.NewServeMux()
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		panic(fmt.Sprintf("hub server start error: %s", err))
	}
	mux.Handle("/", http.FileServer(http.Dir("./public")))
}

func StartHttpProxyServer(port uint16) {
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

func StartSocksProxyServer(port uint16) {
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
