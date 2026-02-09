# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Global flags (`--output`, `--timeout`, `--log-level`, etc.) are now hidden from main help output for clarity. Use `etu options` to view all global flags.
- Watch command now uses the global `-o, --output` flag instead of a boolean flag. Use `-o simple` (default) for raw values or `-o json` for full event JSON.

### Removed
- **BREAKING**: Removed `fields` output format. Use `etu get --show-metadata -o simple` instead to display key metadata alongside values.
