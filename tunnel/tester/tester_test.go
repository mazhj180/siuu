package tester

import (
	"fmt"
	"siuu/tunnel/proxy"
	"siuu/tunnel/proxy/torjan"
	"testing"
)

var (
	proxies []proxy.Proxy
)

func TestNewTester(t *testing.T) {
	prx := &torjan.TrojanProxy{
		Type:     proxy.TROJAN,
		Name:     "xxxxzxz",
		Server:   "xxxxxx.com",
		Port:     uint16(8080),
		Password: "1111111",
		Protocol: proxy.TCP,
		Sni:      "3231311111",
	}

	prx1 := &proxy.DirectProxy{
		Type:     proxy.DIRECT,
		Name:     "direct",
		Protocol: proxy.TCP,
	}

	prx2 := &torjan.TrojanProxy{
		Type:     proxy.TROJAN,
		Name:     "kkkkkkk",
		Server:   "ddddd.com",
		Port:     uint16(8080),
		Password: "dddddddd",
		Protocol: proxy.TCP,
		Sni:      "dddddadddd",
	}

	proxies = append(proxies, prx, prx1, prx2)

	tt := NewTester("https://shows.youtube.com", "show.youtube.com", proxies)
	tt.Test()
	res, err := tt.GetResult()
	if err != nil {
		return
	}
	for k, v := range res {
		fmt.Printf("%s : %f\n", k, v)
	}
}
