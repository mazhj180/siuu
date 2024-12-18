package proxy

import (
	"errors"
	"maps"
	"net"
	"sync"
)

var (
	proxyTable map[string]Proxy
	rwx        sync.RWMutex
)

const (
	DIRECT Type = iota
	REJECT
	HTTPS
	SOCKS
	SHADOW
	TROJAN

	TCP Protocol = 1
	UDP Protocol = 2
)

var ErrProxyTypeNotSupported = errors.New("proxy type not supported")
var ErrProtocolNotSupported = errors.New("protocol not supported")
var ErrProxyResp = errors.New("proxy response error")

func init() {
	direct := &DirectProxy{
		Type:     DIRECT,
		Name:     "direct",
		Protocol: TCP,
	}
	proxyTable = make(map[string]Proxy)
	proxyTable["direct"] = direct
	selectedProxy = direct
}

type Type int

type Protocol byte

type Client struct {
	Conn  net.Conn
	Host  string
	Port  uint16
	IsTLS bool
}

type Proxy interface {
	Act(*Client) error
	GetType() Type
	GetName() string
	GetServer() string
	GetPort() uint16
	GetProtocol() Protocol
}

type tcp interface {
	actOfTcp(*Client) error
}

type udp interface {
	actOfUdp(*Client) error
}

func GetDirect(server string, port uint16) *DirectProxy {
	prx := proxyTable["direct"].(*DirectProxy)
	cprx := &DirectProxy{}
	*cprx = *prx
	cprx.Server = server
	cprx.Port = port
	return cprx
}

func GetProxyTable() map[string]Proxy {
	rwx.RLock()
	defer rwx.RUnlock()
	var duplicate map[string]Proxy
	maps.Copy(duplicate, proxyTable)
	return duplicate
}

func GetProxies() []Proxy {
	rwx.RLock()
	defer rwx.RUnlock()
	var proxies = make([]Proxy, 0, len(proxyTable))
	for _, proxy := range proxyTable {
		proxies = append(proxies, proxy)
	}
	return proxies
}

func AddProxies(proxies ...Proxy) {
	rwx.Lock()
	defer rwx.Unlock()
	for _, proxy := range proxies {
		proxyTable[proxy.GetName()] = proxy
	}
}

func RemoveProxies(name string) {
	rwx.Lock()
	defer rwx.Unlock()
	proxyTable[name] = nil
}

var (
	selectedProxy Proxy
	rwxS          sync.RWMutex
)

func GetSelectedProxy() Proxy {
	rwxS.RLock()
	defer rwxS.RUnlock()
	return selectedProxy
}

func SelectProxy(proxy Proxy) error {
	rwxS.Lock()
	defer rwxS.Unlock()
	if p, ok := proxyTable[proxy.GetName()]; ok {
		selectedProxy = p
		return nil
	}
	return errors.New("proxy not exist")
}
