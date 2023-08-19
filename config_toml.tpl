[Client]
ProxyPort = 4396
OnlyListenIPv4 = true
RunAsDaemon = true  # 配置是否后台运行，后台运行会将日志输出到文件，前台运行会将日志输出到控制台
Mode = "HTTPS"  # 支持 "HTTPS" 和 "WSS" ，应该根据对应的服务端来配置，默认使用 HTTPS
WorkDir = ""  # 工作目录，默认为可执行文件所在的当前目录
PriKeyPath = ""  # 证书路径，初次运行会自动生成，默认为 WorkDir 下的 sidecar-client.pri
CertPath = ""  # 私钥路径，初次运行会自动生成，默认为 WorkDir 下的 sidecar-client.crt
LogLevel = "info"
Resolver = "1.1.1.1"  # 指定 DNS 服务器，可以来避免某些 DNS 污染问题
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
