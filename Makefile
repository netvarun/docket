DEPS = $(shell go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)

default: all

all: deps format build

deps:
	@echo "--> Installing build dependencies"
	@go get -d -v ./... $(DEPS)

updatedeps: deps
	@echo "--> Updating build dependencies"
	@go get -d -f -u ./... $(DEPS)

format: deps
	@echo "--> Running go fmt"
	@go fmt ./...

build: deps
	@echo "--> Building client"
	@go build -o client/client client/client.go
	@echo "--> Building registry"
	@go build -o registry/registry registry/datastore.go registry/registry.go

test: deps
	@go test ./...

testrace: deps
	@go test -race ./...

clean:
	@echo "--> Cleaning binaries"
	@go clean ./client
	@go clean ./registry