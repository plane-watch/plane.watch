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

clean:
	rm bin/*

build-doc:
	docker run --rm -it -v ${PWD}/docs:/app/docs asyncapi/generator docs/pw_ws_broker.async-api.yaml @asyncapi/html-template -o docs --force-write