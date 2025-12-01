# etu

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev/)

**Etcd Terminal Utility** - A modern CLI tool and Go library for managing etcd configurations with a kubectl-like UX.

## Features

- Multi-context management for multiple clusters
- Beautiful terminal output using [Charm](https://charm.sh/)
- Comprehensive validation (keys, values, JSON/YAML, URLs)
- Dry run mode to preview changes
- JSON output for automation
- Flexible configuration (file, env vars, CLI flags)

## Installation

```bash
# Using go install
go install github.com/kazuma-desu/etu@latest

# From source
git clone https://github.com/kazuma-desu/etu.git
cd etu && go build -o etu .
```

Download binaries from the [releases page](https://github.com/kazuma-desu/etu/releases).

## Quick Start

```bash
# Login to cluster
etu login dev --endpoints http://localhost:2379 --no-auth

# Parse and preview configuration
etu parse -f config.txt

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
etu parse -f <file> [--json]          # Parse and display
etu validate -f <file> [--strict]     # Validate configuration
etu apply -f <file> [--dry-run]       # Apply to etcd
```

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

## Security

Passwords are stored in plain text in config (like Docker). For better security:
- Use `--password` flag at runtime
- Use `ETCD_PASSWORD` environment variable
- Avoid storing credentials in config for production

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
  -e ALLOW_NONE_AUTHENTICATION=yes bitnami/etcd:latest

./etu login dev --endpoints http://localhost:2379 --no-auth
./etu apply -f examples/sample.txt --dry-run
```

## Roadmap

- Additional parsers (Helm, TOML, JSON, YAML)
- Get/watch operations
- Diff operations
- TLS support
- Shell completion

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT License - see [LICENSE](LICENSE).

## Acknowledgments

Built with [Charm](https://charm.sh/) and etcd's [Go client](https://github.com/etcd-io/etcd/tree/main/client/v3).
