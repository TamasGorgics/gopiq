# Makefile for gopiq

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOINSTALL=$(GOCMD) install
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Lint parameters
LINT_CMD=golangci-lint

# Build variables
BINARY_NAME=gopiq
BINARY_UNIX=$(BINARY_NAME)_unix

.PHONY: all build test cover bench vet lint clean help docs-serve docs-build

all: build

# Build commands
build:
	@echo "Building gopiq..."
	@$(GOBUILD) -o $(BINARY_NAME) -v

# Test commands
test:
	@echo "Running tests..."
	@$(GOTEST) -v -race ./...

# Test with coverage commands
cover:
	@echo "Running tests with coverage..."
	@$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic -coverpkg=github.com/TamasGorgics/gopiq,github.com/TamasGorgics/gopiq/internal/engine ./...
	@echo "To view coverage report, run: make cover-html"

cover-html: cover
	@echo "Opening coverage report in browser..."
	@go tool cover -html=coverage.out

# Benchmark commands
bench:
	@echo "Running benchmarks..."
	@$(GOTEST) -run '^$$' -bench . -benchmem ./...

vet:
	@echo "Running go vet..."
	@$(GOCMD) vet ./...

# Lint commands
lint:
	@echo "Running linter..."
	@$(LINT_CMD) run

# Docs commands
docs-serve:
	@echo "Serving documentation at http://127.0.0.1:8000"
	@python3 -m venv venv && source venv/bin/activate && pip install -r docs/requirements.txt && mkdocs serve -f docs/mkdocs.yml

docs-build:
	@echo "Building documentation..."
	@python3 -m venv venv && source venv/bin/activate && pip install -r docs/requirements.txt && mkdocs build -f docs/mkdocs.yml

# Clean command
clean:
	@echo "Cleaning up..."
	@$(GOCLEAN)
	@rm -f $(BINARY_NAME) $(BINARY_UNIX) coverage.out

# Help command
help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  build         Build the application binary"
	@echo "  test          Run tests with race condition detection"
	@echo "  cover         Run tests and generate a coverage profile"
	@echo "  cover-html    Open the HTML coverage report in a browser"
	@echo "  bench         Run all benchmarks"
	@echo "  vet           Run go vet"
	@echo "  lint          Run the golangci-lint linter"
	@echo "  docs-serve    Serve the documentation site locally"
	@echo "  docs-build    Build the documentation site"
	@echo "  clean         Clean up build artifacts and coverage files"
	@echo "  help          Show this help message"

.DEFAULT_GOAL := help 