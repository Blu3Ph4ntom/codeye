# Changelog

All notable changes to this project will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
This project uses [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- GitHub Pages deployment workflow for `codeye.bluephantom.dev`
- Community health files: contributing guide, security policy, support guide, code of conduct, issue templates, and pull request template
- Tests for layered config loading and comma-separated language filters

### Changed

- Activated real `.codeye.toml` and `CODEYE_*` config resolution in the CLI
- Added compatibility aliases for older flags such as `--ref` and `--hotspot`
- Refreshed the README, website, and docs to match actual current behavior
- Simplified release metadata so GoReleaser targets the real repository and available assets

## [0.1.0] — 2024-01-01

### Added

- Initial release
- Git-native LoC scanner using `git ls-files` — only tracked files counted
- 150+ language detection by extension, filename, and shebang
- Parallel goroutine pool scanner for fast throughput
- Content-addressable cache keyed by git tree SHA
- Output formats: table, bar, json, ndjson, csv, badge (shields.io), compact
- Blame mode — lines of code aggregated per author email
- History mode — LoC growth over time with day/week/month/quarter/year buckets
- Hotspot mode — file churn score (commits × lines changed)
- `diff` subcommand — compare LoC between two git refs
- `langs` subcommand — list all supported languages
- `doctor` subcommand — system health checks
- `cache status/clear` subcommands
- `completion` subcommand — bash, zsh, fish, and powershell completions
- `version` subcommand with build metadata (commit, date, runtime)
- Config file support via `.codeye.toml` (TOML format, auto-detected)
- Vendor/generated/test file exclusion filters
- Glob-based include/exclude patterns
- `--top N` to limit output to top N languages
- `--lang` filter for specific languages
- `--sort` by lines/code/blank/comment/files/lang
- `--pct` percentage column
- `--emoji` language emoji in table output
- `--wide` all columns mode
- `--no-color` for CI / pipe output
- Worker pool size configurable via `--workers`
- Static binary with CGO_ENABLED=0 — no system dependencies
- Cross-platform builds: Linux (amd64/arm64/386/armv7), macOS (amd64/arm64), Windows (amd64), FreeBSD (amd64)
- Makefile with build/test/bench/cross/snapshot/release targets
- GoReleaser config with deb/rpm/apk packages, brew formula, checksums, cosign signatures
- GitHub Actions CI: multi-platform test matrix, benchmarks, golangci-lint, codecov coverage
- GitHub Actions release workflow with GoReleaser
- One-line install script (`install.sh`)
