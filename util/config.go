package util

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path"
	"reflect"
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

func SetConfig(key string, value any) error {
	v := CreateConfig("conf", "toml")
	if !v.IsSet(key) {
		return fmt.Errorf("key %s is not set", key)
	}
	v.Set(key, value)
	if err := v.WriteConfig(); err != nil {
		return err
	}
	return nil
}

func GetConfig[T ~int64 | ~string | ~bool | ~[]int64 | ~[]string](key string) T {
	var zero T
	v := CreateConfig("conf", "toml")
	if v.Get(key) == nil {
		return zero
	}
	return v.Get(key).(T)
}

func GetSettings() []string {
	var res []string
	v := CreateConfig("conf", "toml")
	settings := v.AllSettings()
	var dfs func(any, string)
	dfs = func(s any, str string) {
		if reflect.TypeOf(s).Kind() != reflect.Map {
			str += fmt.Sprintf("=%v", s)
			res = append(res, str)
			return
		}

		for k, value := range s.(map[string]any) {
			var c string
			if str == "" {
				c = k
			} else {
				c = str + "." + k
			}
			dfs(value, c)
		}
	}
	dfs(settings, "")
	return res
}
