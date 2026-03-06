# codeye

Fast lines-of-code scanner for git repositories.

```
$ codeye .

codeye · main · 25 files · 10ms
───────────────────────────────────────────────────────────────────
 Language    Files       Code      Total
───────────────────────────────────────────────────────────────────
▌ Go             23      2,672      3,204
▌ Markdown        1        142        188
▌ Gitignore       1         26         33
───────────────────────────────────────────────────────────────────
 Total          25      2,840      3,425
───────────────────────────────────────────────────────────────────
 ⚡ 10ms · cache miss · tree 16b20023
```

## Features

- **Git-native**: scans only tracked files via `git ls-files`, respects `.gitignore`
- **Fast**: goroutine pool, content-addressable cache, typically <100ms for large repos
- **150+ languages**: detected by extension, filename, and shebang line
- **7 output formats**: table, bar, json, ndjson, csv, badge (shields.io), compact
- **Blame mode**: lines of code per author (`codeye --blame`)
- **History mode**: LoC growth over time (`codeye --history`)
- **Hotspot mode**: churn analysis by file (`codeye --hotspots`)
- **Diff mode**: compare two git refs (`codeye diff HEAD~1 HEAD`)
- **Auto-cache**: tree-SHA keyed, instant repeat scans
- **No CGO**: pure Go, single static binary

## Install

### Homebrew (macOS / Linux)

```sh
brew install blu3ph4ntom/tap/codeye
```

### One-line installer (Linux / macOS)

```sh
curl -sSfL https://raw.githubusercontent.com/codeye/codeye/main/install.sh | sh
```

### Go install

```sh
go install github.com/blu3ph4ntom/codeye/cmd/codeye@latest
```

### Download binary

Pre-built binaries for Linux, macOS, Windows, and FreeBSD are available on the [releases page](https://github.com/blu3ph4ntom/codeye/releases).

```sh
# Linux amd64
curl -sSfL https://github.com/blu3ph4ntom/codeye/releases/latest/download/codeye_linux_amd64.tar.gz | tar xz
sudo mv codeye /usr/local/bin/
```

## Usage

```sh
# Scan working tree
codeye .
codeye /path/to/repo

# Scan a specific git ref
codeye --ref=HEAD~5 .

# Filter output
codeye --top=10 .                 # top 10 languages
codeye --lang=Go,TypeScript .     # specific languages only
codeye --sort=code .              # sort by code lines (default: lines)
codeye --no-vendor .              # exclude vendor/
codeye --no-tests .               # exclude test files
codeye --no-generated .           # exclude generated files

# Output formats
codeye --format=json .            # structured JSON
codeye --format=ndjson .          # newline-delimited JSON
codeye --format=csv .             # CSV
codeye --format=bar .             # horizontal bar chart
codeye --format=badge .           # shields.io badge JSON
codeye --format=compact .         # single-line summary

# Blame: lines per author
codeye --blame .
codeye --blame --ref=HEAD .

# History: LoC growth over time
codeye --history .
codeye --history --interval=month .

# Hotspots: churn analysis
codeye --hotspots .
codeye --hotspots --top=20 .

# Diff two refs
codeye diff HEAD~10 HEAD
codeye diff v0.1.0 v0.2.0

# Show all supported languages
codeye langs

# Shell completion
codeye completion bash   >> ~/.bash_completion.d/codeye
codeye completion zsh    > ~/.zsh/completions/_codeye
codeye completion fish   > ~/.config/fish/completions/codeye.fish

# System health check
codeye doctor
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `table` | Output format: table, bar, json, ndjson, csv, badge, compact |
| `--sort` | `lines` | Sort by: lines, code, blank, comment, files, lang |
| `--top N` | `0` (all) | Show only top N languages |
| `--lang` | — | Comma-separated language filter |
| `--ref` | HEAD | Git ref to scan |
| `--no-vendor` | false | Exclude vendor/ directories |
| `--no-generated` | false | Exclude auto-generated files |
| `--no-tests` | false | Exclude test files |
| `--no-cache` | false | Skip cache lookup |
| `--no-color` | false | Disable color output |
| `--wide` | false | Show all columns |
| `--compact` | false | One file per line, no table chrome |
| `--pct` | false | Show percentage column |
| `--emoji` | false | Language emoji in table |
| `--workers N` | `runtime.NumCPU()` | Goroutine pool size |
| `--exclude` | — | Glob patterns to exclude |
| `--include` | — | Glob patterns to include (empty = all) |
| `--min-lines N` | `0` | Skip languages with fewer than N lines |
| `--blame` | false | Lines of code per author |
| `--history` | false | LoC growth over time |
| `--hotspots` | false | File churn analysis |
| `--interval` | `month` | History bucket: day, week, month, quarter, year |
| `--since` | — | Limit history/blame to commits after date |
| `--until` | — | Limit history/blame to commits before date |
| `--limit N` | `100` | Max commits for history/blame |
| `--config` | — | Config file (default: .codeye.toml) |

## Configuration

Create `.codeye.toml` in your repository root:

```toml
format     = "table"
sort       = "code"
top        = 15
no_vendor  = true
no_generated = true
workers    = 8
exclude    = ["testdata/**", "*.pb.go"]
```

## Output Formats

### JSON

```sh
codeye --format=json . | jq '.total'
```

```json
{
  "languages": [...],
  "total": {
    "files": 25,
    "lines": 3425,
    "code": 2840,
    "blank": 412,
    "comment": 173
  },
  "meta": {
    "ref": "HEAD",
    "tree_sha": "16b20023",
    "scanned_at": "2024-01-01T00:00:00Z",
    "elapsed_ms": 10,
    "cache_hit": false
  }
}
```

### Badge (shields.io)

```sh
codeye --format=badge . > badge.json
```

Use as a dynamic badge: `https://img.shields.io/endpoint?url=<hosted-badge-json-url>`

### Compact (one-liner)

```sh
codeye --format=compact .
# → 3425 lines · 25 files · Go TypeScript Markdown · 10ms
```

## Shell Completions

**bash:**
```sh
mkdir -p ~/.bash_completion.d
codeye completion bash > ~/.bash_completion.d/codeye
echo 'source ~/.bash_completion.d/codeye' >> ~/.bashrc
```

**zsh:**
```sh
codeye completion zsh > "${fpath[1]}/_codeye"
```

**fish:**
```sh
codeye completion fish > ~/.config/fish/completions/codeye.fish
```

## Performance

| Repo | Files | Lines | Time |
|------|-------|-------|------|
| codeye (itself) | 25 | 3,425 | 8ms |
| kubernetes/kubernetes | ~4,000 | ~1.8M | ~120ms |
| torvalds/linux | ~35,000 | ~28M | ~900ms |

Repeated scans hit the content-addressable cache and complete in <5ms regardless of repo size.

## Building from Source

```sh
git clone https://github.com/blu3ph4ntom/codeye.git
cd codeye
make build
./codeye --version
```

Requirements: Go 1.22+, git 2.x

```sh
make test      # run all tests
make bench     # run benchmarks
make cross     # cross-compile for all platforms
make snapshot  # goreleaser dry run
```

## License

MIT — see [LICENSE](LICENSE)
