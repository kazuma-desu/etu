# Contributing to etu

Thank you for considering contributing to etu! This document provides guidelines and instructions for contributing.

## Code of Conduct

Be respectful, inclusive, and professional in all interactions.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork**:
   ```bash
   git clone https://github.com/YOUR-USERNAME/etu.git
   cd etu
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/kazuma-desu/etu.git
   ```
4. **Create a feature branch**:
   ```bash
   git checkout -b feature/my-new-feature
   ```

## Development Setup

### Prerequisites

- Go 1.24 or later
- golangci-lint (for linting)
- Docker (optional, for integration tests)

### Install Dependencies

```bash
go mod download
```

### Run Tests

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage
go tool cover -html=coverage.out
```

### Run Linter

```bash
golangci-lint run
```

### Build

```bash
go build -o etu .
```

## Making Changes

### Code Style

- Follow standard Go conventions and idioms
- Run `gofmt` and `goimports` on your code
- Keep functions small and focused
- Add comments for exported functions and types
- Write clear commit messages

### Commit Messages

Follow the conventional commits format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:
```
feat(parse): add tree view visualization

Add --tree flag to parse command that displays etcd paths
in a hierarchical tree structure using lipgloss tree package.

Closes #123
```

```
fix(validator): handle nil values correctly

Previously nil values would cause a panic. Now they are
properly validated and return appropriate warnings.
```

### Testing

- Write tests for new features
- Ensure all tests pass before submitting PR
- Aim for good test coverage
- Include both unit tests and integration tests where appropriate

### Documentation

- Update README.md for user-facing changes
- Add comments for exported functions and types
- Update examples if needed

## Submitting Changes

1. **Sync with upstream**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Push to your fork**:
   ```bash
   git push origin feature/my-new-feature
   ```

3. **Create a Pull Request**:
   - Go to the repository on GitHub
   - Click "New Pull Request"
   - Select your branch
   - Fill out the PR template with:
     - Description of changes
     - Related issues
     - Testing performed
     - Screenshots (if UI changes)

4. **Address review feedback**:
   - Make requested changes
   - Push new commits to the same branch
   - Respond to comments

## Pull Request Guidelines

- Keep PRs focused on a single feature or fix
- Include tests for new functionality
- Update documentation as needed
- Ensure CI passes (linting, tests, builds)
- Rebase on main before submitting
- Squash commits if requested

## Project Structure

```
etu/
├── cmd/              # CLI commands (cobra commands)
├── pkg/              # Public library packages
│   ├── client/       # etcd client wrapper
│   ├── config/       # Configuration management
│   ├── models/       # Data models
│   ├── output/       # Terminal output styling
│   ├── parsers/      # File format parsers
│   └── validator/    # Configuration validation
├── examples/         # Example configuration files
├── .github/          # GitHub Actions workflows
└── main.go          # Application entry point
```

## Adding New Features

### Adding a New Command

1. Create a new file in `cmd/` (e.g., `cmd/mynewcommand.go`)
2. Implement the cobra command
3. Register it in `cmd/root.go`
4. Add tests in `cmd/mynewcommand_test.go`
5. Update README.md

### Adding a New Parser

1. Implement the `Parser` interface in `pkg/parsers/`
2. Register it in the parser registry
3. Add comprehensive tests
4. Update documentation

### Adding a New Output Format

1. Add functions to `pkg/output/output.go`
2. Use lipgloss for styling consistency
3. Test with various inputs
4. Document the new format

## Integration Tests

Integration tests require a running etcd instance:

```bash
# Start etcd with Docker
docker run -d --name etcd-test -p 2379:2379 \
  quay.io/coreos/etcd:v3.5.12 \
  /usr/local/bin/etcd \
  --listen-client-urls http://0.0.0.0:2379 \
  --advertise-client-urls http://0.0.0.0:2379

# Run integration tests
go test -v ./...

# Clean up
docker rm -f etcd-test
```

## Releasing

Only maintainers can create releases. To create a new release:

### Step 1: Prepare the Release

1. Ensure all changes are merged to `main`
2. Update version in documentation if needed
3. Verify all tests pass: `go test -v ./...`
4. Verify linter passes: `golangci-lint run`

### Step 2: Create and Push Tag

```bash
# Create an annotated tag
git checkout main
git pull origin main
git tag -a v0.1.0 -m "Release v0.1.0"

# Push the tag
git push origin v0.1.0
```

### Step 3: Automated Release Process

GitHub Actions will automatically:
1. Build binaries for all platforms (Linux, macOS, Windows)
2. Create archives (.tar.gz for Unix, .zip for Windows)
3. Generate changelog from commit history
4. Create GitHub release with all binaries
5. Publish release notes with installation instructions

### Step 4: Verify Release

1. Check [GitHub Releases](https://github.com/kazuma-desu/etu/releases)
2. Verify all binaries are present:
   - `etu_0.1.0_linux_amd64.tar.gz`
   - `etu_0.1.0_linux_arm64.tar.gz`
   - `etu_0.1.0_darwin_amd64.tar.gz`
   - `etu_0.1.0_darwin_arm64.tar.gz`
   - `etu_0.1.0_windows_amd64.zip`
3. Test download and installation instructions
4. Announce the release

### Release Checklist

- [ ] All PRs merged
- [ ] Tests passing
- [ ] Linter passing
- [ ] Documentation updated
- [ ] Tag created and pushed
- [ ] Release verified
- [ ] Release announced

## CI/CD Pipeline

### Continuous Integration (CI)

Runs on every push and pull request to `main` or `develop`:

**Lint Job:**
- Runs golangci-lint with v2 configuration
- Checks code quality and style
- Reports issues in PR

**Test Job:**
- Sets up etcd service container
- Runs all tests with race detection
- Generates code coverage report
- Uploads coverage to SonarCloud and Codecov
- Performs SonarCloud quality analysis

**Build Job:**
- Builds for all target platforms
- Verifies successful compilation
- Uploads build artifacts (7-day retention)

### Release Pipeline

Triggered by version tags (`v*`):

**Build Job:**
- Compiles binaries for all platforms in parallel
- Embeds version info in binaries
- Creates platform-specific archives
- Uploads artifacts for release job

**Release Job:**
- Downloads all build artifacts
- Generates changelog from git history
- Creates GitHub release
- Uploads all binaries with installation instructions

### Code Quality Tools

**golangci-lint:**
- Version: v2 configuration
- Enabled linters: errcheck, govet, staticcheck, misspell, gocritic, revive, and more
- Formatters: goimports

**SonarCloud:**
- Code coverage tracking
- Quality gate enforcement
- Security vulnerability detection
- Code smell identification
- Pull request decoration

## Getting Help

- Open an issue for bugs or feature requests
- Ask questions in discussions
- Join community chat (if available)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
