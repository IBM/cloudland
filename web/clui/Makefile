export GO111MODULE=on

all: proto build

version:
	git describe --always --tags > .version

build: version
	go build -ldflags "-X \"main.Version=`cat .version`\""

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

