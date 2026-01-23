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
# Login to cluster
etu login dev --endpoints http://localhost:2379 --no-auth

# Parse and preview configuration
etu parse -f config.txt

# View configuration as a tree
etu parse -f config.txt --tree

# Validate before applying
etu validate -f config.txt

# Apply to etcd
etu apply -f config.txt --dry-run  # preview first
etu apply -f config.txt            # then apply
```

## Commands

### Context Management

```bash
etu login <context> --endpoints <url> [--username <user>] [--password <pass>]
etu config use-context <context>
etu config get-contexts
etu config delete-context <context>
```

### Configuration Operations

```bash
etu parse -f <file> [--json|--tree]   # Parse and display
etu validate -f <file> [--strict]     # Validate configuration
etu apply -f <file> [--dry-run]       # Apply to etcd
etu diff -f <file>                    # Compare file with etcd state
etu put <key> <value>                 # Put single key-value
etu get <key> [--prefix]              # Get keys
etu delete <key> [--prefix]           # Delete keys
```

#### Parse Output Formats

- **Default**: List view showing all key-value pairs
- **Tree view** (`--tree`): Hierarchical visualization of configuration paths
- **JSON** (`--json`): Machine-readable output for automation

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
| `ETCD_ENDPOINTS` | Comma-separated endpoints |
| `ETCD_USERNAME` | Username for auth |
| `ETCD_PASSWORD` | Password for auth |

### Global Flags

- `--context <name>`: Use specific context
- `--log-level <level>`: Set log level (debug, info, warn, error)
- `--dry-run`: Preview without applying
- `--json`: JSON output

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

Use the `--tree` flag to view your configuration hierarchically:

```bash
etu parse -f config.txt --tree
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

./etu login dev --endpoints http://localhost:2379 --no-auth
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

## Roadmap

- Additional parsers (Helm, TOML)
- Watch operations
- TLS support
- Backup/restore commands

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, testing, and release process.


## License

MIT License - see [LICENSE](LICENSE).

## Acknowledgments

Built with [Charm](https://charm.sh/) and etcd's [Go client](https://github.com/etcd-io/etcd/tree/main/client/v3).
