.PHONY: build lint test run clean help

BINARY_NAME=reviewer
BUILD_DIR=bin
CMD_PATH=cmd/reviewer/main.go

help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  lint           - Run golangci-lint"
	@echo "  test           - Run tests with verbose output"
	@echo "  test-race      - Run tests with race detector"
	@echo "  run            - Run the application"
	@echo "  clean          - Clean build artifacts"

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)

lint:
	golangci-lint run ./...

test:
	go test -v ./...

test-race:
	go test -race -v ./...

run:
	go run $(CMD_PATH)

clean:
	rm -rf $(BUILD_DIR)

.DEFAULT_GOAL := help
