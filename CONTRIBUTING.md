# Contributing to etu

Thank you for your interest in contributing to etu! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project follows a Code of Conduct. By participating, you are expected to uphold this code:

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on constructive criticism
- Accept responsibility for mistakes

## How to Contribute

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When creating a bug report, include:

- **Clear title and description**
- **Steps to reproduce** the issue
- **Expected behavior** vs **actual behavior**
- **Environment details** (OS, Go version, etu version)
- **Logs and error messages** if applicable

```markdown
**Bug Description**
A clear description of the bug.

**To Reproduce**
1. Run command '...'
2. See error

**Expected Behavior**
What should happen

**Environment**
- OS: Ubuntu 22.04
- Go: 1.21
- etu: v1.0.0
```

### Suggesting Features

Feature suggestions are welcome! Please provide:

- **Clear use case** - why is this needed?
- **Proposed solution** - how should it work?
- **Alternatives considered** - other approaches?
- **Additional context** - examples, mockups, etc.

### Pull Requests

1. **Fork the repository**

```bash
git clone https://github.com/kazuma-desu/etu.git
cd etu
git checkout -b feature/my-feature
```

2. **Make your changes**

- Write clear, readable code
- Follow existing code style
- Add tests for new functionality
- Update documentation as needed

3. **Test your changes**

```bash
# Run tests
go test ./...

# Build
go build -o etu .

# Test manually
./etu --help
```

4. **Commit with clear messages**

```bash
git commit -m "Add feature: brief description

Longer explanation of what changed and why.

Fixes #123"
```

Use conventional commit format:
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `refactor:` - Code refactoring
- `test:` - Adding tests
- `chore:` - Maintenance tasks

5. **Push and create pull request**

```bash
git push origin feature/my-feature
```

Then create a PR on GitHub with:
- Clear title describing the change
- Description of what changed and why
- Reference to related issues
- Screenshots/examples if applicable

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git
- Docker (optional, for testing with etcd)

### Local Development

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/etu.git
cd etu

# Install dependencies
go mod download

# Build
go build -o etu .

# Run tests
go test ./...

# Run locally
./etu --help
```

### Running Tests

etu uses a comprehensive testing strategy including unit tests and integration tests with real etcd containers using testcontainers.

#### Quick Test Commands

```bash
# Run all tests (unit + integration)
go test ./...

# Run only unit tests (fast, no Docker required)
go test -short ./...

# Run tests with coverage
go test -cover ./...

# Run tests with detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run tests for specific package
go test ./pkg/parsers

# Run with verbose output
go test -v ./...
```

#### Integration Tests

Integration tests use **testcontainers** to spin up real etcd instances. These tests provide confidence that the code works with actual etcd servers.

**Requirements:**
- Docker must be running
- Docker daemon accessible (usually automatic on Linux/macOS)

**Running Integration Tests:**

```bash
# Run all integration tests (requires Docker)
go test ./...

# Run specific integration test suites
go test ./pkg/client/... -run TestClient_Integration
go test ./cmd/... -run TestApplyCommand_Integration

# Skip integration tests (for CI or quick checks)
go test -short ./...
```

**What Integration Tests Cover:**
- Real etcd container lifecycle (start, use, cleanup)
- Client operations (Put, Get, PutAll, Status)
- End-to-end CLI workflows (parse → validate → apply)
- Error handling with actual network failures
- Different data types and edge cases

#### Test Organization

```
pkg/client/
├── etcd.go                    # Main client code
├── etcd_integration_test.go   # Integration tests (uses Docker)

pkg/output/
├── output.go                  # Output formatting code
├── output_test.go             # Unit tests (no Docker needed)

cmd/
├── apply.go                   # Apply command
├── apply_integration_test.go  # End-to-end tests (uses Docker)
```

#### Writing Tests

**Unit Tests:**
```go
func TestFormatValue(t *testing.T) {
    result := formatValue("test")
    assert.Equal(t, "test", result)
}
```

**Integration Tests:**
```go
func TestClient_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Setup etcd container
    endpoint, cleanup := setupEtcdContainer(t)
    defer cleanup()
    
    // Test with real etcd
    client, err := NewClient(&Config{
        Endpoints: []string{endpoint},
    })
    require.NoError(t, err)
    defer client.Close()
    
    // Perform operations...
}
```

**Key Guidelines:**
- Always skip integration tests in short mode: `if testing.Short() { t.Skip(...) }`
- Use `require` for setup assertions (fail fast)
- Use `assert` for test assertions (continue testing)
- Clean up resources with `defer`
- Use table-driven tests for multiple scenarios

#### Coverage Goals

Current coverage by package:
- `pkg/models`: 100% ✅
- `pkg/output`: 100% ✅
- `pkg/validator`: 97.4% ✅
- `pkg/parsers`: 83.5%
- `pkg/client`: 65.5% (with integration tests)
- `pkg/config`: 47.7%
- `cmd`: 22.0% (with integration tests)

**For new code:**
- Aim for at least 80% coverage
- Critical paths should have 100% coverage
- Include both happy path and error cases
- Add integration tests for etcd interactions

#### Continuous Integration

Tests run automatically on pull requests:
- Unit tests run on every commit (fast)
- Integration tests run on PR creation/update
- Coverage reports are generated and tracked

```yaml
# Example CI workflow
- name: Run unit tests
  run: go test -short ./...

- name: Run integration tests
  run: go test ./...

- name: Generate coverage
  run: go test -coverprofile=coverage.out ./...
```

### Code Style

- Follow standard Go conventions ([Effective Go](https://golang.org/doc/effective_go.html))
- Use `gofmt` to format code
- Run `go vet` to check for issues
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions small and focused

```bash
# Format code
gofmt -w .

# Check for issues
go vet ./...

# Run linter (if installed)
golangci-lint run
```

## Project Structure

```
etu/
├── cmd/           # CLI command implementations
├── pkg/           # Public packages (library API)
│   ├── client/    # etcd client wrapper
│   ├── config/    # Configuration management
│   ├── models/    # Data models
│   ├── parsers/   # File format parsers
│   ├── validator/ # Validation logic
│   └── output/    # Output formatting
├── examples/      # Example files
├── internal/      # Private packages (future use)
└── main.go        # Entry point
```

### Adding New Features

#### Adding a New Parser

1. Create parser file in `pkg/parsers/`
2. Implement the `Parser` interface
3. Register in `NewRegistry()`
4. Add format detection logic
5. Write tests
6. Update documentation

Example:

```go
// pkg/parsers/yaml.go
type YAMLParser struct{}

func (p *YAMLParser) Parse(path string) ([]*models.ConfigPair, error) {
    // Implementation
}

func (p *YAMLParser) FormatName() string {
    return "yaml"
}
```

#### Adding a New Command

1. Create command file in `cmd/`
2. Define cobra command
3. Register in `init()`
4. Implement command logic
5. Add tests
6. Update help text and README

## Documentation

- Update README.md for user-facing changes
- Add godoc comments for exported functions
- Update examples if needed
- Add comments for complex logic

### Writing Good Commit Messages

```
feat: add YAML parser support

- Implement YAMLParser struct
- Add parser registration
- Include tests for YAML parsing
- Update README with YAML examples

Closes #42
```

## Release Process

Releases are managed by maintainers:

1. Update version in relevant files
2. Update CHANGELOG.md
3. Create git tag (e.g., `v1.0.0`)
4. Push tag to trigger release workflow
5. Create GitHub release with notes

## Questions?

- Open an issue for questions
- Join discussions on GitHub Discussions
- Check existing documentation first

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Thank You!

Your contributions make etu better for everyone. We appreciate your time and effort!
