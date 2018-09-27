#!/bin/bash
export GOPATH=/go
export BUILD=/build 
export PATH=$PATH:$GOPATH/bin

if [ ! -d $GOPATH ]
then
  mkdir $GOPATH
fi

curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

cd /build
dep ensure
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build controller.go
