.PHONY: all build test test-unit test-integration clean lint fmt

# Variables
BINARY_NAME=ghsecrets
GO=go
GOTEST=$(GO) test
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOGET=$(GO) get
GOFMT=gofmt

# Build the binary
all: build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v main.go

# Run all tests
test: test-unit

# Run unit tests only
test-unit:
	$(GOTEST) -v -short ./...

# Run integration tests (requires credentials)
test-integration:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Format code
fmt:
	$(GOFMT) -s -w .

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Install dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

# Run the application
run: build
	./$(BINARY_NAME)

# Install the binary to $GOPATH/bin
install: build
	$(GO) install