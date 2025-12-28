.PHONY: build clean test test-integration test-all test-coverage test-verbose install run-example help

# Version info (injected at build time via ldflags)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X github.com/kazuma-desu/etu/cmd.Version=$(VERSION) \
                     -X github.com/kazuma-desu/etu/cmd.Commit=$(COMMIT) \
                     -X github.com/kazuma-desu/etu/cmd.BuildDate=$(BUILD_DATE)"

# Build the binary
build:
	go build $(LDFLAGS) -o etu .

# Install to $GOPATH/bin
install:
	go install $(LDFLAGS)

# Run unit tests (no containers needed)
test:
	@echo "Running unit tests..."
	@go test ./pkg/... ./cmd/... -race

# Run integration tests (requires Podman/Docker)
test-integration:
	@echo "Running integration tests (requires Podman/Docker)..."
	@go test ./cmd/... -race -tags=integration

# Run all tests
test-all:
	@echo "Running all tests..."
	@go test ./... -race -tags=integration

# Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	@go test ./... -v -race -tags=integration

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -cover -coverprofile=coverage.out -tags=integration
	@go tool cover -func=coverage.out
	@echo ""
	@echo "To view HTML coverage report, run: go tool cover -html=coverage.out"

# Clean build artifacts
clean:
	rm -f etu
	rm -f coverage.out coverage.html

# Run example with parse command
run-example:
	./etu parse -f examples/sample.txt

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/etu-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/etu-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/etu-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/etu-windows-amd64.exe .

# Show help
help:
	@echo "Available targets:"
	@echo "  build            - Build the etu binary"
	@echo "  install          - Install to \$$GOPATH/bin"
	@echo "  test             - Run unit tests (no containers needed)"
	@echo "  test-integration - Run integration tests (requires Podman/Docker)"
	@echo "  test-all         - Run all tests (unit + integration)"
	@echo "  test-verbose     - Run all tests with verbose output"
	@echo "  test-coverage    - Run all tests with coverage report"
	@echo "  clean            - Remove build artifacts"
	@echo "  run-example      - Run example parse command"
	@echo "  build-all        - Build for multiple platforms"
	@echo "  help             - Show this help message"
