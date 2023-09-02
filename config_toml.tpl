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
