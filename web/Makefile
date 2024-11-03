export GO111MODULE=on

all: build

version:
	git describe --always --tags > .version

base: version
	cd cmds/base && go build -o ../../clbase -ldflags "-X \"main.Version=`cat ../../.version`\""

api: version
	cd cmds/api && go build -o ../../clapi -ldflags "-X \"main.Version=`cat ../../.version`\""

test:
	go test ./...

tidy:
	go mod tidy

deps:
	go build -v ./...

upgrade:
	go get -u

