package api

// curl -X POST http://127.0.0.1:8080/api/router/clients

type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
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

type ClientInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Server      string `json:"server"`
	Port        uint16 `json:"port"`
	TrafficType string `json:"traffic_type"`
}

type RuleInfo struct {
	Typ   string `json:"typ"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type RouterInfo struct {
	Clients       []ClientInfo      `json:"clients"`
	Mappings      map[string]string `json:"mappings"`
	Rules         []RuleInfo        `json:"rules"`
	DefaultOutlet string            `json:"default_outlet"`
}

type PingResult struct {
	Name  string `json:"name"`
	Delay string `json:"delay"`
}

type SystemConfigInfo struct {
	LogPath              string   `json:"log_path"`
	LogLevelSystem       string   `json:"log_level_system"`
	LogLevelProxy        string   `json:"log_level_proxy"`
	ServerPort           int      `json:"server_port"`
	ServerProxyHttpPort  int      `json:"server_proxy_http_port"`
	ServerProxySocksPort int      `json:"server_proxy_socks_port"`
	ServerProxyMode      string   `json:"server_proxy_mode"`
	ServerProxyTables    []string `json:"server_proxy_tables"`
	PprofEnable          bool     `json:"pprof_enable"`
	PprofPort            int      `json:"pprof_port"`
}
