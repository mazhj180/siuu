package util

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
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

func BuildConfiguration(root string) error {

	if err := os.MkdirAll(root+"/conf", os.ModePerm); err != nil {
		return err
	}

	file, err := os.OpenFile(root+"/conf/conf.toml", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var builder strings.Builder

	builder.WriteString("[log]\n")
	builder.WriteString(fmt.Sprintf("path = '%s'\n", root+"/log/"))
	builder.WriteString("level.system = 'DEBUG'\n")
	builder.WriteString("level.proxy = 'INFO'\n\n")

	builder.WriteString("[proxy]\n")
	builder.WriteString(fmt.Sprintf("path = '%s'\n", root+"/conf/pr.toml"))

	builder.WriteString("[router]\n")
	builder.WriteString("enable = true\n\n")

	builder.WriteString("[router.path]\n")
	builder.WriteString(fmt.Sprintf("table = '%s'\n", root+"/conf/pr.toml"))
	builder.WriteString(fmt.Sprintf("xdb = '%s'\n", root+"/conf/ip2region.xdb"))

	builder.WriteString("[server]\n")
	builder.WriteString(fmt.Sprintf("port = %d\n", 17777))
	builder.WriteString("[server.http]\n")
	builder.WriteString(fmt.Sprintf("port = %d\n", 18888))
	builder.WriteString("[server.socks]\n")
	builder.WriteString(fmt.Sprintf("port = %d\n", 19999))

	_, err = file.WriteString(builder.String())
	if err != nil {
		return err
	}
	return nil
}

func DownloadIp2Region(dir string) error {
	filePath := dir + "/ip2region.xdb"

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return nil
	}

	url := "https://github.com/lionsoul2014/ip2region/raw/master/data/ip2region.xdb"
	_, _ = fmt.Fprintf(os.Stdout, "Starting download of ip2region.xdb...")

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("network access is not available : %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
