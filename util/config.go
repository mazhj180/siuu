package util

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path"
	"runtime"
)

var (
	ProjectRootPath = path.Dir(concurrentPath()+"/../") + "/"
	AppRootPath     = executePath() + "/"
)

func concurrentPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return path.Dir(filename)
}

func executePath() string {
	filename := GetHomeDir() + "/.siuu/conf/"
	dir := path.Dir(filename)
	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			panic(fmt.Errorf("cannot create directory %s, err: %s", dir, err))
		}
	}
	return path.Dir(dir + "/../")
}

func CreateConfig(file string, fileType string) *viper.Viper {

	configPath := path.Join(AppRootPath, "conf/")
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
