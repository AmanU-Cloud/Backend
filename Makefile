.PHONY: lint

help:
	@echo "Available targets:"
	@echo "  lint           - Run golangci-lint"

lint:
	golangci-lint run ./...

.DEFAULT_GOAL := help
