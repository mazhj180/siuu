package store

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"maps"
	"os"
	"siuu/logger"
	"siuu/server/config/constant"
	"siuu/tunnel/proxy"
	"siuu/tunnel/proxy/http"
	"siuu/tunnel/proxy/shadow"
	"siuu/tunnel/proxy/socks"
	"siuu/tunnel/proxy/torjan"
	"siuu/tunnel/tester"
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

func InitProxy(filepath []string) {
	direct = &proxy.DirectProxy{
		Type:     proxy.DIRECT,
		Name:     "direct",
		Protocol: proxy.TCP,
	}
	proxyTable = make(map[string]proxy.Proxy)
	selected = direct

	constant.Signature = make(map[string]string)

	v := viper.New()
	v.SetConfigType("toml")
	for _, f := range filepath {

		if _, err := os.Stat(f); os.IsNotExist(err) {
			logger.SWarn("failed to initialize proxy [%s]", f)
			continue
		}
		fin, err := os.OpenFile(f, os.O_RDONLY, 0666)
		if err != nil {
			logger.SWarn("failed to initialize proxy [%s]", f)
			continue
		}
		defer fin.Close()

		hasher := sha256.New()
		_, err = io.Copy(hasher, fin)
		if err != nil {
			logger.SWarn("failed to initialize proxy [%s]", f)
			continue
		}

		signature := fmt.Sprintf("%xproxy", hasher.Sum(nil))
		if s, ok := constant.Signature[f]; ok && s == signature {
			continue
		}
		constant.Signature[f] = signature

		_, err = fin.Seek(0, io.SeekStart)
		if err != nil {
			logger.SWarn("failed to initialize proxy [%s]", f)
			continue
		}

		if err = v.ReadConfig(fin); err != nil {
			logger.SWarn("failed to initialize proxy [%s]", f)
			continue
		}
		proxies := v.GetStringSlice("proxy.proxies")
		if err = AddProxies(proxies...); err != nil {
			logger.SWarn("failed to initialize proxy [%s]", f)
		}
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
			prx = &http.HttpProxy{
				Type:     proxy.HTTPS,
				Name:     val[1],
				Server:   val[2],
				Port:     uint16(port),
				Protocol: protocol,
			}
		case proxy.SOCKS.String():
			prx = &socks.SocksProxy{
				Type:     proxy.SOCKS,
				Name:     val[1],
				Server:   val[2],
				Port:     uint16(port),
				Username: val[4],
				Password: val[5],
				Protocol: protocol,
			}
		case proxy.SHADOW.String():
			prx = &shadow.ShadowSocksProxy{
				Type:     proxy.SHADOW,
				Name:     val[1],
				Server:   val[2],
				Port:     uint16(port),
				Cipher:   val[4],
				Password: val[5],
				Protocol: protocol,
			}
		case proxy.TROJAN.String():
			prx = &torjan.TrojanProxy{
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
			logger.SWarn("the same agent already exists : [%s]", prx.GetName())
			continue
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

func TestProxyConnection() map[string]float64 {
	t := tester.NewTester("https://github.com/", "github.com", GetProxies())
	t.Test()
	res, err := t.GetResult()
	if err != nil {
		return nil
	}
	return res
}
