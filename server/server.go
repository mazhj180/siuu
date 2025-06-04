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
	visitor "siuu/server/resources_visitor"
	"siuu/server/session"
	"siuu/tunnel/routing"
	"siuu/util"
	"strings"
	"sync"
	"time"
)

var (
	srv = &Server{
		PprofPort: 6060,
		mux:       http.NewServeMux(),
	}
)

func Siuu() *Server {
	return srv
}

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
	sync.RWMutex
	config.Config

	PprofPort   uint16
	PprofIsOpen bool

	Router routing.Router
	loader routing.Loader

	mux            *http.ServeMux
	handlerMapping map[string]http.HandlerFunc

	ppSrv  *http.Server
	ppMute sync.Mutex

	Model constant.Model
}

func (s *Server) Start(_ service.Service) error {

	// read conf info from conf file
	s.Config = config.InitConfig(service.Interactive())

	// Initialize the default router
	if s.EnabledRule && s.Router == nil {
		s.Router, _ = router.NewDefaultRouter()
	}

	// Add customized router if it's existing
	if _, ok := s.Router.(*routing.BasicRouter); !ok {
		_ = s.Router.Boot(s.loader)
	}

	// Add customized http handler func to the server mux
	if s.handlerMapping != nil {
		for k, v := range s.handlerMapping {
			if strings.HasPrefix(k, "/prx") || strings.HasPrefix(k, "/route") || strings.HasPrefix(k, "/pprof") {
				logger.SWarn("The mapping path of the handler func conflicts with the system default; [%s]", k)
				continue
			}
			s.mux.HandleFunc(k, v)
		}
	}

	// start pprof server
	if s.EnablePProf {
		go s.startPprofServer()
	}

	go s.startServer() // start hub server

	if s.Model == constant.NORMAL {
		go s.startHttpProxyServer()  // start http proxy server
		go s.startSocksProxyServer() // start socks5 proxy server

		return nil
	}

	// start tun model
	// todo

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

	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.ServerPort), loggMiddleware(srv.mux)); err != nil {
		panic(fmt.Sprintf("hub server start error: %s", err))
	}
}

func (s *Server) RegisterRouter(r routing.Router, loader routing.Loader) {

	if r == nil || loader == nil {
		return
	}

	s.Lock()
	defer s.Unlock()

	srv.Router = r
	srv.loader = loader
}

func (s *Server) RegisterHandlerFunc(handlerMapping map[string]http.HandlerFunc) {
	if handlerMapping == nil {
		return
	}

	s.Lock()
	defer s.Unlock()

	s.handlerMapping = handlerMapping
}

func (s *Server) startPprofServer() {

	if s.PprofIsOpen {
		return
	}

	s.ppMute.Lock()

	s.ppSrv = &http.Server{
		Addr:    fmt.Sprintf(":%d", srv.PprofPort),
		Handler: http.DefaultServeMux,
	}

	s.PprofIsOpen = true
	s.ppMute.Unlock()

	_ = s.ppSrv.ListenAndServe()

}

func (s *Server) stopPprofServer() {

	s.ppMute.Lock()
	defer s.ppMute.Unlock()

	if !srv.PprofIsOpen || srv.ppSrv == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	_ = srv.ppSrv.Shutdown(ctx)
	defer cancel()

	srv.ppSrv = nil
	srv.PprofIsOpen = false
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
		go handler.Run(sess, &srvVisitor{Server: s})
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
		go handler.Run(sess, &srvVisitor{Server: s})
	}
}

// srvVisitor is a visitor for server
// Attention: this visitor is not thread-safe, so do not use it as a singleton
//
//	one session to one visitor
type srvVisitor struct {
	*Server

	// rw is a read-write lock flag : 0 for none, 1 for read, 2 for write
	rw byte

	// state is a lock flag : true for locked, false for unlocked
	// Prevents duplicate locking and unlocking in the same session.
	state bool
}

func (s *srvVisitor) Visit() visitor.AccessibleResources {
	return visitor.AccessibleResources{
		Router: s.Router,
	}
}

func (s *srvVisitor) RLock() {
	if s.rw == 2 || s.state {
		return
	}
	s.state = true
	s.rw = 1
	s.Server.RLock()
}

func (s *srvVisitor) RUnlock() {
	if s.rw == 2 || !s.state {
		return
	}
	s.Server.RUnlock()
	s.state = false
	s.rw = 1
}

func (s *srvVisitor) Lock() {
	if s.rw == 1 || s.state {
		return
	}
	s.state = true
	s.rw = 2
	s.Server.Lock()
}

func (s *srvVisitor) Unlock() {
	if s.rw == 1 || !s.state {
		return
	}
	s.Server.Unlock()
	s.state = false
	s.rw = 2
}
