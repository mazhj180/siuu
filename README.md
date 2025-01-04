<p align="center" style="text-align: center">
	<img alt="logo" src="./docs/logo.png" height="220px" width="220px">
</p>


**there are no any proxy serve, "siuu" is just a local proxy app**


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

git clone https://github.com/siuu/siuu.git # download the source code
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
- siuu-cli: the cli tool of siuu
- siuu: the app of siuu

### How to config?

**init configuration file:**
```toml
# you don't need to edit the init configuration file 
[log]
path = "./siuu/log/" # log path; no change recommended
level.system = "DEBUG" # system log level
level.proxy = "INFO"   # proxy log level

[server]
port = 17777        # server listen port
http.port = 18888   # http proxy listen port
socks.port = 19999  # socks5 proxy listen port

# routing related configurations
[router]
enable = true      # whether to turn on the router 
path.table = "~/.siuu/conf/pr.toml"  # routing table config file path
path.xdb = "~/.siuu/conf/ip2region.xdb"  # xdb file path

[proxy]
path = "~/.siuu/conf/pr.toml"   # proxy config file path
```

**proxies and routing table configuration file:**
```toml
# proxy config
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




# route table config
[route]

# First priority is excat match
# Second priority is wildcard match
# Third priority is geo match
# If there is no matching rule, it will not use any proxies

# If you want to use the default proxy,you can set the proxy name to default.
# You can set the default proxy as you like
# eg : [xxxxx],default

# exact match rules
# Example: [domain],[proxy name]
exacts = [

    # bing
    "www.cn.bing.com,direct",
    "cn.bing.com,direct",

    # openai chatgpt
    "openai.com,xxxxx",
    "chat.openai.com,xxxxx",
    "www.openai.com,xxxxxxx",
    "ios.chat.openai.com,xxxxx",
    "ab.chatgpt.com,xxxxxx",

    # github
    "github.com,xxxxx",

    # google
    "www.google.com,default",
    "google.com,default",
    "imap.gmail.com,default", 
    "content-autofill.googleapis.com,default",

]


# wildcard match rules
# Example: [*domain],[proxy name]
wildcards = [
    "*github.com,default",
    "*cn,direct",
]

# geo match rules
# Example: [country/region/city],[proxy name]
geo = [
    "新加坡,xxxxx",
]
```

### How to start it?

```bash
./siuu-cli proxy on/off    # turn on/off the global proxy
./siuu-cli start/stop      # start/stop the app
```
