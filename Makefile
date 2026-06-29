.PHONY: run mcp mcp-build build test vet tidy

run:
	go run ./cmd/api

mcp:
	go run ./cmd/mcp

mcp-build:
	go build -o bin/pm-mcp ./cmd/mcp

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

tidy:
	go mod tidy
