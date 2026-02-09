# AGENTS.md - etu CLI Development Guide

This file provides guidance for AI agents working on the etu CLI codebase.

## Quick Start

```bash
# Build
make build

# Run unit tests (fast, no containers)
make test

# Run single test
go test -v -run TestFunctionName ./pkg/client/

# Run integration tests (requires Podman/Docker)
export DOCKER_HOST=unix:///run/user/1000/podman/podman.sock
make test-integration

# Run all tests with race detection
go test ./... -race
```

## Project Structure

```
etu/
├── cmd/              # Cobra CLI commands (one file per command)
│   ├── root.go       # Root command, global flags
│   ├── get.go        # Individual command implementations
│   └── *_test.go     # Unit and integration tests
├── pkg/
│   ├── client/       # etcd client wrapper and interface
│   ├── config/       # Configuration management
│   ├── models/       # Shared data types
│   ├── output/       # Output formatting (table, json, yaml, etc.)
│   ├── parsers/      # Input file parsers (yaml, json, etcdctl)
│   └── validator/    # Configuration validation
├── examples/         # Example configuration files
└── Makefile          # Build automation
```

## Code Style Guidelines

### Naming Conventions
- **Commands**: `cmd/command_name.go` with `commandNameCmd` variable
- **Options**: `commandOpts` struct for command-specific flags
- **Functions**: `runCommandName()` for command entry points
- **Tests**: `TestCommandName_Description` pattern

### Command Structure Pattern
```go
var (
    commandOpts struct {
        flagName string
    }
    
    commandCmd = &cobra.Command{
        Use:   "command [args]",
        Short: "Brief description",
        Long:  `Long description with examples`,
        RunE:  runCommand,
    }
)

func init() {
    rootCmd.AddCommand(commandCmd)
    commandCmd.Flags().StringVar(&commandOpts.flagName, "flag", "", "description")
}

func runCommand(cmd *cobra.Command, args []string) error {
    // Implementation
}
```

### Error Handling
- Use `fmt.Errorf("context: %w", err)` for error wrapping
- Use `output.Error()` helper for user-facing errors
- Return errors from `RunE` functions (don't call os.Exit)
- Include actionable hints in error messages

### Output Formatting
- Support `-o, --output` flag with formats: simple, json, yaml, table, tree
- Use `pkg/output` package helpers for consistent formatting
- Check `isQuietOutput()` for non-interactive mode

### Testing
- Unit tests: `go test -short ./...`
- Integration tests: Use `-tags=integration` with testcontainers
- Mock client: `pkg/client/mock.go` for unit tests
- Reset flags: Call `resetCommandFlags()` in tests

### Imports Ordering
```go
import (
    // Standard library
    "fmt"
    "os"
    
    // Third-party
    "github.com/spf13/cobra"
    
    // Internal
    "github.com/kazuma-desu/etu/pkg/client"
    "github.com/kazuma-desu/etu/pkg/models"
)
```

## Common Tasks

### Adding a New Command
1. Create `cmd/newcommand.go` following the pattern
2. Add command to `rootCmd` in `init()`
3. Create `cmd/newcommand_test.go` with tests
4. Update `Makefile` if needed

### Running etcd for Testing
```bash
# Start local etcd (no auth)
make etcd-dev

# Start with auth
make etcd-dev-auth

# Connect
./etu login --context-name dev --endpoints http://localhost:2379 --no-auth
./etu config use-context dev
```

### Debugging Integration Tests
```bash
# Run specific integration test
export DOCKER_HOST=unix:///run/user/1000/podman/podman.sock
go test ./cmd/... -tags=integration -v -run TestName
```

## Key Conventions

1. **Key Format**: All etcd keys must start with `/`
2. **Type Safety**: All values are stored and retrieved as strings (string-canonical model)
3. **Config Security**: Config file uses 0600 permissions; warn if more open
4. **Output Consistency**: All commands support `-o` flag with standard formats
5. **Flag Naming**: Use `-f` for file paths, `-o` for output format, `--dry-run` for dry runs

## Dependencies

Key external dependencies:
- `github.com/spf13/cobra` - CLI framework
- `go.etcd.io/etcd/client/v3` - etcd client
- `github.com/charmbracelet/*` - Terminal UI components
- `github.com/stretchr/testify` - Testing utilities
- `github.com/testcontainers/testcontainers-go` - Integration testing

## ⚠️ DO NOT COMMIT

**Never commit the following to git:**

1. **`.sisyphus/` folder** - Contains agentic planning files, not part of the codebase
2. **Agent-generated documentation** - Any docs created by AI agents (untracked agentic docs should not be commited unless explicitly specified)
3. **Temporary files** - Files with names like `EOF`, `out.yaml`, etc.

**Always check before committing:**
```bash
git status
# Ensure .sisyphus/ is not in the staged changes
# If it is, unstage with: git reset HEAD .sisyphus/
```
