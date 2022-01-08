#! /bin/sh
GOPATH=`pwd` go build -gccgoflags '-static' plabel.go
