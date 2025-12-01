# etu

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev/)

**Etcd Terminal Utility**

A modern, production-ready CLI tool and Go library for managing etcd configurations. Built with Go and designed with kubectl like UX in mind.

## Features

- **Multi-Context Management** - Manage multiple etcd clusters (dev, staging, prod) with easy context switching
- **Beautiful Output** - Modern, colorful terminal output using [Charm](https://charm.sh/) libraries
- **Comprehensive Validation** - Validates keys, values, structured data, URLs, and more before applying
- **Multiple Format Support** - Extensible parser system (currently supports etcdctl output format)
- **Native etcd Integration** - Uses etcd/client/v3 for direct cluster communication
- **Dry Run Mode** - Preview changes before applying them to prevent cluster outages
- **JSON Output** - Machine-readable output for scripting and automation
- **Flexible Configuration** - Config file, environment variables, or CLI flags
- **Configurable Defaults** - Set default log levels, formats, and validation modes
- **Secure** - Optional authentication with security warnings for stored credentials

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/kazuma-desu/etu.git
cd etu

# Build
go build -o etu .

# Install to $GOPATH/bin
go install
```

### Using go install

```bash
go install github.com/kazuma-desu/etu@latest
```

### Download Binary

Download pre-built binaries from the [releases page](https://github.com/kazuma-desu/etu/releases).

## Quick Start

### 1. Login to etcd cluster

```bash
# Login without authentication
etu login dev --endpoints http://localhost:2379 --no-auth

# Login with authentication
etu login prod --endpoints http://prod:2379 --username admin --password secret

# View saved contexts
etu config get-contexts
```

### 2. Manage your configuration

```bash
# Parse and display configuration
etu parse -f examples/sample.txt

# Validate configuration
etu validate -f examples/sample.txt

# Apply to etcd (with validation)
etu apply -f examples/sample.txt

# Dry run (preview without applying)
etu apply -f examples/sample.txt --dry-run
```

## Commands

### Context Management

```bash
# Login and save connection details
etu login <context-name> --endpoints <url> [--username <user>] [--password <pass>]

# Switch between contexts
etu config use-context <context-name>

# List all contexts
etu config get-contexts

# Show current context
etu config current-context

# Delete a context
etu config delete-context <context-name>

# View configuration
etu config view
```

### Configuration Management

```bash
# Set default log level
etu config set log-level debug

# Set default file format
etu config set default-format etcdctl

# Enable strict validation by default
etu config set strict true
```

### File Operations

#### parse

Parse and display configuration from a file.

```bash
# Human-readable output
etu parse -f config.txt

# JSON output for scripting
etu parse -f config.txt --json

# Pipe to jq for filtering
etu parse -f config.txt --json | jq '.[] | select(.key | contains("database"))'
```

#### validate

Validate configuration without applying to etcd.

```bash
# Standard validation
etu validate -f config.txt

# Strict mode (warnings treated as errors)
etu validate -f config.txt --strict
```

**Validation Checks:**
- Key format (must start with `/`, valid characters, length/depth limits)
- Value validation (non-null, size limits)
- Structured data (JSON/YAML) validity
- URL validation for keys containing "url"
- Duplicate key detection

#### apply

Apply configuration to etcd.

```bash
# Apply with validation
etu apply -f config.txt

# Preview changes (dry run)
etu apply -f config.txt --dry-run

# Skip validation (not recommended)
etu apply -f config.txt --no-validate

# Strict validation
etu apply -f config.txt --strict

# Use specific context
etu apply -f config.txt --context prod
```

## Configuration

etu supports three configuration sources with the following priority:

**Priority: CLI Flags > Config File > Environment Variables**

### Config File

Configuration is stored in `~/.config/etu/config.yaml`:

```yaml
current-context: prod
log-level: warn
default-format: auto
strict: false
no-validate: false

contexts:
  dev:
    endpoints:
      - http://localhost:2379
  prod:
    endpoints:
      - http://prod:2379
    username: admin
    password: secret  # Stored in plain text - see security note below
```

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `ETCD_ENDPOINTS` | Comma-separated etcd endpoints | `http://localhost:2379,http://localhost:2380` |
| `ETCD_HOST` | Single endpoint (backwards compatible) | `http://localhost:2379` |
| `ETCD_USERNAME` | Username for authentication | `root` |
| `ETCD_PASSWORD` | Password for authentication | `password` |
| `ETCD_USERPASS` | Combined user:pass format | `root:password` |

### Global Flags

All commands support these flags:

- `--context <name>`: Use specific context (overrides current context)
- `--log-level <level>`: Set log level - debug, info, warn, error (default: warn)
- `-h, --help`: Show help

## Configuration File Format

### etcdctl Format

Plain text format where keys start with `/`:

```
/app/config/database/host
db.example.com

/app/config/database/port
5432

/app/config/features/max_retries
5

/app/i18n/welcome_message
en: Welcome
es: Bienvenido
fr: Bienvenue
```

**Features:**
- Automatic type inference (integers, floats, strings)
- Multi-line values supported
- Dictionary parsing for `key: value` patterns
- Quote stripping for string values

## Security

**Important:** Passwords are stored in plain text in `~/.config/etu/config.yaml` (similar to Docker's approach).

For better security in production/CI environments:
- Don't store passwords in config - provide via `--password` flag at runtime
- Use environment variables (`ETCD_PASSWORD`)
- Config file permissions are automatically set to `0600`
- Use etcd's TLS/mTLS authentication (future feature)

## Using as a Library

etu can be used as a Go library in your own projects:

```go
import (
    "context"
    "github.com/kazuma-desu/etu/pkg/client"
    "github.com/kazuma-desu/etu/pkg/parsers"
    "github.com/kazuma-desu/etu/pkg/validator"
)

func main() {
    // Create etcd client
    cfg := &client.Config{
        Endpoints: []string{"http://localhost:2379"},
        Username:  "root",
        Password:  "password",
    }
    etcdClient, _ := client.NewClient(cfg)
    defer etcdClient.Close()

    // Parse configuration file
    registry := parsers.NewRegistry()
    parser, _ := registry.GetParser("etcdctl")
    pairs, _ := parser.Parse("config.txt")

    // Validate
    v := validator.NewValidator(false)
    result := v.Validate(pairs)
    
    if result.Valid {
        // Apply to etcd
        ctx := context.Background()
        etcdClient.PutAll(ctx, pairs)
    }
}
```

## Project Structure

```
etu/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command and logging setup
│   ├── login.go           # Login command
│   ├── config.go          # Config management commands
│   ├── apply.go           # Apply command
│   ├── validate.go        # Validate command
│   └── parse.go           # Parse command
├── pkg/                    # Public packages (library API)
│   ├── client/            # etcd client wrapper
│   ├── config/            # Configuration management  
│   ├── models/            # Domain models and types
│   ├── parsers/           # Extensible parser system
│   │   ├── parser.go      # Parser interface and registry
│   │   └── etcdctl.go     # etcdctl format parser
│   ├── validator/         # Configuration validation
│   └── output/            # Styled output formatting
├── examples/              # Example configuration files
├── main.go               # Entry point
└── go.mod                # Go module definition
```

## Development

### Prerequisites

- Go 1.24 or later
- Access to an etcd cluster (for testing apply operations)

### Building

```bash
# Build
go build -o etu .

# Run tests
go test ./...

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o etu-linux-amd64 .
GOOS=darwin GOARCH=amd64 go build -o etu-darwin-amd64 .
GOOS=windows GOARCH=amd64 go build -o etu-windows-amd64.exe .
```

### Running Examples

```bash
# Start a local etcd instance (using Docker)
docker run -d --name etcd-test \
  -p 2379:2379 \
  -e ALLOW_NONE_AUTHENTICATION=yes \
  bitnami/etcd:latest

# Login
./etu login dev --endpoints http://localhost:2379 --no-auth

# Try the examples
./etu parse -f examples/sample.txt
./etu validate -f examples/sample.txt
./etu apply -f examples/sample.txt --dry-run

# Clean up
docker rm -f etcd-test
```

## Architecture

### Extensible Parser System

The parser system is designed for extensibility. To add a new format:

1. **Implement the Parser interface:**

```go
type Parser interface {
    Parse(path string) ([]*models.ConfigPair, error)
    FormatName() string
}
```

2. **Register your parser:**

```go
// In parsers/parser.go NewRegistry()
r.Register(models.FormatYourFormat, &YourParser{})
```

3. **Add format detection logic:**

```go
// In parsers/parser.go DetectFormat()
// Add logic to detect your format
```

Example implementations:
- `pkg/parsers/etcdctl.go` - etcdctl output format parser

## Roadmap

- Additional parsers: Helm values.yaml, TOML, HCL, JSON, YAML
- Get operations: Retrieve and display etcd keys
- Watch operations: Monitor etcd keys for changes
- Diff operations: Compare local config with etcd state
- Batch operations: Transaction support for atomic updates
- TLS support: Secure etcd connections
- Shell completion: Bash/Zsh/Fish completion scripts
- Backup/restore: etcd backup and restore operations

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Ways to Contribute

- Report bugs
- Suggest new features
- Improve documentation
- Submit pull requests
- Star the project

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by [kubectl](https://kubernetes.io/docs/reference/kubectl/) and [helm](https://helm.sh/) CLI design
- Built with [Charm](https://charm.sh/) tools (lipgloss, log)
- Uses etcd's official [Go client](https://github.com/etcd-io/etcd/tree/main/client/v3)

## Support

- [Documentation](https://github.com/kazuma-desu/etu/wiki)
- [Issue Tracker](https://github.com/kazuma-desu/etu/issues)
- [Discussions](https://github.com/kazuma-desu/etu/discussions)
