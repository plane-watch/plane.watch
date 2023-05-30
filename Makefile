export GOBIN=$(shell pwd)/bin
export GOAMD64=v3

.PHONY: all

all: tidy vet test build

tidy:
	go mod tidy
	go mod verify

build:
	go install ./...

vet:
	go vet ./...

test:
	go test ./...

lint:
	docker pull golangci/golangci-lint:latest
	docker run -t --rm -v `pwd`:/app -w /app golangci/golangci-lint:latest golangci-lint run -v

leakcheck:
	docker pull ghcr.io/gitleaks/gitleaks:latest
	docker run -it --rm -w /app -v `pwd`:/app ghcr.io/gitleaks/gitleaks:latest detect --source="/app" --report-path gitleaks-report.json

race:
	go install -race ./...

clean:
	rm bin/*

build-doc:
	docker run --rm -it -v ${PWD}/docs:/app/docs asyncapi/generator docs/pw_ws_broker.async-api.yaml @asyncapi/html-template -o docs --force-write
