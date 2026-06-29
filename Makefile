.PHONY: run mcp mcp-build web web-install build test vet tidy

run:
	go run ./cmd/api

mcp:
	go run ./cmd/mcp

mcp-build:
	go build -o bin/pm-mcp ./cmd/mcp

web-install:
	npm --prefix web install

web:
	npm --prefix web run dev

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

tidy:
	go mod tidy
