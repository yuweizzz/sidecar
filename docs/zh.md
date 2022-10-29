# Sidecar

基于 MITM 的 Https 代理服务器，用于站点访问加速，是 [dev-sidecar](https://github.com/docmirror/dev-sidecar) 的 golang 实现。

## 使用方法

支持系统包括 Linux ， macOS 和 Windows 。

运行服务时需要 `conf.toml` 文件：

``` bash
# conf.toml
[Sidecar]
ProxyPort = 4396
OnlyListenIPv4 = true
LogLevel = "info"

[RemoteProxy]
Server = "remote.server.com"  # 使用时需要替换为实际的远端服务域名
ComplexPath = "FreeFreeFree"  # 使用时需要替换为实际的远端服务入口路径
GfwListUrl = "https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt"

[RemoteProxy.CustomHeaders]
# 使用时需要替换为实际的远端服务允许通过的 Header
AuthHeader = "Secret"

# 这部分信息可以生成远端服务的 nginx.conf ，仅在 create-nginx-conf 时被使用
[RemoteProxyConf]
EnableListenHTTP2 = true
EnableWebSocketProxy = true
EnableModernTLSOnly = true
NginxWorkDir = "/usr/local/openresty/nginx/logs"
SSLCertificatePath = "/usr/local/openresty/nginx/conf/proxy/proxy.crt"
SSLPrivateKeyPath = "/usr/local/openresty/nginx/conf/proxy/proxy.pri"
```

修改 `conf.toml` 文件后存放到对应目录，启动服务时会自动生成对应的证书，信任并安装根证书后就可以正常使用。

通过 `sidecar-server start [-conf tomlfile] [-dir workdir] [-daemon]` 启动服务，通过 `sidecar-server stop [-dir workdir]` 停止服务。

不指定 `-conf` 则配置文件必须命名为 `conf.toml` 并且处于当前可执行文件所在的目录；不指定 `-dir` 则默认工作目录为当前可执行文件所在的目录，生成的证书和服务运行时的锁文件都会处于这个目录中；不指定 `-daemon` 则服务会占用当前终端，否则会运行在后台中。

### Windows

在启动服务后，需要手动设置代理服务，或者使用提供的 bat 脚本。

### Linux

在启动服务后，需要声明环境变量：

``` bash
export HTTPS_PROXY=127.0.0.1:4396
```

## 感谢

项目代码参考了以下项目：

- [dev-sidecar](https://github.com/docmirror/dev-sidecar)
- [go-mitmproxy](https://github.com/lqqyt2423/go-mitmproxy)
- [koko](https://github.com/jumpserver/koko) 
