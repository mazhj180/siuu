package proxy

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"maps"
	"net"
	"os"
	"reflect"
	"siu/logger"
	"sync"
)

var (
	direct     *DirectProxy
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

func Init(v *viper.Viper) {
	direct = &DirectProxy{
		Type:     DIRECT,
		Name:     "direct",
		Protocol: TCP,
	}
	proxyTable = make(map[string]Proxy)
	selectedProxy = direct

	out = pout(v.GetString("proxy.conf.path"))
	proxies, err := out.read()
	if err != nil {
		panic(fmt.Errorf("read proxy config error: %v", err))
	}
	for _, p := range proxies {
		proxyTable[p.GetName()] = p
	}
}

type Type int

func (t *Type) MarshalJSON() ([]byte, error) {
	var typ string
	switch *t {
	case DIRECT:
		typ = "direct"
	case REJECT:
		typ = "reject"
	case HTTPS:
		typ = "https"
	case SOCKS:
		typ = "socks"
	case SHADOW:
		typ = "shadow"
	case TROJAN:
		typ = "trojan"
	default:
		return nil, fmt.Errorf("%w: %d", ErrProxyTypeNotSupported, t)
	}
	return json.Marshal(typ)
}

func (t *Type) UnmarshalJSON(data []byte) error {
	var typ string
	if err := json.Unmarshal(data, &typ); err != nil {
		return err
	}
	switch typ {
	case "direct":
		*t = DIRECT
	case "reject":
		*t = REJECT
	case "https":
		*t = HTTPS
	case "socks":
		*t = SOCKS
	case "shadow":
		*t = SHADOW
	case "trojan":
		*t = TROJAN
	default:
		return fmt.Errorf("%w: %s", ErrProxyTypeNotSupported, typ)
	}
	return nil
}

type Protocol byte

func (p *Protocol) MarshalJSON() ([]byte, error) {
	var proto string
	switch *p {
	case TCP:
		proto = "tcp"
	case UDP:
		proto = "udp"
	default:
		return nil, fmt.Errorf("%w: %d", ErrProtocolNotSupported, p)
	}
	return json.Marshal(proto)
}

func (p *Protocol) UnmarshalJSON(data []byte) error {
	var proto string
	if err := json.Unmarshal(data, &proto); err != nil {
		return err
	}
	switch proto {
	case "tcp":
		*p = TCP
	case "udp":
		*p = UDP
	default:
		return fmt.Errorf("%w: %s", ErrProtocolNotSupported, proto)
	}
	return nil
}

type Client struct {
	Sid   string
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

type ProxyWrapper struct {
	Type  Type  `json:"type"`
	Value Proxy `json:"value"`
}

func (pw *ProxyWrapper) UnmarshalJSON(data []byte) error {

	pm := make(map[string]any)
	if err := json.Unmarshal(data, &pm); err != nil {
		return err
	}
	pm = pm["value"].(map[string]any)
	var proto Protocol
	p := pm["Protocol"].(string)
	if p == "tcp" {
		proto = TCP
	} else if p == "udp" {
		proto = UDP
	} else {
		return fmt.Errorf("%w: %s", ErrProxyTypeNotSupported, p)
	}

	var typ reflect.Type
	switch pm["Type"] {
	case "direct":
		pw.Type = DIRECT
		typ = reflect.TypeOf((*DirectProxy)(nil)).Elem()
		prx := reflect.New(typ).Interface().(*DirectProxy)
		prx.Type = DIRECT
		prx.Name = pm["Name"].(string)
		prx.Server = pm["Server"].(string)
		prx.Port = uint16(pm["Port"].(float64))
		prx.Protocol = proto
		pw.Value = prx

	case "https":
		pw.Type = HTTPS
		typ = reflect.TypeOf((*HttpProxy)(nil)).Elem()
		prx := reflect.New(typ).Interface().(*HttpProxy)
		prx.Type = HTTPS
		prx.Name = pm["Name"].(string)
		prx.Server = pm["Server"].(string)
		prx.Port = uint16(pm["Port"].(float64))
		prx.Protocol = proto
		pw.Value = prx

	case "socks":
		pw.Type = SOCKS
		typ = reflect.TypeOf((*SocksProxy)(nil)).Elem()
		prx := reflect.New(typ).Interface().(*SocksProxy)
		prx.Type = SOCKS
		prx.Name = pm["Name"].(string)
		prx.Server = pm["Server"].(string)
		prx.Port = uint16(pm["Port"].(float64))
		prx.Username = pm["Username"].(string)
		prx.Password = pm["Password"].(string)
		prx.Protocol = proto
		pw.Value = prx

	case "shadow":
		pw.Type = SHADOW
		typ = reflect.TypeOf((*ShadowSocksProxy)(nil)).Elem()
		prx := reflect.New(typ).Interface().(*ShadowSocksProxy)
		prx.Type = SHADOW
		prx.Name = pm["Name"].(string)
		prx.Server = pm["Server"].(string)
		prx.Port = uint16(pm["Port"].(float64))
		prx.Cipher = pm["Cipher"].(string)
		prx.Password = pm["Password"].(string)
		prx.Protocol = proto
		pw.Value = prx

	case "trojan":
		pw.Type = TROJAN
		typ = reflect.TypeOf((*TrojanProxy)(nil)).Elem()
		prx := reflect.New(typ).Interface().(*TrojanProxy)
		prx.Type = TROJAN
		prx.Name = pm["Name"].(string)
		prx.Server = pm["Server"].(string)
		prx.Port = uint16(pm["Port"].(float64))
		prx.Password = pm["Password"].(string)
		prx.Sni = pm["Sni"].(string)
		prx.Protocol = proto
		pw.Value = prx
	default:
		return fmt.Errorf("%w: %d", ErrProxyTypeNotSupported, pw.Type)
	}

	return nil
}

var out pout

type pout string

func (p pout) add(proxies ...Proxy) error {
	f, err := os.OpenFile(string(p), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	for _, prx := range proxies {

		typ := prx.GetType()
		pw := &ProxyWrapper{
			Type:  typ,
			Value: prx,
		}

		data, err := json.Marshal(pw)
		if err != nil {
			return fmt.Errorf("failed to marshal proxy: %w", err)
		}

		// Write to the file and append newlines as separators
		if _, err = f.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("failed to write proxy to file: %w", err)
		}
	}
	return nil
}

func (p pout) remove() {
	go func() {
		err := os.Truncate(string(p), 0)
		if err != nil {
			logger.SError("remove goroutine was wrong :%w", err)
			return
		}
		err = p.add(GetProxies()[1:]...)
		if err != nil {
			_, _ = os.Stdout.WriteString(err.Error())
			logger.SError("remove goroutine was wrong :%w", err)
			return
		}
		logger.SDebug("the change has sync to file")
	}()
	return
}

func (p pout) read() ([]Proxy, error) {
	f, err := os.OpenFile(string(p), os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	proxies := make([]Proxy, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		var wrapper ProxyWrapper
		if err = json.Unmarshal(line, &wrapper); err != nil {
			return nil, err
		}

		prx := wrapper.Value

		proxies = append(proxies, prx)
	}
	return proxies, nil
}

func GetDirect() *DirectProxy {
	return direct
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

	return append([]Proxy{direct}, proxies...)
}

func AddProxies(proxies ...Proxy) error {
	rwx.Lock()
	defer rwx.Unlock()
	for _, proxy := range proxies {
		if _, ok := proxyTable[proxy.GetName()]; ok {
			return errors.New("the same agent already exists")
		}
		proxyTable[proxy.GetName()] = proxy
	}
	if err := out.add(proxies...); err != nil {
		return err
	}
	return nil
}

func RemoveProxies(names ...string) {
	rwx.Lock()
	defer rwx.Unlock()
	for _, n := range names {
		delete(proxyTable, n)
	}
	out.remove()
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
