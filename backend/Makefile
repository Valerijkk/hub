.PHONY: generate test lint build build-prod

generate:
	go generate ./ent

test:
	go test -race -count=1 ./...

lint:
	golangci-lint run ./...

build:
	go build -o bin/server ./cmd/server

# ME-09: production build with dev endpoints compiled out.
# /api/dev/* is a no-op stub; /api/dev/login + /api/dev/seed are
# physically not in the binary, regardless of DEV_MODE env var.
build-prod:
	go build -tags release -o bin/server ./cmd/server
