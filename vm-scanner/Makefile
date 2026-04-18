# VM Scanner Makefile

.PHONY: build test clean fmt vet lint deps help fix imports

# Variables
BINARY_NAME=vm-scanner
BUILD_DIR=build

# Default target
all: deps fmt vet test build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/main.go

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 cmd/main.go
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 cmd/main.go
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe cmd/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code (Go equivalent of Black)
fmt:
	@echo "Formatting code (go fmt)..."
	go fmt ./...

# Organize imports (Go equivalent of isort)
imports:
	@echo "Organizing imports (goimports)..."
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not found, installing..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
		goimports -w .; \
	fi

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run golangci-lint (Go equivalent of Flake8)
lint:
	@echo "Running linter (golangci-lint)..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

# Type checking (Go equivalent of MyPy)
typecheck:
	@echo "Running type checking (go vet + staticcheck)..."
	go vet ./...
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "staticcheck not found, installing..."; \
		go install honnef.co/go/tools/cmd/staticcheck@latest; \
		staticcheck ./...; \
	fi

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Fix all code issues (Go equivalent of Black + Flake8 + MyPy + isort)
fix: imports fmt typecheck lint
	@echo "✅ All code fixes applied!"

# Install dependencies for development
install-deps:
	@echo "Installing development dependencies..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest

# Run the application (example)
run:
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME) -help

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage"
	@echo "  fix          - Fix all code issues (imports + fmt + typecheck + lint)"
	@echo "  imports      - Organize imports (goimports)"
	@echo "  fmt          - Format code (go fmt)"
	@echo "  typecheck    - Type checking (go vet + staticcheck)"
	@echo "  lint         - Run linter (golangci-lint)"
	@echo "  vet          - Run go vet"
	@echo "  deps         - Download dependencies"
	@echo "  clean        - Clean build artifacts"
	@echo "  install-deps - Install development dependencies"
	@echo "  run          - Run the application"
	@echo "  help         - Show this help"