#!/usr/bin/env bash

CURDIR=`pwd`
OLDGOPATH="$GOPATH"
export GOPATH="$CURDIR:{{range .Depends}}{{.}}:{{end}}"

gofmt -tabs=false -tabwidth=4 -w src

go install {{.Name}}

export GOPATH="$OLDGOPATH"

echo 'finished'
