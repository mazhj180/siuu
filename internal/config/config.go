package config

import (
	"fmt"
	"siuu/pkg/proxy/client"
	"siuu/pkg/proxy/client/http"
	"siuu/pkg/proxy/client/shadow"
	"siuu/pkg/proxy/client/socks"
	"siuu/pkg/proxy/client/trojan"
	"siuu/pkg/proxy/mux"
	"siuu/pkg/proxy/route"
	"siuu/pkg/util/net"
	"strings"
)

type SystemConfig struct {
	Log struct {
		Path  string `mapstructure:"path"`
		Level struct {
			System string `mapstructure:"system"`
			Proxy  string `mapstructure:"proxy"`
		} `mapstructure:"level"`
	} `mapstructure:"log"`

	Server struct {
		Port  int `mapstructure:"port"`
		Pprof struct {
			Enable bool   `mapstructure:"enable"`
			Port   uint16 `mapstructure:"port"`
		} `mapstructure:"pprof"`

		Proxy struct {
			Mode string `mapstructure:"mode"`

			Http struct {
				Enable bool `mapstructure:"enable"`
				Port   int  `mapstructure:"port"`
			} `mapstructure:"http"`
			Socks struct {
				Enable bool `mapstructure:"enable"`
				Port   int  `mapstructure:"port"`
			} `mapstructure:"socks"`

			Tables []string `mapstructure:"tables"`
		} `mapstructure:"proxy"`
	} `mapstructure:"server"`
}

type RouterConfig struct {
	Proxies  []string `mapstructure:"proxies"`
	Rules    []string `mapstructure:"rules"`
	Mappings []string `mapstructure:"mappings"`
}

func (r *RouterConfig) GetProxies() ([]client.ProxyClient, error) {
	prxs := make([]client.ProxyClient, 0, len(r.Proxies))

	for _, p := range r.Proxies {
		p = strings.TrimSpace(p)
		parsed, err := net.ParseURL(p)
		if err != nil {
			continue
		}

		name := parsed.Params["name"]
		if len(name) == 0 {
			return nil, fmt.Errorf("proxy name is required: %s", p)
		}

		typs := parsed.Params["t"]
		if len(typs) == 0 {
			typs = []string{"tcp"}
		}

		muxs := parsed.Params["mux"]
		if len(muxs) == 0 {
			muxs = []string{"none"}
		}
		multiplexer, _ := mux.GetMultiplexer(muxs[0])

		server := parsed.Host
		port := parsed.Port
		username := parsed.Params["username"]
		password := parsed.Params["password"]

		base := client.BaseClient{
			Server:      server,
			Port:        port,
			Mux:         multiplexer,
			TrafficType: typs[0],
		}
		var prx client.ProxyClient
		switch parsed.Scheme {
		case "trojan":
			sni := parsed.Params["sni"]
			if len(sni) == 0 {
				return nil, fmt.Errorf("sni is required: %s", p)
			}
			prx = trojan.New(base, name[0], password[0], sni[0])
		case "shadow":
			cipher := parsed.Params["cipher"]
			if len(cipher) == 0 {
				return nil, fmt.Errorf("cipher is required: %s", p)
			}
			if len(password) == 0 {
				return nil, fmt.Errorf("password is required: %s", p)
			}
			prx = shadow.New(base, name[0], cipher[0], password[0])
		case "http":
			prx = http.New(base, name[0])
		case "socks":
			if len(username) == 0 {
				return nil, fmt.Errorf("username is required: %s", p)
			}
			if len(password) == 0 {
				return nil, fmt.Errorf("password is required: %s", p)
			}
			prx = socks.New(base, name[0], username[0], password[0])
		default:
			continue
		}
		prxs = append(prxs, prx)
	}

	return prxs, nil
}

func (r *RouterConfig) GetRouterRules() []route.RouteRule {
	rules := make([]route.RouteRule, 0, len(r.Rules))

	for _, rule := range r.Rules {
		values := strings.Split(rule, ",")
		if len(values) != 3 {
			continue
		}

		rules = append(rules, route.NewRouteRule(values[0], values[1], values[2]))
	}

	return rules
}

func (r *RouterConfig) GetMappings() map[string]string {
	mappings := make(map[string]string, len(r.Mappings))

	for _, mapping := range r.Mappings {
		parts := strings.SplitN(mapping, ":", 2)

		str := parts[1]
		str = strings.TrimPrefix(str, "[")
		str = strings.TrimSuffix(str, "]")

		keys := strings.Split(str, ",")
		for _, key := range keys {
			mappings[key] = parts[0]
		}
	}

	return mappings
}
