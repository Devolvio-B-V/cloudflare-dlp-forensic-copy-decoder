.PHONY: build test clean install cross-build help

# Binary name
BINARY_NAME=cf-dlp-decode
BUILD_DIR=bin
DIST_DIR=dist

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet
GOFMT=$(GOCMD) fmt

# Version
VERSION?=2.0.0
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Platforms for cross-compilation
PLATFORMS=linux darwin windows
ARCHS=amd64 arm64

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary for current platform
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/decoder
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "Tests completed"

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...

fmt: ## Format code with go fmt
	@echo "Formatting code..."
	$(GOFMT) ./...

tidy: ## Run go mod tidy
	@echo "Running go mod tidy..."
	$(GOMOD) tidy

clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.txt coverage.out coverage.html
	@echo "Clean completed"

install: build ## Install the binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME) to $(GOPATH)/bin..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "Installation completed"

cross-build: ## Build for all platforms (linux, darwin, windows) and architectures (amd64, arm64)
	@echo "Cross-compiling for multiple platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		for arch in $(ARCHS); do \
			output_name=$(DIST_DIR)/$(BINARY_NAME)-$$platform-$$arch; \
			if [ $$platform = "windows" ]; then \
				output_name=$$output_name.exe; \
			fi; \
			echo "Building for $$platform/$$arch..."; \
			GOOS=$$platform GOARCH=$$arch $(GOBUILD) $(LDFLAGS) -o $$output_name ./cmd/decoder; \
		done \
	done
	@echo "Cross-compilation completed. Binaries in $(DIST_DIR)/"

run: build ## Build and run the application
	@$(BUILD_DIR)/$(BINARY_NAME) --help

all: clean fmt vet test build ## Run all checks and build

ci: vet test build ## Run CI checks (vet, test, build)
