#!/bin/bash

# This script is used to compile dgraph and dgraph-live-loader with build flags.
# The build flags are useful in finding information about the binary.

release_version="$(git describe --abbrev=0)-dev";
lastCommitSHA1=$(git rev-parse --short HEAD);
gitBranch=$(git rev-parse --abbrev-ref HEAD)
lastCommitTime=$(git log -1 --format=%ci)
dgraph_cmd=$GOPATH/src/gopkg.in/adibiarsotp/dgraph.v82/cmd;

release="gopkg.in/adibiarsotp/dgraph.v82/x.dgraphVersion"
branch="gopkg.in/adibiarsotp/dgraph.v82/x.gitBranch"
commitSHA1="gopkg.in/adibiarsotp/dgraph.v82/x.lastCommitSHA"
commitTime="gopkg.in/adibiarsotp/dgraph.v82/x.lastCommitTime"

echo -e "\033[1;33mBuilding binaries\033[0m"
echo "dgraph"
cd $dgraph_cmd/dgraph && \
   go build -ldflags \
   "-X $release=$release_version -X $branch=$gitBranch -X $commitSHA1=$lastCommitSHA1 -X '$commitTime=$lastCommitTime'" ;

echo "dgraph-live-loader"
cd $dgraph_cmd/dgraph-live-loader && \
   go build -ldflags \
   "-X $release=$release_version -X $branch=$gitBranch -X $commitSHA1=$lastCommitSHA1 -X '$commitTime=$lastCommitTime'" .;

echo "dgraph-bulk-loader"
cd $dgraph_cmd/dgraph-bulk-loader && \
   go build -ldflags \
   "-X $release=$release_version -X $branch=$gitBranch -X $commitSHA1=$lastCommitSHA1 -X '$commitTime=$lastCommitTime'" .;


