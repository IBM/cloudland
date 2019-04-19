#!/bin/bash
goversion=$(go version | awk '{print $3}' | sed 's/^go//g')
major=$(echo $goversion|cut -d. -f1)
minor=$(echo $goversion|cut -d. -f2)
if [ "$major" -gt 1 -o "$minor" -gt 10 ]; then
    go env | grep ^GOROOT | cut -d'"' -f2
    exit 0
fi
# Now install go1.11.2
which wget &> /dev/null || sudo apt-get install -qq wget
cd $GOPATH/..
sudo rm -f go1.11.2.linux-amd64.tar.gz
sudo wget -q https://dl.google.com/go/go1.11.2.linux-amd64.tar.gz
sudo tar xfz go1.11.2.linux-amd64.tar.gz
GOROOT=$(cd go; pwd)
$GOROOT/bin/go env | grep ^GOROOT | cut -d'"' -f2
