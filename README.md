<p align="center" style="text-align: center">
	<img alt="logo" src="./docs/logo.png" height="220px" width="220px">
</p>


**Siuu is a local proxy app. It does not provide any remote proxy services.**

**Notice**: `siuucli` is deprecated. Please use [siuu-alfredworkflow](https://github.com/mazhj180/siuu-alfredworkflow.git) instead.

## About
- implements local proxy clients: http, socks, shadow (Shadowsocks), trojan
- extensible: you can add custom proxy protocols and routers
- runs as a user-level daemon or a normal program
- stable for daily use; improvements will continue
- open source; contributions are welcome

## Install
Recommended Go version: 1.23.0+

### Download from Release
[Latest release](https://github.com/siuu/siuu/releases/latest)

### Build from source

Run as a daemon service
```bash
git clone https://github.com/mazhj180/siuu.git
cd siuu
go build -o siuu .
# You can move the app anywhere before running "./siuu install"
./siuu install  # register the daemon service (do not move files afterward)
./siuu start    # start the daemon service
```

Run as a normal program
```bash
git clone https://github.com/mazhj180/siuu.git
cd siuu
go build -o siuu .
./siuu
```

## Usage
At first run, Siuu creates a directory named `.siuu` in your home directory containing:
- conf/: configuration files
- log/: log files
- siuu: the app binary
- siuucli: deprecated

### Configuration

Core configuration (`conf/conf.toml`):
```toml
# you don't need to edit the init configuration file 


# Log configuration
[log]

# - path: log file path, using '~' means the home directory of user
# - level.system: log level of system, including error, warn, info, debug
# - level.proxy: log level of proxy, including  error, warn, info, debug
path = '~/.siuu/log/'
level.system = 'DEBUG'
level.proxy = 'INFO'



# Server configuration
[server]

# - port: the main port on which the server listens for control command connections
port = 17777



# listens for pprof ,the port is 6060
[server.pprof]
enable = false
port = 6060


# proxy configuration
[server.proxy]

# - model: proxy model, including system, tun. The default value is system
# - tables: routing tables configuration file path
mode = "system"
tables = [
    "~/.siuu/conf/route_table.toml",
]


# os http proxy service configuration
# if the tun mode is enabled, the following configuration regarding the os proxy will be ignored.
[server.proxy.http]

# - http.enable: if open the http proxy
# - http.port: the port dedicated for HTTP proxy connections
enable = true
port = 18888

# os socks proxy service configuration
[server.proxy.socks]

# - socks.enable: if open the socks proxy
# - socks.port: the port dedicated for SOCKS5 proxy connections
enable = true
port = 19999
```

Proxies and routing table configuration (`~/.siuu/conf/route_table.toml`):
```toml
# You can add your proxies here
# Supported schemes: http, socks, shadow (Shadowsocks), trojan
#
# Format: <scheme>://<host>:<port>?name=<proxyName>&[params]
# Common params:
# - name: required, unique proxy name
# - t: optional traffic type, one of tcp/udp (default: tcp)
# - mux: optional multiplexer, one of none/smux/yamux (default: none)
#
# Scheme-specific params:
# - http: no extra params required beyond name
# - socks: username, password are required
# - shadow: cipher, password are required (e.g. AEAD_AES_128_GCM)
# - trojan: password, sni are required
#
# Examples:
# - http://proxy.example.com:8080?name=h1&t=tcp&mux=none
# - socks://socks.example.com:1080?name=s5&username=user&password=pass&t=tcp
# - shadow://ss.example.com:8388?name=ss1&cipher=AEAD_AES_128_GCM&password=secret&t=tcp&mux=smux
# - trojan://trojan.example.com:443?name=p1&password=123456&sni=sni.siuu.com&t=tcp&mux=none
proxies = [
    # 'trojan://siuu.com:443?name=p1&password=123456&sni=sni.siuu.com&t=tcp&mux=none',
]


# Mapping tags to proxies
# Format: '<proxyName:[tag1,tag2]>'
# You can reference these tags as the rule target value, they will resolve to the mapped proxy.
mappings = [
    # 'p1:[openai,google]',
]


# Routing rules
# - Rule format: '<type>,<key>,<value>'
#   - type: domain | ip | special
#   - key: for domain/ip, supports exact host like 'github.com' or wildcard like '*.google.com'
#   - value: proxy name or tag from mappings; use 'direct' to bypass proxy
# - Priority: exact match > wildcard match > default outlet (if configured); otherwise direct
# - Note: default outlet cannot be set in this file; it can be set at runtime via API

rules = [
    # 'domain,*.google.com,p1'
    # 'domain,www.cn.bing.com,direct',
    # 'ip,192.168.4.73,p1',
    # 'domain,github.com,openai',
]
```

### Start/Stop

```bash
./siuu start     # start the app (daemon mode after install)
./siuu stop      # stop the app
```

## REST API (Router Control)
- GET `/api/router/clients`: list proxies, mappings, rules, default outlet
- POST `/api/router/set/mappings`: set a mapping
  - body: `{ "mapping_name": "<tag>", "proxy_name": "<proxy>" }`
- GET `/api/router/set/default_outlet?proxy_name=<name>`: set default outlet

## Logs
Logs are written to the directory configured by `log.path` (default: `~/.siuu/log/`).

## Troubleshooting
- cannot connect: verify proxy URLs, required params, and network reachability
- route not taking effect: check rule order and wildcard vs exact matches
- default outlet not used: set via API; it cannot be set in route table file

## License
MIT
