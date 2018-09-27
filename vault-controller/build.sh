#!/bin/bash
export GOPATH=/go
export BUILD=/build 
export PATH=$PATH:$GOPATH/bin

if [ ! -d $GOPATH ]
then
  mkdir $GOPATH
fi

go get https://github.com/pearsontechnology/bitesize-controllers
cd $GOPATH/src/github.com/pearsontechnology/bitesize-controllers/vault-controller
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build controller.go
