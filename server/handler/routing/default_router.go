package routing

import (
	"crypto/sha256"
	"fmt"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/spf13/viper"
	"io"
	"net"
	"os"
	"siuu/logger"
	"siuu/server/config/constant"
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

func NewDefaultRouter(routeFile []string, xdbp string) (*DefaultRouter, error) {
	v := viper.New()
	v.SetConfigType("toml")

	r := &DefaultRouter{
		route: routeTable{
			exacts:    make(map[string]proxy.Proxy),
			wildcards: make([]*wildcard, 0),
			geo:       make(map[string]proxy.Proxy),
		},
	}

	for _, f := range routeFile {

		if _, err := os.Stat(f); os.IsNotExist(err) {
			continue
		}
		file, err := os.OpenFile(f, os.O_RDONLY, 0666)
		if err != nil {
			continue
		}
		defer file.Close()

		hasher := sha256.New()
		_, err = io.Copy(hasher, file)
		if err != nil {
			logger.SWarn("failed to initialize routing [%s]", f)
			continue
		}

		signature := fmt.Sprintf("%xroute", hasher.Sum(nil))
		if s, ok := constant.Signature[f]; ok && s == signature {
			continue
		}
		constant.Signature[f] = signature

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			logger.SWarn("failed to initialize routing [%s]", f)
			continue
		}

		if err = v.ReadConfig(file); err != nil {
			return nil, err
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
	}

	xdbb, err := xdb.LoadVectorIndexFromFile(xdbp)
	if err != nil {
		return r, nil
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
		return nil, fmt.Errorf("failed to load ip router please check or disable autorouting err: %w", err)
	}

	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return nil, fmt.Errorf("failed to lookup ip err: %w", err)
	}
	str, err := searcher.SearchByStr(ips[0].String())
	if err != nil {
		return nil, fmt.Errorf("failed to search ip router by str err: %w", err)
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
