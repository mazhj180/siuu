package test

import (
	"evil-gopher/util"
	"fmt"
	"testing"
)

type Log struct {
	Level string `mapstructure:"level"`
}

func TestReadConfig(t *testing.T) {
	config := util.CreateConfig("conf", "toml")
	fmt.Println(config.GetString("log.path"))
	p := util.ExpandHomePath(config, "log.path")
	fmt.Println(p)
}
