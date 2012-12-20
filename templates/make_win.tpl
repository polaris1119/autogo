@echo off

setlocal

if exist install.bat goto ok
echo install.bat must be run from its folder
goto end

:ok

set OLDGOPATH=%GOPATH%
set GOPATH=%~dp0;{{range .Depends}}{{.}};{{end}}

::打开代码格式化可能会导致监控两次
::gofmt -tabs=false -tabwidth=4 -w src

go {{.GoWay}} {{.Options}} {{.MainFile}}

set GOPATH=%OLDGOPATH%

:end
echo finished
