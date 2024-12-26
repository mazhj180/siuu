package routing

import (
	"bufio"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"siu/tunnel/proxy"
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
	area      map[string]proxy.Proxy
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

	//config := util.CreateConfig("conf", "toml")
	//filepath := util.ExpandHomePath(config, "route.table.path")
	filepath := ""
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

func InitRouter(v *viper.Viper) {
	enable := v.GetBool("router.true")
	if enable {
		p := v.GetString("router.route.table.path")
		router = NewDefaultRouter(p, p)

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
}

type Router interface {
	Route(string) (proxy.Proxy, error)
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
		return proxy.GetDirect(), "none", nil
	}

	r, err := router.Route(host)
	if err != nil {
		return nil, "", err
	}
	return r, "ip", nil
}
