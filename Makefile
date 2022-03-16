export GOBIN=$(shell pwd)/bin
export GOAMD64=v3

.PHONY: all

all: vet test build

build:
	go install ./...

vet:
	go vet ./...

test:
	go test ./...

race:
	go install -race ./...
