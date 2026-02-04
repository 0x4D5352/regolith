.PHONY: build test clean generate install release all

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

# Default target
all: generate build test

# Build for current platform
build:
	go build $(LDFLAGS) ./cmd/regolith

# Install to GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/regolith

# Run tests
test:
	go test -v ./...

# Run tests with coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Generate parser from grammar
generate:
	@which pigeon > /dev/null || (echo "Installing pigeon..." && go install github.com/mna/pigeon@latest)
	pigeon -o internal/parser/parser.go internal/parser/grammar.peg

# Clean build artifacts
clean:
	rm -f regolith
	rm -f coverage.out coverage.html
	rm -rf dist/

# Cross-compile for all platforms
release: clean
	mkdir -p dist
	# Linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/regolith-linux-amd64 ./cmd/regolith
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/regolith-linux-arm64 ./cmd/regolith
	# macOS
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/regolith-darwin-amd64 ./cmd/regolith
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/regolith-darwin-arm64 ./cmd/regolith
	# Windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/regolith-windows-amd64.exe ./cmd/regolith
	# Create checksums
	cd dist && shasum -a 256 * > checksums.txt

# Update golden test files
golden:
	GOLDEN_UPDATE=1 go test ./internal/renderer/...

# Lint code
lint:
	@which golangci-lint > /dev/null || (echo "Install golangci-lint: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Show help
help:
	@echo "Available targets:"
	@echo "  build     - Build for current platform"
	@echo "  install   - Install to GOPATH/bin"
	@echo "  test      - Run tests"
	@echo "  coverage  - Run tests with coverage report"
	@echo "  generate  - Regenerate parser from grammar"
	@echo "  clean     - Remove build artifacts"
	@echo "  release   - Cross-compile for all platforms"
	@echo "  golden    - Update golden test files"
	@echo "  lint      - Run linter"
	@echo "  fmt       - Format code"
