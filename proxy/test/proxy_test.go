package test

import (
	"encoding/json"
	"evil-gopher/proxy"
	"evil-gopher/util"
	"fmt"
	"testing"
	"time"
)

func init() {
	v := util.CreateConfig("conf", "toml")
	proxy.Init(v)
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

	var tp proxy.TrojanProxy
	err := json.Unmarshal([]byte(data), &tp)
	if err != nil {
		t.Error(err)
	}

	_ = proxy.AddProxies(&tp)

	for _, v := range proxy.GetProxies() {
		fmt.Println(v)
	}

}

func TestReadPrx(t *testing.T) {
	for _, v := range proxy.GetProxies() {
		fmt.Println(v)
	}
}

func TestRemovePrx(t *testing.T) {
	proxy.RemoveProxies("proxy1")
	for _, v := range proxy.GetProxies() {
		fmt.Println(v)
	}
	time.Sleep(2 * time.Second)
}
