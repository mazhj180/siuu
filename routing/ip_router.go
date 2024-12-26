package routing

import (
	"fmt"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"net"
	"os"
	"path"
	"siu/tunnel/proxy"
	"siu/util"
	"strings"
)

type DefaultRouter struct {
	route routeTable
	ipXdb []byte
}

func NewDefaultRouter(route, sw string) *DefaultRouter {

	route = util.ExpandHomePath(path.Dir(route))
	exacts := path.Join(route, "exacts")
	fe, err := os.OpenFile(exacts, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(fe)

	wildcards := path.Join(route, "wildcards")
	fe, err = os.OpenFile(wildcards, os.O_RDONLY|os.O_CREATE, 0666)
	xdbp := path.Join(route, "ip2region.xdb")
	xdbb, err := xdb.LoadContentFromFile(xdbp)

	if err != nil {
		panic(fmt.Errorf("failed to load ip router please check or disable autorouting err: %s", err))
	}
	return &DefaultRouter{ipXdb: xdbb}
}

func (r *DefaultRouter) Route(host string) (proxy.Proxy, error) {

	if p, ok := r.route.exacts[host]; ok {
		return p, nil
	}

	for _, w := range r.route.wildcards {
		if strings.Contains(host, w.rule[1:]) {
			return w.prx, nil
		}
	}

	searcher, err := xdb.NewWithBuffer(r.ipXdb)
	if err != nil {
		return nil, fmt.Errorf("failed to load ip router please check or disable autorouting err: %s", err)
	}

	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return nil, nil
	}
	str, err := searcher.SearchByStr(ips[0].String())
	if err != nil {
		return nil, fmt.Errorf("failed to search ip router by str err: %s", err)
	}

	for k, v := range r.route.area {
		if strings.Contains(str, k) {
			return v, nil
		}
	}

	return proxy.GetSelectedProxy(), nil
}
