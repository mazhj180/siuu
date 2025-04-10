package server

import (
	"fmt"
	"github.com/kardianos/service"
	"net"
	"net/http"
	"os"
	"path"
	"siuu/logger"
	"siuu/server/config"
	"siuu/server/handle"
	"siuu/server/handler"
	"siuu/server/session"
	"siuu/util"
)

type Server struct{}

func (s *Server) Start(_ service.Service) error {

	var serverPort, httpPort, socksPort uint16
	config.InitConfig(&serverPort, &httpPort, &socksPort, service.Interactive())

	go s.startServer(serverPort)
	go s.startHttpProxyServer(httpPort)
	go s.startSocksProxyServer(socksPort)

	return nil
}

func (s *Server) Stop(_ service.Service) error {
	_, _ = fmt.Fprintf(os.Stdout, "[program] Stopping... (clean up resources)")
	return nil
}

func (s *Server) InstallConfig() {
	home := util.GetHomeDir()
	root := path.Dir(home + "/.siuu/")

	// export env

	// build config file
	if err := config.BuildConfiguration(root); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "cannot create config file %s, err: %s\n", root, err)
	}

	//download ip2region.xdb
	if err := util.DownloadIp2Region(root + "/conf"); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "download ip2region.xdb failed %s\n", err)
	}
}

func (s *Server) UninstallConfig() {
	home := util.GetHomeDir()
	if err := os.RemoveAll(home + "/.siuu"); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "uninstall config was wrong")
	}
}

func (s *Server) startServer(port uint16) {
	mux := http.NewServeMux()
	handle.RegisterProxyHandle(mux, "/prx")
	handle.RegisterRouterHandle(mux, "/route")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		panic(fmt.Sprintf("hub server start error: %s", err))
	}
}

func (s *Server) startHttpProxyServer(port uint16) {
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

func (s *Server) startSocksProxyServer(port uint16) {
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
