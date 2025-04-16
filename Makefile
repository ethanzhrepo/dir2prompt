.PHONY: build install clean test test-cover cross-build build-windows build-linux build-darwin build-all

BINARY_NAME=dir2prompt
GOFLAGS=-ldflags="-s -w"
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# 平台和架构定义
PLATFORMS=windows linux darwin
ARCHITECTURES=amd64 arm64

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

# Windows 构建
build-windows-amd64:
	@echo "Building for Windows (amd64)..."
	@GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -o bin/$(BINARY_NAME)_windows_amd64.exe

build-windows-arm64:
	@echo "Building for Windows (arm64)..."
	@GOOS=windows GOARCH=arm64 go build $(GOFLAGS) -o bin/$(BINARY_NAME)_windows_arm64.exe

build-windows: build-windows-amd64 build-windows-arm64
	@echo "Windows builds completed"

# Linux 构建
build-linux-amd64:
	@echo "Building for Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o bin/$(BINARY_NAME)_linux_amd64

build-linux-arm64:
	@echo "Building for Linux (arm64)..."
	@GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -o bin/$(BINARY_NAME)_linux_arm64

build-linux: build-linux-amd64 build-linux-arm64
	@echo "Linux builds completed"

# macOS 构建
build-darwin-amd64:
	@echo "Building for macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) -o bin/$(BINARY_NAME)_darwin_amd64

build-darwin-arm64:
	@echo "Building for macOS (arm64)..."
	@GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) -o bin/$(BINARY_NAME)_darwin_arm64

build-darwin: build-darwin-amd64 build-darwin-arm64
	@echo "macOS builds completed"

# 构建所有平台
build-all: build-windows build-linux build-darwin
	@echo "All platform builds completed"

# 创建发布压缩包
release: build-all
	@echo "Creating release archives..."
	@mkdir -p release
	@for platform in windows linux darwin; do \
		for arch in amd64 arm64; do \
			if [ "$$platform" = "windows" ]; then \
				zip -j release/$(BINARY_NAME)_$(VERSION)_$${platform}_$${arch}.zip bin/$(BINARY_NAME)_$${platform}_$${arch}.exe; \
			else \
				tar -czf release/$(BINARY_NAME)_$(VERSION)_$${platform}_$${arch}.tar.gz -C bin $(BINARY_NAME)_$${platform}_$${arch}; \
			fi; \
		done; \
	done
	@echo "Release archives created in release/ directory"

all: clean build test release