package store

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"maps"
	"siuu/tunnel/proxy"
	"strconv"
	"strings"
	"sync"
)

var (
	direct     *proxy.DirectProxy
	proxyTable map[string]proxy.Proxy
	rwx        sync.RWMutex

	selected proxy.Proxy
	rwxS     sync.RWMutex
)

func InitProxy(filepath string) {
	direct = &proxy.DirectProxy{
		Type:     proxy.DIRECT,
		Name:     "direct",
		Protocol: proxy.TCP,
	}
	proxyTable = make(map[string]proxy.Proxy)
	selected = direct
	v := viper.New()
	v.SetConfigFile(filepath)
	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("failed to initialize proxy: %w", err))
	}
	proxies := v.GetStringSlice("proxy.proxies")
	if err := AddProxies(proxies...); err != nil {
		panic(fmt.Errorf("failed to initialize proxy: %w", err))
	}
}

func AddProxies(proxies ...string) error {
	rwx.Lock()
	defer rwx.Unlock()
	for _, p := range proxies {
		p = strings.TrimSpace(p)
		val := strings.Split(p, ",")
		port, err := strconv.ParseUint(val[3], 10, 16)
		if err != nil {
			return err
		}
		protocol := proxy.TCP
		if val[len(val)-1] == "udp" {
			protocol = proxy.UDP
		}

		var prx proxy.Proxy
		switch val[0] {
		case proxy.HTTPS.String(), "http":
			prx = &proxy.HttpProxy{
				Type:     proxy.HTTPS,
				Name:     val[1],
				Server:   val[2],
				Port:     uint16(port),
				Protocol: protocol,
			}
		case proxy.SOCKS.String():
			prx = &proxy.SocksProxy{
				Type:     proxy.SOCKS,
				Name:     val[1],
				Server:   val[2],
				Port:     uint16(port),
				Username: val[4],
				Password: val[5],
				Protocol: protocol,
			}
		case proxy.SHADOW.String():
			prx = &proxy.ShadowSocksProxy{
				Type:     proxy.SHADOW,
				Name:     val[1],
				Server:   val[2],
				Port:     uint16(port),
				Cipher:   val[4],
				Password: val[5],
				Protocol: protocol,
			}
		case proxy.TROJAN.String():
			prx = &proxy.TrojanProxy{
				Type:     proxy.TROJAN,
				Name:     val[1],
				Server:   val[2],
				Port:     uint16(port),
				Password: val[4],
				Protocol: protocol,
				Sni:      val[5],
			}
		default:
			return fmt.Errorf("%w: %s", proxy.ErrProxyTypeNotSupported, val[0])
		}
		if _, ok := proxyTable[prx.GetName()]; ok {
			return errors.New("the same agent already exists")
		}
		proxyTable[prx.GetName()] = prx
	}
	return nil
}

func GetDirect() *proxy.DirectProxy {
	return direct
}

func GetProxyTable() map[string]proxy.Proxy {
	rwx.RLock()
	defer rwx.RUnlock()
	var duplicate map[string]proxy.Proxy
	maps.Copy(duplicate, proxyTable)
	return duplicate
}

func GetProxy(name string) proxy.Proxy {
	if name == "direct" {
		return direct
	}
	if name == "default" {
		return selected
	}
	rwx.RLock()
	defer rwx.RUnlock()
	return proxyTable[name]
}

func GetProxies() []proxy.Proxy {
	rwx.RLock()
	defer rwx.RUnlock()
	var proxies = make([]proxy.Proxy, 0, len(proxyTable))
	for _, p := range proxyTable {
		proxies = append(proxies, p)
	}

	return append([]proxy.Proxy{selected, direct}, proxies...)
}

func RemoveProxies(names ...string) {
	rwx.Lock()
	defer rwx.Unlock()
	for _, n := range names {
		delete(proxyTable, n)
	}
}

func SetSelectedProxy(prx string) error {
	rwxS.Lock()
	defer rwxS.Unlock()
	if prx == "direct" {
		selected = direct
		return nil
	}

	if p, ok := proxyTable[prx]; ok {
		selected = p
		return nil
	}
	return errors.New("proxy not exist")
}

func GetSelectedProxy() proxy.Proxy {
	rwxS.RLock()
	defer rwxS.RUnlock()
	return selected
}
