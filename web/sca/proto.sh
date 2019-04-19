#!/bin/bash
basedir=`cd $(dirname $0); pwd`
which protoc-gen-go &>/dev/null || for pkg in \
    google.golang.org/grpc \
        github.com/golang/protobuf/proto \
        github.com/golang/protobuf/protoc-gen-go; do
    if [ ! -d "$GOPATH/src/$pkg" ]; then
        go get $pkg
    fi
done

find ${basedir} -type f -name *.proto | while read f; do
    cd $(dirname $f)
    proto=$(basename $f)
    protoc -I/usr/local/include -I. -I${basedir} \
           -I${GOPATH}/src \
           --go_out=plugins=grpc,paths=source_relative:. \
           $proto
done
