---
title: "Usage"
description: "Command reference for codeye."
---

# Usage

## Basic scans

```bash
# Scan the current repository
codeye .

# Scan a specific branch, tag, or commit
codeye --branch main .
codeye --branch HEAD~5 .

# Filter to specific languages
codeye --lang Go,Markdown .
codeye --lang Go --lang TypeScript .
```

`--ref` still works as a compatibility alias for `--branch`, but `--branch` is the canonical flag.

## Main flags

| Flag | Default | Description |
|------|---------|-------------|
| `--branch <ref>` | `HEAD` | Scan a branch, tag, or commit SHA |
| `--format <fmt>` | `table` | `table`, `bar`, `spark`, `json`, `ndjson`, `csv`, `badge`, `markdown`, `compact` |
| `--sort <field>` | `lines` | `lines`, `files`, `code`, `blank`, `comment`, `lang` |
| `--top <n>` | `0` | Limit output rows |
| `--lang <list>` | all | Comma-separated or repeated language filter |
| `--exclude <glob>` | none | Exclude paths matching one or more globs |
| `--path-filter <glob>` | none | Include only matching paths |
| `--no-vendor` | `false` | Exclude vendor-like directories |
| `--no-generated` | `false` | Exclude generated files such as `*.pb.go` |
| `--no-tests` | `false` | Exclude common test file patterns |
| `--min-lines <n>` | `0` | Hide languages below a line threshold |
| `--no-cache` | `false` | Skip cache read and write |
| `--no-color` | `false` | Disable ANSI colors |
| `--no-header` | `false` | Suppress header and footer chrome |
| `--wide` | `false` | Show blank and comment columns in table output |
| `--compact` | `false` | Force compact terminal rendering |
| `--pct` | `true` | Show percentage column |
| `--theme <name>` | `dark` | `dark`, `light`, `mono` |
| `--workers <n>` | `GOMAXPROCS` | Worker pool size |
| `--dry-run` | `false` | Print matching files without scanning |
| `--config <path>` | auto | Explicit `.codeye.toml` path |
| `--version` | `false` | Print version information and exit |

## Analysis modes

```bash
# Ownership by author
codeye --blame .

# Repository growth
codeye --history --history-interval month .

# High-churn files
codeye --hotspots --top 20 .

# Snapshot the repo at a historical date
codeye --at 2025-12-31 .
```

Related flags:

| Flag | Description |
|------|-------------|
| `--blame` | Aggregate line ownership by author |
| `--history` | Show repository growth over time |
| `--history-interval <unit>` | `day`, `week`, `month`, `quarter`, or `year` |
| `--history-limit <n>` | Maximum commits walked for history and hotspots |
| `--hotspots` | Show most-changed files by churn score |
| `--since <date>` | Limit history or hotspot analysis to newer commits |
| `--until <date>` | Limit history analysis to older commits |
| `--at <date>` | Resolve the repo state at or before a date |
| `--speedtest` | Run three uncached scans and print timings |

`--interval` and `--hotspot` remain available as compatibility aliases for older docs.

## Subcommands

```bash
codeye diff <ref1> <ref2>
codeye langs
codeye doctor
codeye cache status
codeye cache clear
codeye completion bash
codeye version
```

## Non-git directories

If the target path is not a git repository, `codeye` falls back to a direct directory walk for standard scans. Git-dependent analysis modes such as `--blame`, `--history`, and `--hotspots` still require a repository.
