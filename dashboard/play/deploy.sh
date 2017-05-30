#!/bin/bash

# TODO - Maybe change it to run the installed version of Dgraph once we verify that Travis is uploading the assets when we push bugfixes to release branch.

# This command is used to run Dgraph on play.dgraph.io
sudo ~/go/src/gopkg.in/adibiarsotp/dgraph.vo/cmd/dgraph/dgraph --port 80 --workerport 12345 --debugmode=true --nomutations=true --ui ~/go/src/gopkg.in/adibiarsotp/dgraph.vo/dashboard/build --bindall=true 2>&1 | tee -a dgraph.log
