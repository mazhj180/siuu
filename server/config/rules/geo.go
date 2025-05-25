package rules

import (
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"net"
	"siuu/tunnel/routing/rule"
	"strings"
)

var (
	xdbb []byte
	xdbp string
)

type GeoRule struct {
	rule.BaseRule
}

func (r *GeoRule) Match(host string) (string, bool) {

	searcher, err := xdb.NewWithVectorIndex(xdbp, xdbb)
	if err != nil {
		return "", false
	}

	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return "", false
	}

	str, err := searcher.SearchByStr(ips[0].String())
	if err != nil {
		return "", false
	}

	if r.Rule[0:1] == "!" {
		if !strings.Contains(str, r.Rule[1:]) {
			return r.Target, true
		}
	}
	if strings.Contains(str, r.Rule) {
		return r.Target, true
	}

	return "", false
}
