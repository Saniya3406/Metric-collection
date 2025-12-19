.PHONY: build run test

BINARY=metric-agent

build:
    go build -o $(BINARY) ./cmd/agent

run: build
    ./$(BINARY)

test:
    go test ./... -v
