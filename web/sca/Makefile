export GO111MODULE=on

all: proto

proto:
	./proto.sh

test:
	go test ./...

tidy:
	go mod tidy

deps:
	go build -v ./...

upgrade:
	go get -u

