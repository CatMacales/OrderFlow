GO ?= go

.PHONY: run build test test-race fmt vet generate

run:
	$(GO) run ./cmd/orderflow

build:
	$(GO) build -o bin/orderflow ./cmd/orderflow

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

generate:
	$(GO) generate ./...
