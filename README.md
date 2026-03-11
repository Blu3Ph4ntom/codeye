# codeye

Fast repository metrics for git projects.

`codeye` scans tracked files, counts code/comment/blank lines per language, and adds git-aware analysis modes for ownership, history, hotspots, and ref-to-ref diffs.

Website: [codeye.bluephantom.dev](https://codeye.bluephantom.dev)  
Repository: [github.com/Blu3Ph4ntom/codeye](https://github.com/Blu3Ph4ntom/codeye)

```text
$ codeye .

codeye · main · 62 files · 55ms
───────────────────────────────────────────────────────────────────
 Language      Files       Code      Total
───────────────────────────────────────────────────────────────────
▌ Go              31      4,247      5,162
▌ Markdown         8        670        879
▌ CSS              1        526        559
───────────────────────────────────────────────────────────────────
 Total           62      6,414      7,781
───────────────────────────────────────────────────────────────────
 ✓ 55ms · cache hit · a5bc1023
```

## Features

- Git-native tracked-file scans with `.codeyeignore` support
- Fast repeat runs using a content-addressed cache
- Language totals with code, comment, and blank line counts
- Blame rollups with per-author line ownership
- Repository growth history over time
- Hotspot analysis for high-churn files
- Ref-to-ref diffs for release and branch comparisons
- Human and machine-friendly output formats

## Install

### Go install

```sh
go install github.com/blu3ph4ntom/codeye/cmd/codeye@latest
```

### Installer scripts

```sh
curl -sSfL https://codeye.bluephantom.dev/install.sh | sh
```

```powershell
iex (irm https://codeye.bluephantom.dev/install.ps1)
```

### Release archives

Download prebuilt binaries from the [releases page](https://github.com/Blu3Ph4ntom/codeye/releases).

## Usage

```sh
# Current repository snapshot
codeye .

# Scan a specific ref
codeye --branch main .
codeye --branch HEAD~5 .

# Filter output
codeye --lang Go,Markdown .
codeye --top 10 .
codeye --no-vendor --no-generated .

# Analysis modes
codeye --blame .
codeye --history --history-interval month .
codeye --hotspots --top 20 .
codeye diff v0.1.0 HEAD

# Structured output
codeye --format json .
codeye --format csv .
codeye --format markdown .

# Utility commands
codeye doctor
codeye cache status
codeye langs
```

`--ref`, `--interval`, and `--hotspot` are kept as compatibility aliases for older docs.

## Configuration

Create `.codeye.toml` in a repository root:

```toml
format = "table"
sort = "code"
top = 12
lang = ["Go", "TypeScript"]
no_vendor = true
no_generated = true
workers = 8
```

Configuration precedence:

1. Built-in defaults
2. `~/.codeye.toml`
3. The nearest project `.codeye.toml`
4. `CODEYE_*` environment variables
5. CLI flags

## Development

```sh
make test
make build
make web
```

CI runs Go tests, linting, and a Hugo site build. GitHub Pages deploys the website from `web/` and serves the custom domain via `web/static/CNAME`.
