package routing

import (
	"siuu/logger"
	"siuu/tunnel/proxy"
	"sync"
)

type Router interface {
	Name() string
	Route(string) (proxy.Proxy, error)
}

var (
	router Router
	rwx    sync.RWMutex
)

func InitRouter(routePath []string, xdbPath string) {
	var err error
	router, err = CreateRouter(routePath, xdbPath)
	if err != nil {
		logger.SWarn("failed to initialize router: %s", err)
	}
}

func CreateRouter(routePath []string, xdbPath string) (Router, error) {
	rwx.Lock()
	defer rwx.Unlock()
	r, err := NewDefaultRouter(routePath, xdbPath)
	return r, err
}

func R() Router {
	rwx.RLock()
	defer rwx.RUnlock()
	return router
}

func CloseRouter() {
	rwx.Lock()
	defer rwx.Unlock()
	router = nil
}

func Refresh(routePath []string, xdbPath string) error {
	r, err := NewDefaultRouter(routePath, xdbPath)
	if err != nil {
		return err
	}
	rwx.Lock()
	defer rwx.Unlock()
	router = r
	return nil
}
