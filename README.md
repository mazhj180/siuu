<p align="center" style="text-align: center">
	<img alt="logo" src="./docs/logo.png" height="220px" width="220px">
</p>


**there are no any proxy serve, "siuu" is just a local proxy app**

**notice** :`siuucli is deprecated, please use` [siuu-alfredworkflow](https://github.com/mazhj180/siuu-alfredworkflow.git)  `instead`

## About
- it implemented http/https, socks5, shadowsocks and trojan proxies now. 
- you can implement any proxy protocols and routers you want.
- it can run as a user-level daemon, and you can use it by cli tool; the cli tool is doing now.
- siuu currently meets my basic needs, and subsequent improvements will be made gradually.
- You can use this program anywhere without any conditions, and if you are interested, you can contribute your code. 

## Install
the best version of golang is 1.23.0 


### Download from Release

[Download](https://github.com/siuu/siuu/releases/latest)

### Build from the Source code

**run it as a daemon service**
```bash

git clone https://github.com/mazhj180/siuu.git # download the source code
cd siuu # come to the source code directory
go build -o siuu . # build the source code
# you can move the app to anywhere before you run command "siuu install"
./siuu install # register the daemon service to os; note: if you execute this command, don't move file directories around 
./siuu start # start the daemon service; note: you kan use cli to start the daemon service
```

**run it as a normal program**
```bash 

git clone https://github.com/siuu/siuu.git
cd siuu
go build -o siuu .
./siuu # run directly
```

## Usage
the app will create a directory named ".siuu" in your home directory, when it is running at first time.
there are some files or dir in the directory:
- conf/: the init configuration file of siuu
- log/: the log file of siuu
- siuucli: the cli tool of siuu (Deprecated)
- siuu: the app of siuu

### How to config?

**init configuration file:**
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
# - http.port: the port dedicated for HTTP proxy connections
# - socks.port: the port dedicated for SOCKS5 proxy connections
port = 17777
http.port = 18888
socks.port = 19999


# - model: proxy model, including NORMAL, TUN. The default value is NORMAL
proxy.model = "NORMAL"


# listens for pprof ,the port is 6060
[server.pprof]
enable = false



# If you want match the rules, you need to configure the router and proxy
# the router section is [rule.router] and all the other configuration what about the router,
# must be in the router section, for example, [path.table] and [path.xdb]
# the proxy section is [rule.proxy]
[rule]
enable = true

[rule.route]
path = [
    '~/.siuu/conf/proxies.toml',
]

xdb = '~/.siuu/conf/ip2region.xdb'

# Proxy related configuration
[rule.proxy]
path = [
    '~/.siuu/conf/proxies.toml'
]
```

**proxies and routing table configuration file:**
```toml
# proxy config ('~/.siuu/conf/proxies.toml')
[proxy]
# You can add your proxies here
# Support proxy protocol: http/https, socks5, shadowsocks, trojan

# Example:
# Http/Https: [http/https],[name],[server],[port],[tcp/udp]
# Socks5: [socks5],[name],[server],[port],[username],[password],[tcp/udp]
# Shadowsocks: [shadow],[name],[server],[port],[cipher],[password],[tcp/udp]
# Trojan: [trojan],[name],[server],[port],[password],[sni],[tcp/udp]
proxies = [
    "trojan,xxxx,xxxxxxxx.com,8080,xxxxxxxxxxxxxxxx,xxxxxxxxx,tcp",
]

# proxy alias
alias = [
    "name:[ALADASD, adasdaw, sadasd, asdawdq, asdawd]",
]


# route table config
[route]

# First priority is excat match
# Second priority is wildcard match
# Third priority is geo match
# If there is no matching rule, it will not use any proxies

# If you want to use the default proxy,you can set the proxy name to default.
# You can set the default proxy as you like
# eg : [type],[xxxxx],default

# Example: [type],[domain],[proxy name/alias]
rules = [

    # bing
    "excat,www.cn.bing.com,direct",
    "excat,cn.bing.com,direct",

    # github
    "excat,github.com,xxxxx",
    "wildcard,*.github.com,xxxxx",

]
```

### How to start it?

```bash
./siuucli proxy on/off # (Deprecated)    turn on/off the global proxy 
./siuu start/stop      # start/stop the app
```
