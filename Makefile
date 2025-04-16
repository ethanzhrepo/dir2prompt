.PHONY: build install clean test test-cover

BINARY_NAME=dir-to-prompt
GOFLAGS=-ldflags="-s -w"

build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(GOFLAGS) -o bin/$(BINARY_NAME)

install: build
	@echo "Installing $(BINARY_NAME)..."
	@go install $(GOFLAGS)

clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out

test:
	@echo "Running tests..."
	@go test -v ./...

test-cover:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

run: build
	@echo "Running $(BINARY_NAME)..."
	@./bin/$(BINARY_NAME) $(ARGS)

all: clean build test 