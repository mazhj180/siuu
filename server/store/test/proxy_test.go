package test

import (
	"encoding/json"
	"fmt"
	"siu/server/store"
	"siu/tunnel/proxy"
	"siu/util"
	"testing"
)

func init() {
	v := util.CreateConfig("conf", "toml")
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

	_ = store.AddProxies(&tp)

	for _, v := range store.GetProxies() {
		fmt.Println(v)
	}

}
