package server

import (
	"fmt"
	"github.com/kardianos/service"
	"log"
	"net"
	"net/http"
	"siuu/handler"
	"siuu/logger"
	"siuu/server/config"
	"siuu/server/handle"
	"siuu/session"
)

type Server struct{}

func (s *Server) Start(_ service.Service) error {

	var serverPort, httpPort, socksPort uint16
	config.InitConfig(&serverPort, &httpPort, &socksPort)

	go startServer(serverPort)
	go startHttpProxyServer(httpPort)
	go startSocksProxyServer(socksPort)

	return nil
}

func (s *Server) Stop(_ service.Service) error {
	log.Println("[program] Stopping... (clean up resources)")
	return nil
}

func startServer(port uint16) {
	mux := http.NewServeMux()
	handle.RegisterProxyHandle(mux, "/prx")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		panic(fmt.Sprintf("hub server start error: %s", err))
	}
}

func startHttpProxyServer(port uint16) {
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

func startSocksProxyServer(port uint16) {
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
