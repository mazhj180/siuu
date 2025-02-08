package test

import (
	"fmt"
	"siuu/util"
	"testing"
)

type Log struct {
	Level string `mapstructure:"level"`
}

func TestReadConfig(t *testing.T) {
	config := util.CreateConfig("conf", "toml")
	fmt.Println(config.GetString("log.path"))
}

func TestGetAllSettings(t *testing.T) {
	r := util.GetSettings()
	for _, v := range r {
		fmt.Println(v)
	}
}

func TestDownloadIp2Region(t *testing.T) {
	err := util.DownloadIp2Region(".")
	if err != nil {
		t.Fatalf(err.Error())
	}
}
