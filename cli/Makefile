SHELL := /bin/bash

.PHONY: deps deps-cleancache lint fix format proto security upgrade build

GOCMD=go

deps: ## Install dependencies
	$(GOCMD) get -u -t -v ./...
	$(GOCMD) mod tidy

deps-cleancache: ## Clear cache in Go module
	$(GOCMD) clean -modcache

lint:
	@golangci-lint run

fix:
	@golangci-lint run --fix

format:
	gofmt -w .

proto: ## Generate Go code from protobufs
	protoc \
		--go_out=. --go_opt=paths=source_relative $(shell find ./internal/infrastructure/grpc/proto/v1 -name "*.proto") \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative $(shell find ./internal/infrastructure/grpc/proto/v1 -name "*.proto")

security:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

upgrade:
	go get -u ./...
	go mod tidy -v
	go list -m -u all

build:
	$(GOCMD) build -o build/runner ./cmd/main.go