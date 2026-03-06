---
title: "Output Formats"
description: "All output formats supported by codeye."
---

# Output Formats

Select a format with `--format <name>` or `CODEYE_FORMAT=<name>`.

## table (default)

Human-readable, colour-coded terminal table with per-language breakdown:

```
codeye · main · 45 files · 11ms
──────────────────────────────────────────────────────────────────
  Language           Files      Code  Comments    Blanks     Total      %
──────────────────────────────────────────────────────────────────
▌ Go                    28     3,733       272       441     4,446  77.0%
▌ TypeScript            12     2,104       188       310     2,602  16.1%
▌ Markdown               4       433         0       126       559   3.5%
▌ Shell                  1        56         3        14        73   0.5%
──────────────────────────────────────────────────────────────────
  Total                 45     6,326       463       891     7,680 100.0%
```

## compact

Single-line summary, useful for status bars or quick checks:

```
Go 77% · TS 16% · MD 3.5% · 45 files · 7,680 lines
```

## json

Machine-readable, complete data:

```json
{
  "ref": "main",
  "tree": "dc7f118f...",
  "scanned_at": "2026-03-06T10:00:00Z",
  "elapsed_ms": 11,
  "languages": [
    {
      "name": "Go",
      "files": 28,
      "code": 3733,
      "comments": 272,
      "blanks": 441,
      "total": 4446,
      "pct": 77.0
    }
  ],
  "totals": {
    "files": 45,
    "code": 6326,
    "comments": 463,
    "blanks": 891,
    "total": 7680
  }
}
```

Usage with `jq`:

```bash
codeye --format json | jq '.languages[] | select(.name == "Go") | .code'
```

## csv

Spreadsheet-friendly:

```csv
Language,Files,Code,Comments,Blanks,Total,Pct
Go,28,3733,272,441,4446,77.0
TypeScript,12,2104,188,310,2602,16.1
```

```bash
codeye --format csv > loc.csv
```

## markdown

Ready to paste into a README or wiki:

```bash
codeye --format markdown
```

Produces a GitHub Flavored Markdown table. Pipe it straight into your README:

```bash
codeye --format markdown >> README.md
```

## badge

[Shields.io](https://shields.io/endpoint) endpoint-compatible JSON. Host it behind a URL and use as a live badge:

```json
{
  "schemaVersion": 1,
  "label": "lines of code",
  "message": "7,680",
  "color": "brightgreen"
}
```

```bash
# Write to a file served by your badge endpoint
codeye --format badge > public/loc-badge.json
```

Then in your README:

```markdown
![Lines of Code](https://img.shields.io/endpoint?url=https://yoursite.dev/loc-badge.json)
```
