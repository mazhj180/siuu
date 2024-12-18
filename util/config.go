package util

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path"
	"runtime"
	"strings"
)

var (
	ProjectRootPath = path.Dir(concurrentPath()+"/../") + "/"
)

func concurrentPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return path.Dir(filename)
}

func CreateConfig(file string, fileType string) *viper.Viper {

	configPath := path.Join(ProjectRootPath, "conf/")
	config := viper.New()
	config.AddConfigPath(configPath)
	config.SetConfigName(file)
	config.SetConfigType(fileType)
	configFile := path.Join(configPath, file+"."+fileType)

	if err := config.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			panic(fmt.Errorf("cannot find the configuration file %s", configFile))
		} else {
			panic(fmt.Errorf("configuration file failed to load %s, err: %s", configFile, err))
		}
	}

	return config
}

func ExpandHomePath(config *viper.Viper, key string) string {
	p := config.GetString(key)
	if strings.HasPrefix(p, "~") {
		dir, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Errorf("cannot get user home dir: %s", err))
		}
		return strings.Replace(p, "~", dir, 1)
	}
	return p
}
