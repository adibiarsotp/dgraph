#!/bin/bash

protos=$GOPATH/src/github.com/adibiarsotp/dgraph/protos
pushd $protos > /dev/null
protoc --gofast_out=plugins=grpc:. -I=. *.proto
