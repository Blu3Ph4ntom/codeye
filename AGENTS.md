# AGENTS.md — codeye build instructions for autonomous coding agents

## Project identity

`codeye` — CLI Lines-of-Code scanner for git repositories.
Single static binary. Sub-100ms scans. Git porcelain only. No CGO.

Module path: `github.com/codeye/codeye`
Language: Go 1.22+
Build constraint: `CGO_ENABLED=0`

---

## Repository layout

```
codeye/
├── cmd/codeye/          main.go — cobra entrypoint
├── internal/
│   ├── git/             git porcelain wrappers (ls-files, log, blame, shortlog)
│   ├── scanner/         parallel LoC scanner, language detection, ignore rules
│   ├── blame/           git blame aggregation, per-author rollup
│   ├── history/         git log --numstat walker, growth time-series
│   ├── output/          renderers: table, bar, sparkline, json, csv, badge
│   ├── config/          .codeye.toml + env-var resolution via viper
│   └── perf/            speedtest harness, pprof helpers
├── testdata/            fixture repos (git bundles), golden output files
├── .github/workflows/   CI: test, lint, bench, release
├── .goreleaser.yaml     cross-platform release config
├── Makefile             build / test / bench / release targets
├── AGENTS.md            ← this file
└── .agent/PLAN.md       full product plan (gitignored)
```

---

## Non-negotiable rules

1. **CGO_ENABLED=0** always. No CGO, no exceptions. Every external Go lib must be pure Go.
2. **No os.Walk** for file enumeration. Use `git ls-files -z` exclusively.  
3. **No regex in hot paths**. Line counting uses `bytes.Count(buf, []byte{'\n'})`.
4. **Single binary target**. `go build ./cmd/codeye` must produce one self-contained binary.
5. **Zero network calls at runtime**. All git operations via local git CLI subprocess only.
6. **`git` porcelain** (`--porcelain`, `-z` flags) wherever available. Never parse non-porcelain output.
7. **Tests must pass before committing**. Run `make test` before every commit.
8. **Commit messages**: lowercase, imperative, ≤72 chars. No tool names, no branding, no "feat:" prefixes. Just what changed.
9. **No print debugging left in committed code**. Use `--verbose` flag + stderr logging only.
10. **Cross-platform**: all file paths via `filepath.Join`, all colors auto-detected, `NO_COLOR` respected.

---

## Core algorithm (must be preserved)

```
1. git rev-parse HEAD → tree SHA
2. check cache: $CACHE_DIR/<repo_hash>/<tree_sha>.bin
3. cache hit  → deserialize + render → exit
4. cache miss → git ls-files -z [flags]
5.            → goroutine pool (GOMAXPROCS workers)
6.            → each worker: read file bytes, count \n, detect lang, strip blanks/comments
7.            → merge map[string]LangStats via channels
8.            → write cache (msgpack)
9.            → render output format → stdout
```

---

## Language detection order (internal/scanner/langs.go)

1. Exact filename match (`Makefile`, `Dockerfile`, `Jenkinsfile`)
2. Extension match, case-insensitive (`.go`, `.py`, `.rs`) — O(1) map lookup
3. Shebang parse: read first 128 bytes, check `#!` line
4. Fallback: `"Unknown"`

Detection must never panic on any input. Fuzz-tested.

---

## Performance invariants (benchmarks enforce these)

| Scenario | p95 target |
|----------|-----------|
| Cache hit, any size | ≤15ms |
| 100k line repo, cache miss | ≤100ms |
| 1M line repo, cache miss | ≤800ms |

Run `make bench` to validate. CI fails if p95 regresses >20%.

---

## Adding a new language

Edit `internal/scanner/langs.go`:
1. Add entry to `extensionMap` (or `filenameMap` for filename-matched langs)
2. Add `LangDef` struct with `Name`, `LineComment`, `BlockCommentStart`, `BlockCommentEnd`
3. Add golden test file to `testdata/langs/<name>/sample.<ext>` with expected counts
4. Run `go test ./internal/scanner/` — must pass

---

## Adding a new output format

1. Create `internal/output/<format>.go`
2. Implement the `Renderer` interface:
   ```go
   type Renderer interface {
       Render(result *ScanResult, opts RenderOpts) error
   }
   ```
3. Register in `internal/output/registry.go`
4. Add `--format <name>` to the cobra flag enum
5. Add snapshot test in `internal/output/<format>_test.go`

---

## Adding a new CLI flag

1. Add flag in `cmd/codeye/main.go` via `cmd.Flags().XxxVar(...)`
2. Add field to `internal/config/config.go` `Config` struct
3. Wire through to the relevant internal package
4. Document in flag table in `.agent/PLAN.md`
5. Add test covering the new flag behavior

---

## Build commands

```bash
export PATH="$HOME/go/bin:$PATH"
export CGO_ENABLED=0

make build          # compile for current OS/arch
make test           # go test -race -count=1 ./...
make bench          # go test -bench=. -benchmem ./...
make lint           # golangci-lint run
make snapshot       # goreleaser snapshot (local multi-platform build)
make release        # goreleaser release (tag required)
```

---

## Commit workflow for agents

```
1. make test              — must be green
2. git add -A
3. git commit -m "<minimal imperative message>"
4. continue building next feature
```

Commit after every logical unit of work. Aim for ≥1 commit per internal package.
Never commit broken builds or failing tests.

---

## Testing philosophy

- Every internal package has a `_test.go`
- Golden files in `testdata/` for deterministic output tests
- Fuzz targets in scanner and git parsers
- Benchmarks alongside functional tests
- Integration tests use real git bundles in `testdata/repos/`

---

## Environment variables honored at runtime

| Variable | Effect |
|----------|--------|
| `CODEYE_CACHE_DIR` | Override cache directory |
| `CODEYE_WORKERS` | Override goroutine pool size |
| `CODEYE_NO_COLOR` | Disable color output |
| `NO_COLOR` | Standard — disables color |
| `CODEYE_FORMAT` | Default output format |
| `CODEYE_CONFIG` | Path to .codeye.toml |

---

## Release checklist (agent must verify before tagging)

- [ ] `make test` passes with `-race` flag
- [ ] `make bench` passes performance targets
- [ ] `make lint` produces zero warnings
- [ ] `make snapshot` builds all platform binaries
- [ ] `codeye --version` prints correct semver + commit hash
- [ ] `codeye doctor` exits 0 on clean environment
- [ ] README.md has accurate install instructions
- [ ] CHANGELOG.md updated
