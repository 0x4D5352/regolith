.PHONY: build test clean generate install release all

# VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
# LDFLAGS := -ldflags "-X main.version=$(VERSION)"
PIGEON := $(shell go env GOPATH)/bin/pigeon

# Default target
all: generate build test

# Build for current platform
build:
	# go build $(LDFLAGS) ./cmd/regolith
	go build ./cmd/regolith

# Install to GOPATH/bin
install:
	# go install $(LDFLAGS) ./cmd/regolith
	go install ./cmd/regolith

# Run tests
test:
	go test -v ./...

# Run tests with coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Generate all parsers from grammars
.PHONY: generate
generate: generate-javascript generate-posix-ere generate-posix-bre generate-gnugrep-bre generate-gnugrep-ere generate-java generate-dotnet generate-pcre

# Generate JavaScript parser
.PHONY: generate-javascript
generate-javascript: $(PIGEON)
	$(PIGEON) -o internal/flavor/javascript/parser.go internal/flavor/javascript/grammar.peg

# Generate POSIX ERE parser
.PHONY: generate-posix-ere
generate-posix-ere: $(PIGEON)
	$(PIGEON) -o internal/flavor/posix_ere/parser.go internal/flavor/posix_ere/grammar.peg

# Generate POSIX BRE parser
.PHONY: generate-posix-bre
generate-posix-bre: $(PIGEON)
	$(PIGEON) -o internal/flavor/posix_bre/parser.go internal/flavor/posix_bre/grammar.peg

# Generate GNU grep BRE parser
.PHONY: generate-gnugrep-bre
generate-gnugrep-bre: $(PIGEON)
	$(PIGEON) -o internal/flavor/gnugrep_bre/parser.go internal/flavor/gnugrep_bre/grammar.peg

# Generate GNU grep ERE parser
.PHONY: generate-gnugrep-ere
generate-gnugrep-ere: $(PIGEON)
	$(PIGEON) -o internal/flavor/gnugrep_ere/parser.go internal/flavor/gnugrep_ere/grammar.peg

# Generate Java parser
.PHONY: generate-java
generate-java: $(PIGEON)
	$(PIGEON) -o internal/flavor/java/parser.go internal/flavor/java/grammar.peg

# Generate .NET parser
.PHONY: generate-dotnet
generate-dotnet: $(PIGEON)
	$(PIGEON) -o internal/flavor/dotnet/parser.go internal/flavor/dotnet/grammar.peg

# Generate PCRE parser
.PHONY: generate-pcre
generate-pcre: $(PIGEON)
	$(PIGEON) -o internal/flavor/pcre/parser.go internal/flavor/pcre/grammar.peg

# Install pigeon if needed
$(PIGEON):
	go install github.com/mna/pigeon@latest

# Clean build artifacts
clean:
	rm -f regolith
	rm -f coverage.out coverage.html
	rm -rf dist/

# Cross-compile for all platforms
release: clean
	mkdir -p dist
	# Linux
	# GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/regolith-linux-amd64 ./cmd/regolith
	GOOS=linux GOARCH=amd64 go build -o dist/regolith-linux-amd64 ./cmd/regolith
	# GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/regolith-linux-arm64 ./cmd/regolith
	GOOS=linux GOARCH=arm64 go build -o dist/regolith-linux-arm64 ./cmd/regolith
	# macOS
	# GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/regolith-darwin-amd64 ./cmd/regolith
	GOOS=darwin GOARCH=amd64 go build -o dist/regolith-darwin-amd64 ./cmd/regolith
	# GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/regolith-darwin-arm64 ./cmd/regolith
	GOOS=darwin GOARCH=arm64 go build -o dist/regolith-darwin-arm64 ./cmd/regolith
	# Windows
	# GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/regolith-windows-amd64.exe ./cmd/regolith
	GOOS=windows GOARCH=amd64 go build -o dist/regolith-windows-amd64.exe ./cmd/regolith
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
	@echo "  build               - Build for current platform"
	@echo "  install             - Install to GOPATH/bin"
	@echo "  test                - Run tests"
	@echo "  coverage            - Run tests with coverage report"
	@echo "  generate            - Regenerate all parsers from grammars"
	@echo "  generate-javascript - Regenerate JavaScript parser"
	@echo "  generate-posix-ere  - Regenerate POSIX ERE parser"
	@echo "  generate-posix-bre  - Regenerate POSIX BRE parser"
	@echo "  generate-java       - Regenerate Java parser"
	@echo "  generate-dotnet     - Regenerate .NET parser"
	@echo "  generate-pcre       - Regenerate PCRE parser"
	@echo "  clean               - Remove build artifacts"
	@echo "  release             - Cross-compile for all platforms"
	@echo "  golden              - Update golden test files"
	@echo "  lint                - Run linter"
	@echo "  fmt                 - Format code"
