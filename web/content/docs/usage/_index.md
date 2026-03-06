---
title: "Usage"
description: "Complete reference for all codeye flags and commands."
---

# Usage

## Basic scan

```bash
# Scan HEAD of the current repo
codeye

# Scan a specific branch or commit
codeye --ref main
codeye --ref HEAD~5
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--ref <ref>` | `HEAD` | Git ref to scan (branch, tag, commit SHA) |
| `--format <fmt>` | `table` | Output format: `table`, `json`, `csv`, `markdown`, `badge`, `compact` |
| `--top <n>` | `0` (all) | Show only top N languages |
| `--sort <field>` | `lines` | Sort by: `lines`, `code`, `files`, `name` |
| `--desc` | `true` | Sort descending |
| `--no-cache` | `false` | Skip cache read/write |
| `--dry-run` | `false` | Print what would be scanned, don't count |
| `--emoji` | `true` | Show language emoji icons |
| `--nf` | `false` | Use Nerd Font glyphs (requires patched font) |
| `--pct` | `true` | Show percentage column |
| `--blame` | `false` | Show per-author line ownership |
| `--history` | `false` | Show week-by-week LoC growth |
| `--hotspot` | `false` | Show most-churned files |
| `--verbose` | `false` | Verbose stderr logging |
| `--version` | | Print version and exit |

## Blame mode

Shows which author owns how many lines across the whole repo:

```bash
codeye --blame
```

```
Author                   Lines      %   Files
────────────────────────────────────────────────
alice@example.com        5,591  89.1%      38  ████████████████████████████████
bob@example.com            685  10.9%       7  ████

total: 6,276 lines across 2 authors
```

Blame runs `git blame --porcelain` on every file in parallel and aggregates by author email.

## History mode

Shows lines-of-code growth bucketed by week, with a sparkline:

```bash
codeye --history
codeye --history --since 2024-01-01
codeye --history --until 2024-06-01
```

Additional history flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--since <date>` | (unbounded) | Start date `YYYY-MM-DD` |
| `--until <date>` | (unbounded) | End date `YYYY-MM-DD` |
| `--bucket <unit>` | `week` | Time bucket: `day`, `week`, `month` |

## Hotspot mode

Shows the files with the highest churn-to-size ratio:

```bash
codeye --hotspot
codeye --hotspot --top 10
```

## Output formats

```bash
codeye --format json   | jq .
codeye --format csv    > loc.csv
codeye --format badge  # Shields.io-compatible JSON
codeye --format markdown >> README.md
```

## Excluding paths

Create a `.codeyeignore` file in the repo root — same syntax as `.gitignore`:

```
vendor/
*.pb.go
testdata/
node_modules/
```

Or pass exclusion patterns inline:

```bash
codeye --exclude vendor/ --exclude '*.pb.go'
```
