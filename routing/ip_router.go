package routing

import (
	"evil-gopher/proxy"
	"evil-gopher/util"
	"fmt"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"net"
	"strings"
)

type IPRouter struct {
	ipXdb []byte
}

func NewIPRouter() *IPRouter {
	config := util.CreateConfig("conf", "toml")
	xdbp := util.ExpandHomePath(config, "route.xdb.path")

	xdbb, err := xdb.LoadContentFromFile(xdbp)
	if err != nil {
		panic(fmt.Errorf("failed to load ip router please check or disable autorouting err: %s", err))
	}
	return &IPRouter{ipXdb: xdbb}
}

func (r *IPRouter) Route(host *TargetHost) (proxy.Proxy, error) {

	searcher, err := xdb.NewWithBuffer(r.ipXdb)
	if err != nil {
		return proxy.GetDirect(host.Domain, host.Port), fmt.Errorf("failed to load ip router please check or disable autorouting err: %s", err)
	}

	ips, err := net.LookupIP(host.Domain)
	if err != nil || len(ips) == 0 {
		return proxy.GetDirect(host.Domain, host.Port), nil
	}
	str, err := searcher.SearchByStr(ips[0].String())
	if err != nil {
		return proxy.GetDirect(host.Domain, host.Port), fmt.Errorf("failed to search ip router by str err: %s", err)
	}

	if strings.Contains(str, "") {
		return proxy.GetDirect(host.Domain, host.Port), nil
	}

	return proxy.GetSelectedProxy(), nil
}
