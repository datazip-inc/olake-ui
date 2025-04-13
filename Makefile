whoami=$(shell whoami)
home=$(shell echo $$HOME)
GIT_VERSION=$(shell git describe --tags `git rev-list --tags --max-count=1`)
GIT_COMMITSHA=$(shell git rev-list -1 HEAD)
LDFLAGS="-w -s -X github.com/datazip/olake-server/constants.version=${GIT_VERSION} -X github.com/datazip/olake-app/constants.commitsha=${GIT_COMMITSHA} -X github.com/datazip/olake-app/constants.releasechannel=${RELEASE_CHANNEL}"
GOPATH = $(shell go env GOPATH)


## Lint check.
golangci:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest;
	cd server; $(GOPATH)/bin/golangci-lint run

build:
	gofmt -l -s -w .
	cd server; go build -ldflags=${LDFLAGS} -o olake-server main.go

gofmt:
	gofmt -l -s -w .
	
run:
	cd server; go mod tidy; \
	bee run;

run-build:
	./olake-server

restart: build run-build

pre-commit:
	chmod +x $(shell pwd)/.githooks/pre-commit
	chmod +x $(shell pwd)/.githooks/commit-msg
	git config core.hooksPath $(shell pwd)/.githooks