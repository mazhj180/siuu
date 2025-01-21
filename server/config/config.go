package config

import (
	"os"
	"path"
	"siuu/logger"
	"siuu/routing"
	"siuu/server/config/constant"
	"siuu/server/store"
	"siuu/util"
)

func InitConfig(p1, p2, p3 *uint16) {
	v := util.CreateConfig("conf", "toml")

	// server
	*p1 = v.GetUint16(constant.ServerPort)
	*p2 = v.GetUint16(constant.ProxyHttpPort)
	*p3 = v.GetUint16(constant.ProxySocksPort)

	// logger
	logPath := v.GetString(constant.LogDirPath)
	logPath = util.ExpandHomePath(logPath)
	logger.InitSystemLog(path.Dir(logPath)+"/system.log", 10*logger.MB, logger.LogLevel(v.GetString(constant.SystemLogLevel)))
	logger.InitProxyLog(path.Dir(logPath)+"/proxy.log", 1*logger.MB, logger.LogLevel(v.GetString(constant.ProxyLogLevel)))

	// proxy
	constant.Signature = make(map[string]string)
	prxPath := v.GetStringSlice(constant.ProxiesConfigPath)
	var prxPathC []string
	for _, p := range prxPath {
		val := util.ExpandHomePath(p)
		prxPathC = append(prxPathC, val)
	}
	store.InitProxy(prxPathC)

	// router
	if v.GetBool(constant.RouterEnabled) {
		routePath := v.GetStringSlice(constant.RouteConfigPath)
		var routePathC []string
		for _, r := range routePath {
			val := util.ExpandHomePath(r)
			routePathC = append(routePathC, val)
		}
		xdbPath := v.GetString(constant.RouteXdbPath)
		xdbPath = util.ExpandHomePath(xdbPath)
		routing.InitRouter(routePathC, xdbPath)
	}
}

func BuildConfiguration(root string) error {
	if err := os.MkdirAll(root+"/conf", os.ModePerm); err != nil {
		return err
	}

	conf, err := os.OpenFile(root+"/conf/conf.toml", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer conf.Close()

	if _, err = conf.Write(constant.Conf); err != nil {
		return err
	}

	prx, err := os.OpenFile(root+"/conf/proxies.toml", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer prx.Close()

	if _, err = prx.Write(constant.Proxies); err != nil {
		return err
	}

	return nil
}
