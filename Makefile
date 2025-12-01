.PHONY: build clean test test-coverage test-verbose install run-example help

# Build the binary
build:
	go build -o etu .

# Install to $GOPATH/bin
install:
	go install

# Run tests
test:
	@echo "Running tests..."
	@go test ./pkg/... -race

# Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	@go test ./pkg/... -v -race

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./pkg/... -cover -coverprofile=coverage.out
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
	GOOS=linux GOARCH=amd64 go build -o dist/etu-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o dist/etu-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o dist/etu-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o dist/etu-windows-amd64.exe .

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the etu binary"
	@echo "  install        - Install to \$$GOPATH/bin"
	@echo "  test           - Run unit tests"
	@echo "  test-verbose   - Run tests with verbose output"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Remove build artifacts"
	@echo "  run-example    - Run example parse command"
	@echo "  build-all      - Build for multiple platforms"
	@echo "  help           - Show this help message"
