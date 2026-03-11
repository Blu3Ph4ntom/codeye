---
title: "Output Formats"
description: "Structured and terminal-friendly output formats supported by codeye."
---

# Output Formats

Select a format with `--format <name>`.

## Terminal formats

| Format | Purpose |
|--------|---------|
| `table` | Default human-readable summary |
| `bar` | Horizontal language bars |
| `spark` | Sparkline-focused output |
| `compact` | Single-line summary for prompts or status bars |
| `markdown` | GitHub-friendly table output |

## Structured formats

| Format | Purpose |
|--------|---------|
| `json` | Full structured payload |
| `ndjson` | One JSON object per line |
| `csv` | Spreadsheet and reporting workflows |
| `badge` | Shields.io endpoint JSON |

## JSON example

```json
{
  "repo": "A:/Projects/codeye",
  "ref": "main",
  "tree_sha": "a5bc10236b9bcf55b5446dd55325359fcb6acbdd",
  "scan_ms": 55,
  "cached": true,
  "scanned_at": "2026-03-11T17:08:35.9428616Z",
  "total": {
    "name": "Total",
    "files": 62,
    "code": 6414,
    "blank": 900,
    "comment": 467,
    "lines": 7781
  },
  "languages": [
    {
      "name": "Go",
      "files": 31,
      "code": 4247,
      "blank": 500,
      "comment": 415,
      "lines": 5162,
      "pct": 66.34
    }
  ]
}
```

## Examples

```bash
# Pipe structured data into jq
codeye --format json . | jq '.total.lines'

# Generate a CSV artifact
codeye --format csv . > loc.csv

# Append a Markdown summary to a changelog or README
codeye --format markdown . >> REPORT.md

# Publish a Shields-compatible badge payload
codeye --format badge . > public/codeye-badge.json
```
