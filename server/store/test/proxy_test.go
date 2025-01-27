package test

import (
	"encoding/json"
	"fmt"
	"siuu/server/store"
	"siuu/tunnel/proxy/torjan"
	"testing"
)

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

	var tp torjan.Proxy
	err := json.Unmarshal([]byte(data), &tp)
	if err != nil {
		t.Error(err)
	}

	for _, v := range store.GetProxies() {
		fmt.Println(v)
	}

}
