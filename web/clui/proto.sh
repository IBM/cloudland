#!/bin/bash
basedir=`cd $(dirname $0); pwd`
which protoc-gen-go &>/dev/null || for pkg in \
    google.golang.org/grpc \
        github.com/golang/protobuf/proto \
        github.com/golang/protobuf/protoc-gen-go \
        github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway; do
    if [ ! -d "$GOPATH/src/$pkg" ]; then
        go get $pkg
    fi
done

find ${basedir} -type f -name *.proto | while read f; do
    cd $(dirname $f)
    proto=$(basename $f)
    protoc -I/usr/local/include -I. -I${basedir} \
           -I${GOPATH}/src \
           -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
           --go_out=plugins=grpc,paths=source_relative:. \
           --grpc-gateway_out=logtostderr=true:. \
           $proto
done
