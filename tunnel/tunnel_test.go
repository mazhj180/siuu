package tunnel

import (
	"fmt"
	"siuu/tunnel/proxy"
	"siuu/tunnel/proxy/torjan"
	"testing"
)

func TestPing(t *testing.T) {
	prx := &torjan.Proxy{
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

	prx2 := &torjan.Proxy{
		Type:     proxy.TROJAN,
		Name:     "aaaaaaaaa",
		Server:   "Cadaxxxd.org",
		Port:     uint16(3306),
		Password: "xaaaaassss",
		Protocol: proxy.TCP,
		Sni:      "aaaaaxxxx",
	}
	_ = []proxy.Proxy{prx, prx1, prx2}

	ping, err := T.Ping(prx1)
	if err != nil {
		t.Fatalf("ping fail: %s\n", err)
	}
	fmt.Printf("%v\n", ping.Delay)
}
