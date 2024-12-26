package test

import (
	"encoding/json"
	"fmt"
	proxy2 "siu/tunnel/proxy"
	"siu/util"
	"testing"
	"time"
)

func init() {
	v := util.CreateConfig("conf", "toml")
	proxy2.Init(v)
}

func TestAddPrx(t *testing.T) {

	data := `{
		"Name": "proxy3",
		"Type": "trojan",
		"Server": "xxxx.com",
		"Port": 9120,
		"Password": "asdqweqsdasa",
		"Protocol": "tcp",
		"Sni": "sxxasd"
	}`

	var tp proxy2.TrojanProxy
	err := json.Unmarshal([]byte(data), &tp)
	if err != nil {
		t.Error(err)
	}

	_ = proxy2.AddProxies(&tp)

	for _, v := range proxy2.GetProxies() {
		fmt.Println(v)
	}

}

func TestReadPrx(t *testing.T) {
	for _, v := range proxy2.GetProxies() {
		fmt.Println(v)
	}
}

func TestRemovePrx(t *testing.T) {
	proxy2.RemoveProxies("proxy1")
	for _, v := range proxy2.GetProxies() {
		fmt.Println(v)
	}
	time.Sleep(2 * time.Second)
}
