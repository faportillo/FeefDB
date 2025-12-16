SHELL := /bin/bash

.PHONY: proto tidy build test lint run

proto:
	@rm -rf gen
	@buf generate

tidy:
	@go mod tidy

build:
	@go build ./...

test:
	@go test ./...

lint:
	@golangci-lint run ./...

run:
	@go run ./cmd/server

