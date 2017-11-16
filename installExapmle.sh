#!/bin/bash
export GOPATH=`pwd`
go get github.com/mgutz/logxi/v1
go install ./src/peer
cp bin/peer example/client0
cp bin/peer example/client1
cp bin/peer example/client2
cp bin/peer example/client3
cp bin/peer example/client4

