.PHONY: build test clean install lint dev

BINARY_NAME=jira-go
BUILD_DIR=./build

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/jira-go

test:
	go test -v ./...

clean:
	rm -rf $(BUILD_DIR)

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

lint:
	golangci-lint run

dev:
	go run ./cmd/jira-go

.DEFAULT_GOAL := build
