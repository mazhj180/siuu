package routing

import (
	"bufio"
	"evil-gopher/proxy"
	"evil-gopher/util"
	"fmt"
	"net"
	"os"
	"slices"
	"sync"
)

type wildcard struct {
	rule string
	prx  proxy.Proxy
}

type routeTable struct {
	exacts    map[string]proxy.Proxy
	wildcards []*wildcard
}

var (
	table = &routeTable{}
	rwx   sync.RWMutex
)

func AddRoute(host string, proxy proxy.Proxy) {
	rwx.Lock()
	defer rwx.Unlock()
	if host[0] != '*' {
		table.exacts[host] = proxy
		return
	}
	w := &wildcard{
		rule: host,
		prx:  proxy,
	}
	table.wildcards = append(table.wildcards, w)
}

func RemoveRoute(host string) error {
	rwx.Lock()
	defer rwx.Unlock()

	if _, ok := table.exacts[host]; ok {
		delete(table.exacts, host)
		return nil
	}
	slices.DeleteFunc(table.wildcards, func(w *wildcard) bool {
		return w.rule == host
	})

	config := util.CreateConfig("conf", "toml")
	filepath := util.ExpandHomePath(config, "route.table.path")

	rf, err := os.OpenFile(filepath, os.O_RDWR|os.O_TRUNC|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	bufW := bufio.NewWriter(rf)
	for _, w := range table.wildcards {
		if _, err = bufW.WriteString(fmt.Sprintf("%s|%s", w.rule, w.prx.GetName())); err != nil {
			return err
		}
	}
	if err = bufW.Flush(); err != nil {
		return err
	}

	if err = rf.Close(); err != nil {
		return err
	}

	return nil
}

func ListAllRoute() map[string]string {
	rwx.RLock()
	defer rwx.RUnlock()
	var routes map[string]string
	for k, v := range table.exacts {
		routes[k] = v.GetName()
	}
	for _, w := range table.wildcards {
		routes[w.rule] = w.prx.GetName()
	}
	return routes
}

var (
	router Router
	rwxr   sync.RWMutex
)

func init() {
	config := util.CreateConfig("conf", "toml")
	if e := config.GetBool("router.enable"); e {
		router = NewIPRouter()
	}
}

func CloseRouter() {
	rwxr.Lock()
	defer rwxr.Unlock()
	router = nil
}

func OpenRouter() {
	rwxr.Lock()
	defer rwxr.Unlock()
	router = NewIPRouter()
}

type Router interface {
	Route(*TargetHost) (proxy.Proxy, error)
}

type TargetHost struct {
	IP     net.IP
	Port   uint16
	Domain string
}

func Route(host string, port uint16) (proxy.Proxy, string, error) {
	if p, ok := table.exacts[host]; ok {
		return p, "exacts", nil
	}

	for _, w := range table.wildcards {
		if w.rule[1:] == host {
			return w.prx, "wildcards", nil
		}
	}

	if router == nil {
		return proxy.GetDirect(host, port), "none", nil
	}

	th := &TargetHost{
		IP:     net.ParseIP(host),
		Port:   port,
		Domain: host,
	}
	r, err := router.Route(th)
	if err != nil {
		return nil, "", err
	}
	return r, "ip", nil
}
