# Mobile Backend CLI Makefile

# Variables
BINARY_NAME=mobile-backend-cli
VERSION=1.0.0
BUILD_DIR=build
DIST_DIR=dist

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOINSTALL=$(GOCMD) install

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"
BUILD_FLAGS=-a -installsuffix cgo

# Default target
.PHONY: all
all: clean deps build

# Clean build artifacts
.PHONY: clean
clean:
	@echo "🧹 Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f $(BINARY_NAME)

# Install dependencies
.PHONY: deps
deps:
	@echo "📦 Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Build the CLI
.PHONY: build
build: deps
	@echo "🔨 Building $(BINARY_NAME)..."
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# Build for multiple platforms
.PHONY: build-all
build-all: deps
	@echo "🔨 Building for all platforms..."
	mkdir -p $(DIST_DIR)
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 .
	
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .
	
	# macOS ARM64
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	
	@echo "✅ Build completed for all platforms"

# Install the CLI
.PHONY: install
install: build
	@echo "📦 Installing $(BINARY_NAME)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "✅ $(BINARY_NAME) installed successfully"

# Uninstall the CLI
.PHONY: uninstall
uninstall:
	@echo "🗑️  Uninstalling $(BINARY_NAME)..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "✅ $(BINARY_NAME) uninstalled successfully"

# Run tests
.PHONY: test
test:
	@echo "🧪 Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report generated: coverage.html"

# Run linting
.PHONY: lint
lint:
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Run security scan
.PHONY: security
security:
	@echo "🔒 Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec not found. Install it with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Format code
.PHONY: fmt
fmt:
	@echo "🎨 Formatting code..."
	$(GOCMD) fmt ./...

# Run all quality checks
.PHONY: quality
quality: fmt lint security test
	@echo "✅ All quality checks passed"

# Create release packages
.PHONY: release
release: build-all
	@echo "📦 Creating release packages..."
	cd $(DIST_DIR) && \
	for file in $(BINARY_NAME)-*; do \
		tar -czf "$$file.tar.gz" "$$file"; \
	done
	@echo "✅ Release packages created in $(DIST_DIR)/"

# Run the CLI locally
.PHONY: run
run: build
	@echo "🚀 Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Show help
.PHONY: help
help:
	@echo "Mobile Backend CLI - Available commands:"
	@echo ""
	@echo "  build          Build the CLI binary"
	@echo "  build-all      Build for all platforms"
	@echo "  install        Install the CLI to /usr/local/bin"
	@echo "  uninstall      Remove the CLI from /usr/local/bin"
	@echo "  test           Run tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  lint           Run code linting"
	@echo "  security       Run security scan"
	@echo "  fmt            Format code"
	@echo "  quality        Run all quality checks"
	@echo "  release        Create release packages"
	@echo "  run            Run the CLI locally"
	@echo "  clean          Clean build artifacts"
	@echo "  deps           Install dependencies"
	@echo "  help           Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make install"
	@echo "  make test"
	@echo "  make run ARGS='--help'"

# Development targets
.PHONY: dev
dev: build
	@echo "🔧 Starting development mode..."
	./$(BUILD_DIR)/$(BINARY_NAME) --help

# Check if required tools are installed
.PHONY: check-tools
check-tools:
	@echo "🔍 Checking required tools..."
	@command -v go >/dev/null 2>&1 || { echo "❌ Go is not installed"; exit 1; }
	@command -v git >/dev/null 2>&1 || { echo "❌ Git is not installed"; exit 1; }
	@echo "✅ All required tools are installed"

# Setup development environment
.PHONY: setup
setup: check-tools deps
	@echo "🛠️  Setting up development environment..."
	@echo "✅ Development environment ready"

# Show version information
.PHONY: version
version:
	@echo "Mobile Backend CLI $(VERSION)"
	@echo "Go version: $(shell go version)"
	@echo "Build time: $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')"
