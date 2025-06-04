package routing

import (
	"errors"
	"fmt"
	"siuu/tunnel/proxy"
	"siuu/tunnel/routing/rule"
	"sync"
)

var (
	NotFoundRuleErr   = errors.New("not found rule")
	MissAnyRulesErr   = errors.New("miss any rules")
	HitIgnoredRuleErr = errors.New("hit ignored rule")
	UnknownProxyErr   = errors.New("unknown proxy")

	IgnoreFlag = "-1"
)

type Router interface {
	Name() string

	// Route dst to proxy
	Route(string, bool) (prx proxy.Proxy, traces []Trace, err error)

	// Boot router
	Boot(loader Loader) error

	// router status manager
	routerStatusManager
}

type routerStatusManager interface {
	AddRule(...rule.Interface)
	IgnoreRule(string) error
	AddProxies(...proxy.Proxy)

	SetDefaultProxy(proxy.Proxy) error
	GetDefaultProxy() proxy.Proxy
	GetProxy(string) proxy.Proxy
	GetAllProxies() []proxy.Proxy

	sync.Locker
	RLock()
	RUnlock()
}

type BasicRouter struct {
	sync.RWMutex

	DefaultProxy proxy.Proxy // default proxy, if no rule hit

	Rules       map[string]rule.Interface // for rules searching
	RuleIds     []string                  // for sorting
	IgnoreRules map[string]struct{}

	Proxies    map[string]proxy.Proxy
	ProxyNames []string          // for sorting
	ProxyAlias map[string]string // proxy alias

	Tracer
}

func (r *BasicRouter) Name() string {
	return "basic"
}

func (r *BasicRouter) Boot(loader Loader) error {
	if loader == nil {
		return NoRouterLoaderErr
	}

	r.Lock()
	defer r.Unlock()
	r.ini()

	return loader(r)
}

func (r *BasicRouter) Route(dst string, trace bool) (proxy.Proxy, []Trace, error) {

	r.RLock()
	defer r.RUnlock()

	var dstId string
	var traces []Trace
	var prx proxy.Proxy

	r.Trace(fmt.Sprintf("router [%s] dst [%s]", r.Name(), dst), PREFERRED, &traces, trace)

	for _, id := range r.RuleIds {
		ru := r.Rules[id]

		r.Trace(fmt.Sprintf("checking rule [%s]", ru), ALL, &traces, trace)

		if _, ok := r.IgnoreRules[id]; ok {
			dstId = IgnoreFlag
			r.Trace(fmt.Sprintf("hit ignored rule [%s] skipped", ru), ALL, &traces, trace)
			continue
		}

		if tar, ok := ru.Match(dst); ok {
			dstId = id
			prx, ok = r.Proxies[tar]
			if !ok {
				prx = r.Proxies[r.ProxyAlias[tar]]
			}
			r.Trace(fmt.Sprintf("matched successfully hit rule [%s]", ru), PREFERRED, &traces, trace)
			break
		}

		r.Trace(fmt.Sprintf("miss rule [%s]", ru), ALL, &traces, trace)
	}

	if dstId == "" {
		r.Trace("miss any rules", PREFERRED, &traces, trace)
		return r.DefaultProxy, traces, MissAnyRulesErr
	}

	if dstId == IgnoreFlag {
		r.Trace("unmatched", PREFERRED, &traces, trace)
		return r.DefaultProxy, traces, HitIgnoredRuleErr
	}

	return prx, traces, nil
}

func (r *BasicRouter) AddRule(rus ...rule.Interface) {

	for _, ru := range rus {
		r.Rules[ru.ID()] = ru
		r.RuleIds = append(r.RuleIds, ru.ID())
	}
}

func (r *BasicRouter) IgnoreRule(id string) error {

	if _, ok := r.Rules[id]; !ok {
		return NotFoundRuleErr
	}

	r.IgnoreRules[id] = struct{}{}
	return nil
}

func (r *BasicRouter) AddProxies(proxies ...proxy.Proxy) {

	for _, prx := range proxies {
		r.Proxies[prx.Name()] = prx
		r.ProxyNames = append(r.ProxyNames, prx.Name())
	}
}

func (r *BasicRouter) SetDefaultProxy(prx proxy.Proxy) error {

	r.DefaultProxy = prx
	if prx == nil {
		r.DefaultProxy = prx
		return nil
	}

	name := prx.Name()

	var ok bool
	if prx, ok = r.Proxies[name]; !ok {
		return UnknownProxyErr
	}

	r.DefaultProxy = prx
	return nil
}

func (r *BasicRouter) GetDefaultProxy() proxy.Proxy {
	return r.DefaultProxy
}

func (r *BasicRouter) GetProxy(name string) proxy.Proxy {
	prx, ok := r.Proxies[name]
	if !ok {
		prx = r.Proxies[r.ProxyAlias[name]]
	}
	return prx
}

func (r *BasicRouter) GetAllProxies() []proxy.Proxy {
	var proxies = make([]proxy.Proxy, 0, len(r.ProxyNames))
	for _, p := range r.ProxyNames {
		proxies = append(proxies, r.Proxies[p])
	}

	return proxies
}

func (r *BasicRouter) ini() {
	r.Rules = make(map[string]rule.Interface)
	r.RuleIds = make([]string, 0)
	r.IgnoreRules = make(map[string]struct{})

	r.Proxies = make(map[string]proxy.Proxy)
	r.ProxyNames = make([]string, 0)
	r.ProxyAlias = make(map[string]string)
}
