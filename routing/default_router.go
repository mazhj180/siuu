package routing

import (
	"fmt"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/spf13/viper"
	"net"
	"os"
	"siuu/server/store"
	"siuu/tunnel/proxy"
	"strings"
)

type wildcard struct {
	rule string
	prx  proxy.Proxy
}

type routeTable struct {
	exacts    map[string]proxy.Proxy
	wildcards []*wildcard
	geo       map[string]proxy.Proxy
}

type DefaultRouter struct {
	route   routeTable
	ipXdb   []byte
	xdbPath string
}

func NewDefaultRouter(route, xdbp string) (*DefaultRouter, error) {
	v := viper.New()
	v.SetConfigFile(route)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	r := &DefaultRouter{
		route: routeTable{
			exacts:    make(map[string]proxy.Proxy),
			wildcards: make([]*wildcard, 0),
			geo:       make(map[string]proxy.Proxy),
		},
	}

	exacts := v.GetStringSlice("route.exacts")
	for _, val := range exacts {
		val = strings.TrimSpace(val)
		kvs := strings.Split(val, ",")

		prx := store.GetProxy(kvs[1])
		if prx == nil {
			_, _ = os.Stdout.WriteString(fmt.Sprintf("failed to get proxy: %s\n", kvs[1]))
			continue
		}
		r.route.exacts[kvs[0]] = store.GetProxy(kvs[1])
	}

	wildcards := v.GetStringSlice("route.wildcards")
	for _, val := range wildcards {
		val = strings.TrimSpace(val)
		kvs := strings.Split(val, ",")

		prx := store.GetProxy(kvs[1])
		if prx == nil {
			_, _ = os.Stdout.WriteString(fmt.Sprintf("failed to get proxy: %s\n", kvs[1]))
			continue
		}
		r.route.wildcards = append(r.route.wildcards, &wildcard{
			rule: kvs[0],
			prx:  prx,
		})
	}

	geo := v.GetStringSlice("route.geo")
	for _, val := range geo {
		val = strings.TrimSpace(val)
		kvs := strings.Split(val, ",")

		prx := store.GetProxy(kvs[1])
		if prx == nil {
			_, _ = os.Stdout.WriteString(fmt.Sprintf("failed to get proxy: %s\n", kvs[1]))
			continue
		}
		r.route.geo[kvs[0]] = prx
	}
	xdbb, err := xdb.LoadVectorIndexFromFile(xdbp)
	if err != nil {
		panic(fmt.Errorf("failed to load ip router please check or disable autorouting err: %s", err))
	}
	r.ipXdb = xdbb
	r.xdbPath = xdbp
	return r, nil
}

func (r *DefaultRouter) Route(host string) (proxy.Proxy, error) {

	if p, ok := r.route.exacts[host]; ok {
		return p, nil
	}

	for _, w := range r.route.wildcards {
		rl := len(w.rule[1:])
		hl := len(host)
		if hl < rl {
			continue
		}
		if host[hl-rl:] == w.rule[1:] {
			return w.prx, nil
		}
	}

	searcher, err := xdb.NewWithVectorIndex(r.xdbPath, r.ipXdb)
	if err != nil {
		return nil, fmt.Errorf("failed to load ip router please check or disable autorouting err: %s", err)
	}

	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return nil, fmt.Errorf("failed to lookup ip err: %s", err)
	}
	str, err := searcher.SearchByStr(ips[0].String())
	if err != nil {
		return nil, fmt.Errorf("failed to search ip router by str err: %s", err)
	}

	for k, v := range r.route.geo {
		if k[0:1] == "!" {
			if !strings.Contains(str, k[1:]) {
				return v, nil
			}
		}
		if strings.Contains(str, k) {
			return v, nil
		}
	}

	return store.GetSelectedProxy(), nil
}

func (*DefaultRouter) Name() string {
	return "default-router"
}
