package config

import (
	"os"
	"path"
	"siuu/logger"
	"siuu/server/config/constant"
	"siuu/util"
)

type Config struct {
	ServerPort     uint16
	HttpProxyPort  uint16
	SocksProxyPort uint16
	EnablePProf    bool
	EnabledRule    bool
	Model          constant.Model
}

func InitConfig(interactive bool) Config {
	v := util.CreateConfig("conf", "toml")

	// server
	p1 := v.GetUint16(constant.ServerPort)
	p2 := v.GetUint16(constant.ProxyHttpPort)
	p3 := v.GetUint16(constant.ProxySocksPort)
	enablePprof := v.GetBool(constant.EnablePProf)

	prxModel := v.GetString(constant.ProxyModel)
	model := constant.NORMAL
	if prxModel == "TUN" {
		model = constant.TUN
	}

	// logger
	if !interactive {
		logPath := v.GetString(constant.LogDirPath)
		logPath = util.ExpandHomePath(logPath)
		logger.InitSystemLog(path.Dir(logPath)+"/system.log", 10*logger.MB, logger.LogLevel(v.GetString(constant.SystemLogLevel)))
		logger.InitProxyLog(path.Dir(logPath)+"/proxy.log", 1*logger.MB, logger.LogLevel(v.GetString(constant.ProxyLogLevel)))
	}

	// rule
	enabledRule := v.GetBool(constant.RuleEnabled)

	return Config{
		ServerPort:     p1,
		HttpProxyPort:  p2,
		SocksProxyPort: p3,
		EnablePProf:    enablePprof,
		EnabledRule:    enabledRule,
		Model:          model,
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
