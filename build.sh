#!/bin/sh

go build -a -ldflags "-linkmode external -extldflags '-static' -s -w -X main.sha1ver=`git rev-parse HEAD` -o main main.go
