package router

import (
	"fmt"
	"siuu/server/config/proxies"
	"siuu/server/config/rules"
	"siuu/tunnel"
	"siuu/tunnel/routing"
	"siuu/tunnel/routing/rule"
)

func NewBasicRouter() *routing.BasicRouter {

	r := &routing.BasicRouter{
		Rules:       make(map[string]rule.Interface),
		IgnoreRules: make(map[string]struct{}),
	}

	_ = rules.LoadRules(r)
	proxies.LoadProxy()

	return r
}

func RefreshBasicRouter(r routing.Router) error {
	br, ok := r.(*routing.BasicRouter)
	if !ok {
		return fmt.Errorf("router [%s] is not BasicRouter", r.Name())
	}

	br.Lock()
	defer br.Unlock()

	tunnel.T.Interrupt()
	br.Rules = make(map[string]rule.Interface)
	br.IgnoreRules = make(map[string]struct{})

	_ = rules.LoadRules(br)
	proxies.LoadProxy()

	return nil
}
