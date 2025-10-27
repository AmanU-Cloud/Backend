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
	@echo "Доступные цели:"
	@echo "  make build    — Собрать бинарник"
	@echo "  make run      — Запустить приложение"
	@echo "  make test     — Запустить тесты"
	@echo "  make clean    — Удалить бинарник"
	@echo "  make help     — Показать эту справку"