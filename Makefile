BINARY_NAME = users
VERSION=$(shell git describe --exact-match --tags HEAD 2>/dev/null || git rev-parse --short HEAD 2>/dev/null)
ifeq ($(VERSION),)
	VERSION = Undefined
endif
LDFLAGS = -ldflags "-X main.version=$(VERSION)"

## help: print the help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## install-tools: installs required tools such as goimports and golangci-lint
.PHONY: install-tools
install-tools:
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest

## tidy: runs additional commands/tools to tidy the codebase
.PHONY: tidy
tidy:
	go mod tidy
	goimports -w .

## test: runs tests
.PHONY: test
test:
	go test -v -race ./...

## test-coverage: runs tests and shows test coverage profile
.PHONY: test-coverage
test-coverage:
	go test -v -race -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

## lint: runs golangci-lint
.PHONY: lint
lint:
	golangci-lint run ./...

## build: builds the application/service
.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME).bin cmd/$(BINARY_NAME)/main.go

## generate Go models from OpenAPI specification via oapi-codegen
oapi-codegen:
	oapi-codegen -package generated -generate=types ports/http/docs/openapi.yaml > ports/http/generated/service.go
