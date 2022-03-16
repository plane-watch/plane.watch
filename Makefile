export GOBIN=$(shell pwd)/bin

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
