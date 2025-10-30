.PHONY: deps generate-mocks build lint test test-race run clean help

BINARY_NAME=reviewer
BUILD_DIR=bin
CMD_PATH=cmd/reviewer/main.go

deps:
	@echo "Installing development"
	@go install go.uber.org/mock/mockgen@latest
	@echo "Mockgen installed"

generate-mocks:
	@echo "Setting up mock generation with go uber mock"
	@echo "Checking mockgen tool..."
	@go run go.uber.org/mock/mockgen@latest -version
	@echo ""
	@echo "TODO: Add specific mockgen commands when interfaces are defined in packages"
	@echo "Ready to generate mocks for future interfaces"


help:
	@echo "Available targets:"
  @echo "  deps           - Install development dependencies (mockgen)"
	@echo "  generate-mocks - Generate mocks for interfaces"
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
