# etu

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev/)
[![CI](https://github.com/kazuma-desu/etu/workflows/CI/badge.svg)](https://github.com/kazuma-desu/etu/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/kazuma-desu/etu)](https://goreportcard.com/report/github.com/kazuma-desu/etu)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=kazuma-desu_etu&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=kazuma-desu_etu)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=kazuma-desu_etu&metric=coverage)](https://sonarcloud.io/summary/new_code?id=kazuma-desu_etu)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=kazuma-desu_etu&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=kazuma-desu_etu)

**Etcd Terminal Utility** - A modern CLI tool and Go library for managing etcd configurations with a kubectl-like UX.

## Features

- Multi-context management for multiple clusters
- Beautiful terminal output using [Charm](https://charm.sh/)
- Tree view visualization for hierarchical configuration paths
- Comprehensive validation (keys, values, JSON/YAML, URLs)
- Dry run mode to preview changes
- Diff command to compare local files with etcd state
- JSON/YAML input file support
- JSON output for automation
- Shell completion (bash, zsh, fish, PowerShell)
- Flexible configuration (file, env vars, CLI flags)

## Installation

### Using go install

```bash
go install github.com/kazuma-desu/etu@latest
```

### Download Binary

Download pre-built binaries for your platform from the [releases page](https://github.com/kazuma-desu/etu/releases).

```bash
# Linux (x86_64)
VERSION=v0.1.0  # Replace with latest version
curl -LO https://github.com/kazuma-desu/etu/releases/download/${VERSION}/etu_${VERSION#v}_linux_amd64.tar.gz
tar xzf etu_${VERSION#v}_linux_amd64.tar.gz
sudo mv etu /usr/local/bin/

# macOS (Apple Silicon)
VERSION=v0.1.0  # Replace with latest version
curl -LO https://github.com/kazuma-desu/etu/releases/download/${VERSION}/etu_${VERSION#v}_darwin_arm64.tar.gz
tar xzf etu_${VERSION#v}_darwin_arm64.tar.gz
sudo mv etu /usr/local/bin/

```

### From Source

```bash
git clone https://github.com/kazuma-desu/etu.git
cd etu && go build -o etu .
```

## Quick Start

```bash
# Login to cluster (interactive)
etu login

# Or with flags
etu login --context-name dev --endpoints http://localhost:2379 --no-auth

# Basic operations
etu get /app/config --prefix              # Get keys
etu put /app/config/host "localhost"      # Put key-value
etu delete /app/config/old --prefix       # Delete keys

# Apply configuration from file
etu apply -f config.txt --dry-run         # Preview first
etu apply -f config.txt                   # Then apply

# Compare local file with etcd state
etu diff -f config.txt
```

## Commands

Run `etu --help` for full command list. Key commands:

### Context Management

```bash
etu login                                 # Interactive setup
etu login --context-name prod --endpoints http://etcd:2379 --username admin --password secret

# TLS/mTLS
etu login --context-name prod --endpoints https://etcd:2379 \
  --cacert /path/to/ca.crt --cert /path/to/client.crt --key /path/to/client.key

etu config use-context <context>
etu config get-contexts
etu config current-context
etu config delete-context <context>
```

### Key Operations

```bash
etu ls <prefix>                           # List keys under prefix
etu ls /app -o json                       # List keys in JSON format
etu get <key> [--prefix] [--keys-only]    # Get keys with values
etu put <key> <value> [--dry-run]         # Put key-value
etu put <key> - < file.txt                # Put from stdin
etu delete <key> [--prefix] [--force]     # Delete keys
etu edit <key>                            # Edit in $EDITOR
```

### Configuration Files

```bash
etu apply -f <file> [--dry-run] [--strict]   # Apply to etcd
etu diff -f <file> [--prefix <p>] [--full]   # Compare with etcd
```

### Cluster Management

```bash
etu status                                # Show cluster health and status
etu status -o json                        # Show status in JSON format
etu status -o yaml                        # Show status in YAML format
```

The `status` command displays:
- Endpoint connectivity and health
- Server version
- Database size
- Leader information
- Raft index and term
- Any cluster errors

### Settings

```bash
etu config set log-level <debug|info|warn|error>
etu config set default-format <etcdctl|auto>
etu config set strict <true|false>
```

## Configuration

Config is stored in `~/.config/etu/config.yaml`:

```yaml
current-context: prod
log-level: warn
contexts:
  dev:
    endpoints:
      - http://localhost:2379
  prod:
    endpoints:
      - http://prod:2379
    username: admin
    password: secret
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `ETUCONFIG` | Custom path to config file (default: `~/.config/etu/config.yaml`) |


### Global Flags

Use `etu options` to see all global flags. Common flags visible in `--help`:

- `--context <name>`: Use specific context
- `--output <format>`: Output format (simple, json, table, tree)
- `--timeout <duration>`: Timeout for operations (default: 30s)
- `--log-level <level>`: Set log level (debug, info, warn, error)

Additional flags available via `etu options`:

- `--username`, `--password`: Override context credentials
- `--password-stdin`: Read password from stdin (for CI/CD)
- `--cacert`, `--cert`, `--key`: Override TLS certificates
- `--insecure-skip-tls-verify`: Skip TLS verification

## File Format (etcdctl)

```
/app/config/database/host
db.example.com

/app/config/database/port
5432

/app/i18n/welcome
en: Welcome
es: Bienvenido
```

Features: Auto type inference, multi-line values, dictionary parsing.

### Visualizing Configuration

Use `-o tree` flag to view your configuration hierarchically:

```bash
etu apply -f config.txt --dry-run -o tree
```

Output:
```
/
╰──app/
   ├──config/
   │  ├──database/
   │  │  ├──host db.example.com
   │  │  └──port 5432
   │  └──api/
   │     └──base_url https://api.example.com
   └──i18n/
      └──welcome [dict: 2 keys]
```

This makes it easy to understand the hierarchical structure of your etcd paths at a glance.

## Using as a Library

```go
import (
    "github.com/kazuma-desu/etu/pkg/client"
    "github.com/kazuma-desu/etu/pkg/parsers"
    "github.com/kazuma-desu/etu/pkg/validator"
)

func main() {
    cfg := &client.Config{Endpoints: []string{"http://localhost:2379"}}
    etcdClient, _ := client.NewClient(cfg)
    defer etcdClient.Close()

    parser, _ := parsers.NewRegistry().GetParser("etcdctl")
    pairs, _ := parser.Parse("config.txt")

    v := validator.NewValidator(false)
    if v.Validate(pairs).Valid {
        etcdClient.PutAll(context.Background(), pairs)
    }
}
```

## Project Structure

```
etu/
├── cmd/              # CLI commands
├── pkg/              # Public library API
│   ├── client/       # etcd client wrapper
│   ├── config/       # Configuration management
│   ├── parsers/      # Extensible parser system
│   ├── validator/    # Validation logic
│   └── output/       # Styled output
├── examples/         # Sample configs
└── main.go          # Entry point
```

## Development

```bash
go build -o etu .
go test ./...

# Test with local etcd
docker run -d --name etcd-test -p 2379:2379 \
  quay.io/coreos/etcd:v3.5.12 \
  /usr/local/bin/etcd \
  --listen-client-urls http://0.0.0.0:2379 \
  --advertise-client-urls http://0.0.0.0:2379

etu login --context-name dev --endpoints http://localhost:2379 --no-auth
./etu apply -f examples/sample.txt --dry-run
```

## Shell Completion

etu supports shell completion for bash, zsh, fish, and PowerShell.

### Bash

```bash
# Linux
etu completion bash > /etc/bash_completion.d/etu

# macOS (Homebrew)
etu completion bash > $(brew --prefix)/etc/bash_completion.d/etu
```

### Zsh

```bash
# Enable completion if not already
echo "autoload -U compinit; compinit" >> ~/.zshrc

# Add etu completions
etu completion zsh > "${fpath[1]}/_etu"

# Restart shell or source
source ~/.zshrc
```

### Fish

```bash
etu completion fish > ~/.config/fish/completions/etu.fish
```

### PowerShell

```powershell
# Add to current session
etu completion powershell | Out-String | Invoke-Expression

# Add to profile for persistence
etu completion powershell >> $PROFILE
```

## Exit Codes

etu uses standard exit codes for automation and scripting:

| Code | Meaning | Description |
|------|---------|-------------|
| 0 | Success | Command executed successfully |
| 1 | General error | An unexpected error occurred |
| 2 | Validation error | Invalid input, missing arguments, or validation failed |
| 3 | Connection error | Failed to connect to etcd cluster |
| 4 | Key not found | The requested key does not exist in etcd |

These codes can be used in shell scripts for error handling:

```bash
etu get /config/app/host
exit_code=$?

if [ $exit_code -eq 4 ]; then
    echo "Key not found, using default"
    host="localhost"
elif [ $exit_code -eq 0 ]; then
    echo "Key retrieved successfully"
else
    echo "Error occurred (exit code: $exit_code)"
    exit $exit_code
fi
```

## Roadmap

- Additional parsers (Helm, TOML)
- Watch operations
- Backup/restore commands

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, testing, and release process.


## License

MIT License - see [LICENSE](LICENSE).

## Acknowledgments

Built with [Charm](https://charm.sh/) and etcd's [Go client](https://github.com/etcd-io/etcd/tree/main/client/v3).

