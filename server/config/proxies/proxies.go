package proxies

import (
	"errors"
	"fmt"
	"siuu/tunnel"
	"siuu/tunnel/proxy"
	"siuu/tunnel/proxy/http"
	"siuu/tunnel/proxy/shadow"
	"siuu/tunnel/proxy/socks"
	"siuu/tunnel/proxy/torjan"
	"strconv"
	"strings"
	"sync"
)

func ParseProxy(p string) (proxy.Proxy, error) {

	p = strings.TrimSpace(p)
	val := strings.Split(p, ",")
	port, err := strconv.ParseUint(val[3], 10, 16)
	if err != nil {
		return nil, err
	}
	protocol := proxy.TCP
	if val[len(val)-1] == "udp" {
		protocol = proxy.UDP
	}

	basePrx := proxy.BaseProxy{
		Server:   val[2],
		Port:     uint16(port),
		Protocol: protocol,
	}

	var prx proxy.Proxy
	switch val[0] {
	case proxy.HTTPS.String(), "http":
		prx = http.New(basePrx, val[1])
	case proxy.SOCKS.String():
		prx = socks.New(basePrx, val[1], val[4], val[5])
	case proxy.SHADOW.String():
		prx = shadow.New(basePrx, val[1], val[4], val[5])
	case proxy.TROJAN.String():
		prx = torjan.New(basePrx, val[1], val[4], val[5])
	default:
		return nil, fmt.Errorf("%w: %s", proxy.ErrProxyTypeNotSupported, val[0])
	}

	return prx, nil
}

func TestProxyConnection(proxies []proxy.Proxy) map[string]float64 {
	n := len(proxies)
	traf := make(chan *testRes, n)

	var wg sync.WaitGroup
	wg.Add(n)
	for i := range proxies {
		go func() {
			defer wg.Done()
			if tr, err := tunnel.T.Ping(proxies[i]); err == nil {
				traf <- &testRes{delay: tr.Delay, prx: proxies[i].Name()}
			} else if errors.Is(err, tunnel.PingTimeoutErr) {
				traf <- &testRes{delay: -1, prx: proxies[i].Name()} // timeout
			} else {
				traf <- &testRes{delay: -2, prx: proxies[i].Name()} // error
			}
		}()
	}

	wg.Wait()
	close(traf)

	res := make(map[string]float64, n)
	for t := range traf {
		res[t.prx] = t.delay
	}

	return res
}

type testRes struct {
	prx   string
	delay float64
}
