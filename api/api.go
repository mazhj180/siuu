package api

// curl -X POST http://127.0.0.1:8080/api/router/clients

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type Data map[string]interface{}

type ProxyParams struct {
	MappingName string `json:"mapping_name"`
	ProxyName   string `json:"proxy_name"`
}

type PingProxyParam struct {
	Proxies []string `json:"proxies"`
	TestUrl string   `json:"test_url"`
}
