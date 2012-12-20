#!/usr/bin/env bash

if [ ! -f install.sh ]; then
    echo 'install.sh must be run within its container folder' 1>&2
    exit 1
fi

CURDIR=`pwd`
OLDGOPATH="$GOPATH"
export GOPATH="$CURDIR:{{range .Depends}}{{.}}:{{end}}"

# 打开代码格式化可能会导致监控两次
# gofmt -tabs=false -tabwidth=4 -w src

go {{.GoWay}} {{.Options}} {{.MainFile}}

export GOPATH="$OLDGOPATH"

echo 'finished'