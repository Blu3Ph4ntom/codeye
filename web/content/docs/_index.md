---
title: "Documentation"
date: 2024-03-12T12:00:00+00:00
draft: false
---

`codeye` is a CLI tool designed to give you instant line-of-code statistics about your git repositories. It uses direct git porcelain commands to ensure the highest degree of accuracy possible.

## Features

- **No CGO**: Statically compiled binary. Zero dependencies.
- **Fast**: Sub-100ms scans for medium to large repos.
- **Git Porcelain**: Does not rely on complex, inconsistent git libraries; uses the CLI you already have.
- **Rich Formats**: JSON, CSV, Sparklines, Tables, and Badge output.

---

## Installation

```bash
go install github.com/codeye/codeye/cmd/codeye@latest
```

Ensure your `$GOPATH/bin` or `$HOME/go/bin` is in your `$PATH`.

## Quick Start

### Basic Scan
Perform a local scan of the current directory.
```bash
codeye
```

### Blame Analysis
Determine who has contributed the most to the current tree by line count.
```bash
codeye --blame
```

### History Trend
View cumulative line growth over the life of the repository.
```bash
codeye --history
```

## Options

| Flag | Description |
|------|-------------|
| `--blame` | Toggle per-author blame analysis |
| `--history` | Show historical growth chart |
| `--format` | Output format (table, json, csv, sparkline, badge) |
| `--wide` | Detailed view including blank lines and comment counts |
| `--no-cache` | Force a full re-scan, bypassing `$CACHE_DIR` |

---

## Benchmarks

Repository: Kubernetes (main)  
Files: ~21,000  
`codeye` Scan Time: **~195ms** (on Apple M2 Pro)

---

## Contributing

We love gophers. PRs are welcome. Make sure to run `make test` before submitting.

<img src="https://golang.org/doc/gopher/pkg.png" class="gopher" alt="Gopher Package">
