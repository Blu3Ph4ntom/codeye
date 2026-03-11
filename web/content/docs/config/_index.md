---
title: "Configuration"
description: "Configure codeye with .codeye.toml and environment variables."
---

# Configuration

`codeye` runs with sensible defaults, but now supports layered configuration:

1. Built-in defaults
2. `~/.codeye.toml`
3. The nearest project `.codeye.toml`
4. `CODEYE_*` environment variables
5. CLI flags

## Example `.codeye.toml`

```toml
format = "table"
sort = "code"
top = 12
lang = ["Go", "TypeScript"]
exclude = ["testdata/**", "*.snap"]
no_vendor = true
no_generated = true
workers = 8
theme = "mono"
```

Project config files are discovered by walking upward from the target directory. Use `--config path/to/.codeye.toml` to force a specific file.

## Supported keys

| Key | Type | Notes |
|-----|------|-------|
| `branch` | string | Default ref for scans |
| `format` | string | Output format |
| `sort` | string | Sort field |
| `top` | integer | Max rows |
| `lang` | array or CSV string | Language filter |
| `exclude` | array or CSV string | Exclusion globs |
| `path_filter` | array or CSV string | Inclusion globs |
| `no_vendor` | boolean | Exclude vendor-like directories |
| `no_generated` | boolean | Exclude generated files |
| `no_tests` | boolean | Exclude common test patterns |
| `min_lines` | integer | Minimum lines per language |
| `no_color` | boolean | Disable ANSI color |
| `no_header` | boolean | Hide table header and footer |
| `compact` | boolean | Force compact rendering |
| `pct` | boolean | Show percentage column |
| `theme` | string | `dark`, `light`, `mono` |
| `history_interval` | string | `day`, `week`, `month`, `quarter`, `year` |
| `history_limit` | integer | Commit limit for history and hotspots |
| `cache_dir` | string | Override cache location |
| `workers` | integer | Worker pool size |
| `verbose` | boolean | Debug logging |

## Environment variables

Common overrides:

| Variable | Description |
|----------|-------------|
| `CODEYE_CONFIG` | Explicit config file path |
| `CODEYE_FORMAT` | Default output format |
| `CODEYE_WORKERS` | Worker pool size |
| `CODEYE_CACHE_DIR` | Cache directory override |
| `CODEYE_NERD_FONTS` | Set to `1` or `true` to enable Nerd Font glyphs |
| `CODEYE_NO_COLOR` | Disable color output |
| `NO_COLOR` | Standard no-color override |

Environment variables are read after config files and before CLI flags.

## `.codeyeignore`

If the repository root contains `.codeyeignore`, its patterns are passed to `git ls-files --exclude-from=.codeyeignore` during git-backed scans.

```gitignore
vendor/
dist/
*.pb.go
testdata/
```

This is useful for persistent exclusions that should not live in every command invocation.
