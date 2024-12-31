package routing

import (
	"fmt"
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

func InitRouter(routePath, xdbPath string) {
	var err error
	router, err = CreateRouter(routePath, xdbPath)
	if err != nil {
		panic(fmt.Errorf("failed to initialize router: %w", err))
	}
}

func CreateRouter(routePath, xdbPath string) (Router, error) {
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
