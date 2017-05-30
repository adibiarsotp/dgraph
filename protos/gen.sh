#!/bin/bash

protos=$GOPATH/src/gopkg.in/adibiarsotp/dgraph.vo/protos
pushd $protos > /dev/null
protoc --gofast_out=plugins=grpc:. -I=. *.proto
