package router

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"siuu/server/config/constant"
	P "siuu/server/config/proxies"
	"siuu/server/config/rules"
	"siuu/tunnel/logger"
	"siuu/tunnel/proxy"
	"siuu/tunnel/routing"
	"siuu/util"
	"strings"
)

func NewDefaultRouter() (*routing.BasicRouter, error) {

	var loader routing.Loader
	loader = func(r routing.Router) error {
		dr, ok := r.(*routing.BasicRouter)
		if !ok {
			return fmt.Errorf("router [%s] is not DefaultRouter", r.Name())
		}

		var err error
		if err = LoadRules(dr); err != nil {
			return err
		}

		LoadProxy(dr)

		return nil
	}

	r := &routing.BasicRouter{DefaultProxy: &proxy.DirectProxy{}}

	return r, r.Boot(loader)
}

func LoadRules(router *routing.BasicRouter) error {

	if router == nil {
		return rules.NoRouterErr
	}

	paths := util.GetConfigSlice(constant.RuleRoutePath)

	v := viper.New()
	v.SetConfigType("toml")

	for _, path := range paths {
		path = util.ExpandHomePath(path)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		fin, err := os.OpenFile(path, os.O_RDONLY, 0666)
		if err != nil {
			continue
		}
		defer fin.Close()

		if err = v.ReadConfig(fin); err != nil {
			return err
		}

		ruls := v.GetStringSlice("route.rules")
		for _, ru := range ruls {
			if rul, e := rules.ParseRule(ru); e == nil {
				router.Rules[rul.ID()] = rul
				router.RuleIds = append(router.RuleIds, rul.ID())
			}
		}

	}

	return nil
}

func LoadProxy(r *routing.BasicRouter) {

	filepath := util.GetConfigSlice(constant.RuleProxyPath)

	v := viper.New()
	v.SetConfigType("toml")

	for _, f := range filepath {
		f = util.ExpandHomePath(f)

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

		if err = v.ReadConfig(fin); err != nil {
			logger.SWarn("failed to initialize proxy [%s]", f)
			continue
		}

		alias := v.GetStringSlice("proxy.alias")
		for _, ali := range alias {
			parts := strings.SplitN(ali, ":", 2)

			str := parts[1]
			str = strings.TrimPrefix(str, "[")
			str = strings.TrimSuffix(str, "]")

			keys := strings.Split(str, ",")
			for _, k := range keys {
				r.ProxyAlias[k] = parts[0]
			}
		}

		prxStr := v.GetStringSlice("proxy.proxies")

		for _, p := range prxStr {
			prx, err := P.ParseProxy(p)
			if err != nil {
				logger.SWarn("failed to initialize proxy [%s]", f)
				continue
			}

			r.Proxies[prx.Name()] = prx
			r.ProxyNames = append(r.ProxyNames, prx.Name())

		}
	}
}
