#!/usr/bin/env bash

if [ ! -f clean ]; then
    echo 'clean must be run within its container folder' 1>&2
    exit 1
fi

rm -rf bin/*
rm -rf pkg/linux_amd64/*

echo 'finished'