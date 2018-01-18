#!/bin/bash

# Usage: source `./env.sh`

export GOPATH=`pwd`/../../../..
export PATH=$PATH:$GOPATH/bin

if [[ ! `which golint` ]]; then
	echo "Install golint..."
	go get -u github.com/golang/lint/golint
fi

make test-go-all