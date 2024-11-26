export GO111MODULE=on

all: base api

version:
	git describe --always --tags > .version

setup:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install golang.org/x/tools/cmd/goimports@latest

docs:
	swag i -g routes.go -dir src/apis,src/common --instanceName v1

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

clean:
	rm -rf docs
