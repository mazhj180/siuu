package tunnel

import (
	"fmt"
	"siuu/tunnel/proxy"
	"siuu/tunnel/proxy/torjan"
	"testing"
)

func TestPing(t *testing.T) {

	base := proxy.BaseProxy{
		Server:   "test.com",
		Port:     8080,
		Protocol: proxy.TCP,
	}

	prx := torjan.New(base, "xxxxzxz", "1111111", "3231311111")

	_ = []proxy.Proxy{prx}

	ping, err := T.Ping(prx)
	if err != nil {
		t.Fatalf("ping fail: %s\n", err)
	}
	fmt.Printf("%v\n", ping.Delay)
}
