package server

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"siuu/internal/config"
	"siuu/internal/server/handlers"
	serv "siuu/internal/server/service"
	"siuu/pkg/logger"
	"siuu/pkg/logger/splitter"
	"siuu/pkg/proxy/route"
	"siuu/pkg/proxy/server"
	proxy_http "siuu/pkg/proxy/server/http"
	proxy_socks "siuu/pkg/proxy/server/socks"
	httputil "siuu/pkg/util/net/http"
	"siuu/pkg/util/path"
	"siuu/pkg/util/toml"
	"time"

	"github.com/kardianos/service"
)

type Siuu struct {
	conf *config.SystemConfig // system config

	ctlServer        *http.Server       // control server
	pprofServer      *http.Server       // pprof server
	httpProxyServer  server.ProxyServer // http proxy server
	socksProxyServer server.ProxyServer // socks proxy server
	router           route.Router       // router

	usedIdx int // used index for router
}

func New(conf *config.SystemConfig) (*Siuu, error) {

	siuu := &Siuu{}

	siuu.conf = conf

	// load router config and initialize router
	if siuu.usedIdx >= len(conf.Server.Proxy.Tables) {
		siuu.usedIdx = 0
	}

	table := conf.Server.Proxy.Tables[siuu.usedIdx]
	table = path.ExpandHomePath(table)

	router := route.NewRouter()
	var routerConfig config.RouterConfig
	if err := toml.LoadTomlFromFile(table, &routerConfig); err != nil {
		return nil, err
	}

	prxs, err := routerConfig.GetProxies()
	if err != nil {
		return nil, err
	}

	if err = router.Initialize(routerConfig.GetRouterRules(), prxs, routerConfig.GetMappings(), nil); err != nil {
		return nil, err
	}

	siuu.router = router

	// initialize system logger
	slog, err := siuu.initSystemLogger()
	if err != nil {
		return nil, err
	}

	// initialize control server handlers
	mux := http.NewServeMux()
	wrapperHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a custom ResponseWriter to capture status code
			ww := &httputil.StatusWriter{ResponseWriter: w, StatusCode: 200}
			next.ServeHTTP(ww, r)

			slog.Debug("[system-ctl] http request method=%s path=%s status=%d duration=%s",
				r.Method, r.URL.Path, ww.StatusCode, time.Since(start))
		})
	}
	siuu.ctlServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", conf.Server.Port),
		Handler: wrapperHandler(mux),
	}

	rh := handlers.GetRouterHandlers(siuu.router, slog)
	for path, handler := range rh {
		mux.HandleFunc(path, handler)
	}

	sh := handlers.GetSystemHandlers(conf, slog)
	for path, handler := range sh {
		mux.HandleFunc(path, handler)
	}

	// initialize proxy logger
	plog, err := siuu.initProxyLogger()
	if err != nil {
		return nil, err
	}

	// initialize proxy server callbacks
	callbacks := serv.GetCallbacks(siuu.router, plog)

	if conf.Server.Proxy.Http.Enable && conf.Server.Proxy.Mode != "tun" {
		c := server.DefaultConfig()
		c.Port = uint16(conf.Server.Proxy.Http.Port)
		c.Callback = callbacks

		siuu.httpProxyServer = proxy_http.New(c)
	}

	if conf.Server.Proxy.Socks.Enable && conf.Server.Proxy.Mode != "tun" {
		c := server.DefaultConfig()
		c.Port = uint16(conf.Server.Proxy.Socks.Port)
		c.Callback = callbacks

		siuu.socksProxyServer = proxy_socks.New(c)
	}

	return siuu, nil
}

func (s *Siuu) Start(_ service.Service) error {

	if s.conf.Server.Pprof.Enable {
		go s.StartPprofServer() // start pprof server
	}

	if s.httpProxyServer != nil {
		go s.httpProxyServer.Start() // start http proxy server
	}

	if s.socksProxyServer != nil {
		go s.socksProxyServer.Start() // start socks proxy server
	}

	go s.ctlServer.ListenAndServe() // start control server

	return nil
}

func (s *Siuu) Stop(_ service.Service) error {
	return nil
}

func (s *Siuu) StartPprofServer() error {

	s.pprofServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.conf.Server.Pprof.Port),
		Handler: http.DefaultServeMux,
	}

	return s.pprofServer.ListenAndServe()
}

func (s *Siuu) StopPprofServer() error {

	if s.pprofServer == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := s.pprofServer.Shutdown(ctx)
	s.pprofServer = nil

	return err
}

func (s *Siuu) initProxyLogger() (*logger.Logger, error) {
	logconf := logger.DefaultConfig()
	logconf.Level = logger.LevelString(s.conf.Log.Level.Proxy)
	logconf.Async = true
	logconf.LogDir = path.ExpandHomePath(s.conf.Log.Path)
	logconf.BaseName = "proxy"
	logconf.Splitter = splitter.NewSizeOverwriteSplitter(1024 * 1024 * 10)
	return logger.New(logconf)
}

func (s *Siuu) initSystemLogger() (*logger.Logger, error) {
	logconf := logger.DefaultConfig()
	logconf.Level = logger.LevelString(s.conf.Log.Level.System)
	logconf.Async = false
	logconf.LogDir = path.ExpandHomePath(s.conf.Log.Path)
	logconf.BaseName = "system"
	logconf.Splitter = splitter.NewSizeOverwriteSplitter(1024 * 1024 * 10)
	return logger.New(logconf)
}
