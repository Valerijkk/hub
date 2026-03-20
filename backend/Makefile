.PHONY: generate test lint build

generate:
	go generate ./ent

test:
	go test -race -count=1 ./...

lint:
	golangci-lint run ./...

build:
	go build -o bin/server ./cmd/server
