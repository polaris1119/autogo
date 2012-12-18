#!/usr/bin/env bash

if [ ! -f make ]; then
    echo 'make must be run within its container folder' 1>&2
    exit 1
fi

CURDIR=`pwd`
OLDGOPATH="$GOPATH"
export GOPATH="$CURDIR:$CURDIR/../toolkits"

gofmt -tabs=false -tabwidth=4 -w src

go install server/center
go install server/register
go install server/room
go install server/saver
go install server/idgenerator

go install tester/dev_test

go install tester/dev_sample
go install tester/dev_center
go install tester/dev_register
go install tester/dev_lonp
go install tester/dev_miop
go install tester/dev_http
go install tester/dev_manager
go install tester/dev_saver
go install tester/dev_idgenerator

go install tester/ben_sample
go install tester/ben_center
go install tester/ben_register
go install tester/ben_lonp
go install tester/ben_miop
go install tester/ben_http
go install tester/ben_manager
go install tester/ben_saver
go install tester/ben_idgenerator

export GOPATH="$OLDGOPATH"

echo 'finished'
