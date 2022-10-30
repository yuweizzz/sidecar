@echo off

set PROXYSERVER=127.0.0.1:4396
set PROXYOVERRIDE=localhost;127.*;10.*;172.16.*;172.17.*;172.18.*;172.19.*;172.20.*;172.21.*;172.22.*;172.23.*;172.24.*;172.25.*;172.26.*;172.27.*;172.28.*;172.29.*;172.30.*;172.31.*;192.168.*

echo set Proxy Server ......
reg add "HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings" /v ProxyServer /t REG_SZ /d %PROXYSERVER% /f > NUL

echo set Proxy Override ......
reg add "HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings" /v ProxyOverride /t REG_SZ /d %PROXYOVERRIDE% /f > NUL

echo Enable Proxy Server ......
reg add "HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings" /v ProxyEnable /t REG_DWORD /d 1 /f > NUL

echo Start Sidecar Server ......
%cd%/sidecar-server.exe start
