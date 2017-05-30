#!/bin/bash

protos=$GOPATH/src/gopkg.in/adibiarsotp/dgraph.v0/protos
pushd $protos > /dev/null
protoc --gofast_out=plugins=grpc:. -I=. *.proto
