package server

import (
	"context"
	"fmt"
	"github.com/kardianos/service"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path"
	"siuu/logger"
	"siuu/server/config"
	"siuu/server/config/constant"
	"siuu/server/config/router"
	"siuu/server/handler"
	"siuu/server/session"
	"siuu/tunnel/routing"
	"siuu/util"
	"sync"
	"time"
)

var (
	Srv = &Server{
		PprofPort: 6060,
		Mux:       http.NewServeMux(),
	}
)

func loggMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("HTTP Request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start),
		)
	})
}

type Server struct {
	config.Config

	PprofPort uint16

	PprofIsOpen bool
	Mux         *http.ServeMux
	Router      routing.Router

	ppSrv  *http.Server
	ppMute sync.Mutex

	Model constant.Model
}

func (s *Server) Start(_ service.Service) error {

	s.Config = config.InitConfig(service.Interactive())

	if s.EnabledRule {
		constant.Signature = make(map[string]string)
		s.Router = router.NewBasicRouter()
	}

	if s.EnablePProf {
		go s.startPprofServer()
	}

	go s.startServer()
	go s.startHttpProxyServer()
	go s.startSocksProxyServer()

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

func (s *Server) startServer() {

	RegisterProxyHandle("/prx")
	RegisterRouterHandle("/route")
	RegisterConfHandle("/pprof")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.ServerPort), loggMiddleware(Srv.Mux)); err != nil {
		panic(fmt.Sprintf("hub server start error: %s", err))
	}
}

func (s *Server) startPprofServer() {

	s.ppMute.Lock()
	defer s.ppMute.Unlock()

	if s.PprofIsOpen {
		return
	}

	s.ppSrv = &http.Server{
		Addr:    fmt.Sprintf(":%d", Srv.PprofPort),
		Handler: http.DefaultServeMux,
	}

	s.PprofIsOpen = true

	if err := s.ppSrv.ListenAndServe(); err != nil {
		panic(fmt.Sprintf("pprof server start error: %s", err))
	}

}

func (s *Server) stopPprofServer() {

	s.ppMute.Lock()
	defer s.ppMute.Unlock()

	if !Srv.PprofIsOpen || Srv.ppSrv == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	_ = Srv.ppSrv.Shutdown(ctx)
	defer cancel()

	Srv.ppSrv = nil
	Srv.PprofIsOpen = false
}

func (s *Server) startHttpProxyServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.HttpProxyPort))
	if err != nil {
		lis, _ = net.Listen("tcp", ":0")
		s.HttpProxyPort = uint16(lis.Addr().(*net.TCPAddr).Port)
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			logger.SError("<%s> http conn accept err : %s", conn.RemoteAddr().String(), err)
			continue
		}
		sess := session.OpenHttpSession(conn)
		go handler.Run(sess, s.Router)
	}
}

func (s *Server) startSocksProxyServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.SocksProxyPort))
	if err != nil {
		lis, _ = net.Listen("tcp", ":0")
		s.SocksProxyPort = uint16(lis.Addr().(*net.TCPAddr).Port)
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			logger.SError("<%s> socks conn accept err : %s", conn.RemoteAddr().String(), err)
			continue
		}
		sess := session.OpenSocksSession(conn)
		go handler.Run(sess, s.Router)
	}
}
