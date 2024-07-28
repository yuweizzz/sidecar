# Sidecar

基于 MITM 的 Https 代理服务器，用于站点访问加速，是 [dev-sidecar](https://github.com/docmirror/dev-sidecar) 的 golang 实现。

## 使用方法

支持系统包括 Linux ， macOS 和 Windows 。

现在同时支持作为客户端或者服务端来运行，一般使用是在本地电脑运行客户端，在远程机器运行服务端。

支持两种工作模式：

* 基于 HTTPS 运行：基于 HTTPS 模式基于 MITM 实现，和 dev-sidecar 原理一致且兼容，初次运行后需要安装信任根证书。
* 基于 WSS 运行：基于 WSS 模式通过 WSS 通道实现，不需要安装信任根证书。

在使用 Nginx 自建服务端时， Sidecar 可以生成对应的 Nginx 配置，不过客户端只能使用 HTTPS 模式才能连接。目前 Sidecar 均支持两种工作模式自建服务端。

### Windows

~~通过命令行启动服务后，需要手动设置代理服务；如果使用提供的 bat 脚本，它已经包含了设置代理的步骤~~。

可以通过命令行或者 bat 脚本直接启动服务，新版本会自动修改代理配置。

### Linux

在启动服务后，需要声明环境变量：

``` bash
export HTTPS_PROXY=127.0.0.1:4396
```

### macOS

需要手动设置代理服务，原理一致，不过没有试验过。

## 使用命令

* 启动客户端： `sidecar client -action start [-conf ./config.toml]`
* 停止客户端： `sidecar client -action stop [-conf ./config.toml]`
* 启动服务端： `sidecar server -action start [-conf ./config.toml]`
* 停止服务端： `sidecar server -action stop [-conf ./config.toml]`
* 生成 nginx 配置文件： `sidecar configure-nginx [-conf ./config.toml]`

运行时需要修改 `config.toml` 文件，支持 `-conf` 重新指定配置文件路径，不指定则默认为当前目录下的 `config.toml` 。

``` bash
# config.toml example
[Client]
ProxyPort = 4396
OnlyListenIPv4 = true                                                                # 是否开启 IPv6 模式
RunAsDaemon = true                                                                   # 配置是否后台运行，后台运行会将日志输出到文件，前台运行会将日志输出到控制台
Mode = "HTTPS"                                                                       # 支持 HTTPS 和 WSS ，应该根据对应的服务端来配置，默认使用 HTTPS
WorkDir = ""                                                                         # 工作目录，留空则默认为可执行文件所在的当前目录
PriKeyPath = ""                                                                      # 证书路径，初次运行会自动生成，默认为 WorkDir 下的 sidecar-client.pri
CertPath = ""                                                                        # 私钥路径，初次运行会自动生成，默认为 WorkDir 下的 sidecar-client.crt
LogLevel = "info"                                                                    # 日志等级有四个等级： debug ， error ， info ， warn ，默认使用 info 
Resolver = "1.1.1.1"                                                                 # 指定 DNS 服务器，可以来避免某些 DNS 污染问题
GfwListUrl = "https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt"  # PAC 功能开关，注释该选项使用全局代理
CustomProxyHosts = [                                                                 # 可以用来补充没有记录在 GfwListUrl 中的域名
    "github.com",
]

[[Client.RemoteServers]]              # Client.RemoteServers 可以存在多个，但至少需要配置一个，默认启用第一个
Host = "remoteA.server.hostname"      # 远端服务的主机名，需要替换为实际对应的域名
ComplexPath = "XxXxXxXxX"             # 混杂路径，需要替换为实际的远端入口路径
[Client.RemoteServers.CustomHeaders]  # 需要替换为远端服务允许通过的 Header ，允许存在多个
AuthHeader = "Secret"

# [[Client.RemoteServers]]
# Host = "remoteB.server.hostname"
# ComplexPath = "XxXxXxXxX"
# [Client.RemoteServers.CustomHeaders]
# AuthHeader = "Secret"


[Server]
ServerPort = 443
OnlyListenIPv4 = true               # 是否开启 IPv6 模式
RunAsDaemon = true                  # 配置是否后台运行，后台运行会将日志输出到文件，前台运行会将日志输出到控制台
Mode = "HTTPS"                      # 支持 HTTPS 和 WSS ，应该根据对应的服务端来配置，默认使用 HTTPS
WorkDir = ""                        # 工作目录，留空则默认为可执行文件所在的当前目录
Resolver = "1.1.1.1"                # 指定 DNS 服务器，可以来避免某些 DNS 污染问题
PriKeyPath = "/path/to/privateKey"  # 远端服务的证书路径，必须自行填写
CertPath = "/path/to/certificate"   # 远端服务的私钥路径，必须自行填写
LogLevel = "info"                   # 日志等级有四个等级： debug ， error ， info ， warn ，默认使用 info 
ComplexPath = "XxXxXxXxX"           # 混杂路径，可以进行自定义的访问入口路径

[Server.CustomHeaders]              # 用于认证流量的 Header ，允许存在多个 Header 
AuthHeader = "Secret"


# NginxConfig 只用于生成 Nginx 配置，可以快速搭建起 Nginx 服务，提供服务给到使用 https 模式的 sidecar client ，而 sidecar 本体不依赖这部分的配置
[NginxConfig]
ServerName = ""                                                 # 应该和证书中的域名一致
ServerPort = 443                                                # 服务端口，建议保持 443 不变
OnlyListenIPv4 = true                                           # 是否监听 IPv6 地址
Location = ""                                                   # 用来混杂路径，不允许留空
Resolver = "1.1.1.1"                                            # 指定 DNS 服务器，留空则使用系统中的 DNS 服务器
SSLCertificate = "/usr/local/openresty/nginx/conf/server.crt"   # 证书存放路径，必须自行填写
SSLPrivateKey = "/usr/local/openresty/nginx/conf/server.pri"    # 私钥存放路径，必须自行填写
WorkDir = "/usr/local/openresty/nginx/logs"                     # 日志输出目录，运行时的 pid 文件也会在这个目录中
EnableListenHTTP2 = true                                        # 是否使用 HTTP2
EnableWebSocketProxy = true                                     # 是否允许代理 Websocket
EnableModernTLSOnly = true                                      # 是否使用低于 TLS v1.3 的协议版本

[NginxConfig.NginxCustomHeader]                                 # 对应 Server 模式的流量认证 Header 
AuthHeader = "Secret"
```

## 特性

### PAC

PAC 使用了 `GfwListUrl` 和 `CustomProxyHosts` 来定义代理规则，可以在 `config.toml` 中将两个配置项全部注释来启用全局代理。

只使用 `GfwListUrl` 则按照 `gfwlist.txt` 中定义的规则进行智能代理，只使用 `CustomProxyHosts` 则按照定义的二级域名列表进行代理，其余请求直接访问。

`GfwListUrl` 和 `CustomProxyHosts` 可以同时使用，但是 `gfwlist.txt` 中白名单的优先级会高于 `CustomProxyHosts` 中定义的二级域名。 `CustomProxyHosts` 可以用于补偿未被 `gfwlist.txt` 收录的规则。

### Custom DNS

可以通过 `Resolver` 来指定客户端和服务端使用的 DNS 服务器，目前测试正常，也确实可以避免一些 DNS 污染，但是由于没有缓存可能会比较慢，所以根据自己的网络环境使用延迟比较低的公共 DNS 使用体验会比较好，当然不启用这个功能直接注释掉这个选项即可，留待后续优化。

## 感谢

项目代码参考了以下项目：

- [dev-sidecar](https://github.com/docmirror/dev-sidecar)
- [go-mitmproxy](https://github.com/lqqyt2423/go-mitmproxy)
- [koko](https://github.com/jumpserver/koko) 
