.PHONY: build lint test run clean help

BINARY_NAME=reviewer
BUILD_DIR=bin
CMD_PATH=cmd/reviewer/main.go

#для отображения кирилицы в PowerShell ввести
#chcp 65001
GOBIN := $(shell go env GOPATH)/bin

.PHONY: all
all: build

.PHONY: build
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)

.PHONY: run
run:
	go run $(CMD_PATH)

.PHONY: test
test:
	go test ./...

.PHONY: test-race
test-race:
	go test -race -v ./...

#не работает в PowerShell - использовать Git Bash
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  lint           - Run golangci-lint"
	@echo "  test           - Run tests with verbose output"
	@echo "  test-race      - Run tests with race detector"
	@echo "  run            - Run the application"
	@echo "  clean          - Clean build artifacts"