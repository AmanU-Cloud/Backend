.PHONY: generate-mocks
generate-mocks:
	@echo "Setting up mock generation with go uber mock"
	@echo "Checking mockgen tool..."
	@go run go.uber.org/mock/mockgen@latest -version
	@echo ""
	@echo "TODO: Add specific mockgen commands when interfaces are defined in packages"
	@echo "Ready to generate mocks for future interfaces"

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build -o bin/server ./cmd