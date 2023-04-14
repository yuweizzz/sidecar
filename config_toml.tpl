[Client]
ProxyPort = 4396
OnlyListenIPv4 = true
RunAsDaemon = true
# 工作目录，默认为可执行文件所在的当前目录
WorkDir = ""
# 证书路径，初次运行会自动生成，默认为 WorkDir 下的 sidecar-client.pri
PriKeyPath = ""
# 私钥路径，初次运行会自动生成，默认为 WorkDir 下的 sidecar-client.crt
CertPath = ""
LogLevel = "info"
GfwListUrl = "https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt"
CustomProxyHosts = [
    "github.com",
]

[[Client.RemoteServers]]
# 远端服务的主机名，需要替换为实际对应的域名
Host = "remote.server.hostname"
# 混杂路径，需要替换为实际的远端入口路径
ComplexPath = "XxXxXxXxX"

# 需要替换为远端服务允许通过的 Header
[Client.RemoteServers.CustomHeaders]
AuthHeader = "Secret"

[Server]
ServerPort = 443
OnlyListenIPv4 = true
RunAsDaemon = true
# 工作模式，目前暂时只有 HTTPS
Mode = "HTTPS"
# 工作目录，默认为可执行文件所在的当前目录
WorkDir = ""
# 远端服务的证书路径，必须自行填写
PriKeyPath = ""
# 远端服务的私钥路径，必须自行填写
CertPath = ""
LogLevel = "info"
# 混杂路径，可以自定义的入口路径
ComplexPath = "XxXxXxXxX"

# 用于认证流量，可以有多个 Header
[Server.CustomHeaders]
AuthHeader = "Secret"

# 可以用于生成远端服务的 nginx.conf 
# This part just use for create nginx.conf.
# You can delete part block if you don't need to create nginx config.
[Server.NginxConf]
EnableListenHTTP2 = true
EnableWebSocketProxy = true
EnableModernTLSOnly = true
NginxWorkDir = "/usr/local/openresty/nginx/logs"
SSLCertificatePath = "/usr/local/openresty/nginx/conf/proxy/proxy.crt"
SSLPrivateKeyPath = "/usr/local/openresty/nginx/conf/proxy/proxy.pri"
