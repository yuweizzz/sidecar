[Sidecar]
ProxyPort = 4396
OnlyListenIPv4 = true
LogLevel = "info"
GfwListUrl = "https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt"
CustomProxyHosts = [
    "github.com",
]

[RemoteProxy]
Server = "remote.server.com"  # 需要替换为实际的远端服务域名
ComplexPath = "FreeFreeFree"  # 需要替换为实际的远端服务入口路径

[RemoteProxy.CustomHeaders]
# 需要替换为实际的远端服务允许通过的 Header
AuthHeader = "Secret"

# 这部分信息可以生成远端服务的 nginx.conf ，仅在 create-nginx-conf 时被使用
[RemoteProxyConf]
EnableListenHTTP2 = true
EnableWebSocketProxy = true
EnableModernTLSOnly = true
NginxWorkDir = "/usr/local/openresty/nginx/logs"
SSLCertificatePath = "/usr/local/openresty/nginx/conf/proxy/proxy.crt"
SSLPrivateKeyPath = "/usr/local/openresty/nginx/conf/proxy/proxy.pri"
