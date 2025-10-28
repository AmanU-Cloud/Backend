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
	go build -o bin/reviewer cmd/reviewer/main.go

.PHONY: run
run:
	go run cmd/reviewer/main.go

.PHONY: test
test:
	go test ./...

#не работает в PowerShell - использовать Git Bash
.PHONY: clean
clean:
	rm -f bin/reviewer

.PHONY: help
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

	@echo "Доступные цели:"
	@echo "  make build    — Собрать бинарник"
	@echo "  make run      — Запустить приложение"
	@echo "  make test     — Запустить тесты"
	@echo "  make clean    — Удалить бинарник"
	@echo "  make help     — Показать эту справку"