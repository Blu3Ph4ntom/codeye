---
title: "Configuration"
description: "Configure codeye via .codeye.toml or environment variables."
---

# Configuration

codeye is zero-config by default, but everything is tuneable.

## Config file

Place `.codeye.toml` in your repo root or home directory (`~/.codeye.toml`):

```toml
# .codeye.toml

# Default output format
format = "table"

# Always show top 10 languages only
top = 10

# Sort by code lines, not total lines
sort = "code"

# Enable Nerd Font icons if your terminal has one
nerd_font = false

# Disable emoji (useful for non-UTF-8 environments)
emoji = true

# Cache directory override
cache_dir = "/tmp/codeye-cache"

# Worker count (default: GOMAXPROCS)
workers = 8
```

## Environment variables

All settings can be set via environment variables. They take priority over the config file.

| Variable | Description |
|----------|-------------|
| `CODEYE_FORMAT` | Default output format |
| `CODEYE_CACHE_DIR` | Override cache directory |
| `CODEYE_WORKERS` | Number of parallel goroutines |
| `CODEYE_NERD_FONTS` | Set to `1` to enable Nerd Font glyphs |
| `CODEYE_NO_COLOR` | Set to `1` to disable color |
| `CODEYE_CONFIG` | Path to config file |
| `NO_COLOR` | Standard — disables all color output |

## Cache

codeye caches scan results keyed by `(repo path, tree SHA)`. Cache is invalidated automatically when the tree changes.

```bash
# Default location
~/.cache/codeye/          # Linux / macOS
%LOCALAPPDATA%\codeye\    # Windows

# Override
CODEYE_CACHE_DIR=/tmp/codeye codeye

# Bypass for one run
codeye --no-cache
```

## .codeyeignore

To exclude paths from all scans, create `.codeyeignore` in the repo root:

```gitignore
# .codeyeignore — same syntax as .gitignore

vendor/
*_generated.go
*.pb.go
testdata/
web/node_modules/
```

Paths in `.codeyeignore` are excluded from `ls-files` output before any counting begins.

## Priority order

Settings are resolved in this order (highest wins):

1. CLI flags
2. Environment variables
3. `.codeye.toml` in current directory
4. `.codeye.toml` in home directory
5. Built-in defaults
