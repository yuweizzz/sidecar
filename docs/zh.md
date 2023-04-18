# Sidecar

基于 MITM 的 Https 代理服务器，用于站点访问加速，是 [dev-sidecar](https://github.com/docmirror/dev-sidecar) 的 golang 实现。

## 使用方法

支持系统包括 Linux ， macOS 和 Windows 。

现在同时支持作为客户端或者服务端来运行，一般使用是在本地电脑运行客户端，在远程机器运行服务端。

支持两种工作模式：基于 HTTPS 运行和基于 WSS 运行，基于 HTTPS 模式就是通过 MITM 实现，和 dev-sidecar 原理一致且兼容，初次运行后需要安装信任根证书；基于 WSS 模式通过 WSS 通道实现，可以不用安装根证书。

在使用 nginx 自建服务端时，服务可以支持 nginx 配置的生成，不过只能使用 HTTPS 模式。

### Windows

在启动服务后，需要手动设置代理服务，或者使用提供的 bat 脚本，它已经包含了设置代理的步骤。

### Linux

在启动服务后，需要声明环境变量：

``` bash
export HTTPS_PROXY=127.0.0.1:4396
```

### macOS

需要手动设置代理服务，原理一致，不过没有试验过。

## 使用命令

* 启动客户端： `sidecar -action client start [-conf ./config.toml]`
* 停止客户端： `sidecar -action client stop [-conf ./config.toml]`
* 启动服务端： `sidecar -action server start [-conf ./config.toml]`
* 停止服务端： `sidecar -action server stop [-conf ./config.toml]`
* 生成 nginx 配置文件： `sidecar -action server create-nginx-conf [-conf ./config.toml]`

运行时需要修改 `config.toml` 文件，支持 `-conf` 重新指定配置文件路径，不指定则默认为当前目录下的 `config.toml` 。

``` bash
# config.toml example
[Client]
ProxyPort = 4396
OnlyListenIPv4 = true
RunAsDaemon = true  # 配置是否后台运行，后台运行会将日志输出到文件，前台运行会将日志输出到控制台
Mode = "HTTPS"  # 支持 "HTTPS" 和 "WSS" ，应该根据对应的服务端来配置，默认使用 HTTPS
WorkDir = ""  # 工作目录，默认为可执行文件所在的当前目录
PriKeyPath = ""  # 证书路径，初次运行会自动生成，默认为 WorkDir 下的 sidecar-client.pri
CertPath = ""  # 私钥路径，初次运行会自动生成，默认为 WorkDir 下的 sidecar-client.crt
LogLevel = "info"
GfwListUrl = "https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt"  # 注释该选项使用全局代理
CustomProxyHosts = [
    # 可以用来补充没有记录在 GfwListUrl 中的域名
    "github.com",
]

# RemoteServers 可以存在多个，但至少需要配置一个，默认启用第一个
# [[Client.RemoteServers]]
# Host = "remote.server.hostname"  # 远端服务的主机名，需要替换为实际对应的域名
# ComplexPath = "XxXxXxXxX"  # 混杂路径，需要替换为实际的远端入口路径
# [Client.RemoteServers.CustomHeaders]  # 需要替换为远端服务允许通过的 Header ，允许存在多个
# AuthHeader = "Secret"

[[Client.RemoteServers]]
Host = "remote.server.hostname"
ComplexPath = "XxXxXxXxX"
[Client.RemoteServers.CustomHeaders]
AuthHeader = "Secret"

[Server]
ServerPort = 443
OnlyListenIPv4 = true
RunAsDaemon = true  # 配置是否后台运行，后台运行会将日志输出到文件，前台运行会将日志输出到控制台
Mode = "HTTPS"  # 支持 "HTTPS" 和 "WSS" ，应该根据对应的服务端来配置，默认使用 HTTPS
WorkDir = ""  # 工作目录，默认为可执行文件所在的当前目录
PriKeyPath = ""  # 远端服务的证书路径，必须自行填写
CertPath = ""  # 远端服务的私钥路径，必须自行填写
LogLevel = "info"
Host = "remote.server.hostname"  # 服务域名，只用于 nginx 配置生成
ComplexPath = "XxXxXxXxX"  # 混杂路径，可以自定义的入口路径
[Server.CustomHeaders]  # 用于认证流量的 Header ，允许存在多个
AuthHeader = "Secret"

# 只用于 nginx 配置生成
# This part just use for create nginx.conf.
# You can delete part block if you don't need to create nginx config.
[Server.NginxConf]
EnableListenHTTP2 = true
EnableWebSocketProxy = true
EnableModernTLSOnly = true
NginxWorkDir = "/usr/local/openresty/nginx/logs"
SSLCertificatePath = "/usr/local/openresty/nginx/conf/proxy/proxy.crt"
SSLPrivateKeyPath = "/usr/local/openresty/nginx/conf/proxy/proxy.pri"
```

## 特性

### PAC

PAC 使用了 `GfwListUrl` 和 `CustomProxyHosts` 来定义代理规则，可以在 `config.toml` 中将两个配置项全部注释来启用全局代理。

只使用 `GfwListUrl` 则按照 `gfwlist.txt` 中定义的规则进行智能代理，只使用 `CustomProxyHosts` 则按照定义的二级域名列表进行代理，其余请求直接访问。

`GfwListUrl` 和 `CustomProxyHosts` 可以同时使用，但是 `gfwlist.txt` 中白名单的优先级会高于 `CustomProxyHosts` 中定义的二级域名。 `CustomProxyHosts` 可以用于补偿未被 `gfwlist.txt` 收录的规则。

## 感谢

项目代码参考了以下项目：

- [dev-sidecar](https://github.com/docmirror/dev-sidecar)
- [go-mitmproxy](https://github.com/lqqyt2423/go-mitmproxy)
- [koko](https://github.com/jumpserver/koko) 
