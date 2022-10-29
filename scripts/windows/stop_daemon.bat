@echo off

echo "Disable Proxy Server ......"
reg add "HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings" /v ProxyEnable /t REG_DWORD /d 0 /f

echo "Stop Sidecar Server ......"
%cd%/sidecar-server.exe stop

echo "Press Any Key To Close This Window ......"
pause
exit
