@echo off

:: reg add "HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings" /v ProxyEnable /t REG_DWORD /d 0 /f > NUL

cmd /c %cd%/sidecar.exe client -action stop

exit
