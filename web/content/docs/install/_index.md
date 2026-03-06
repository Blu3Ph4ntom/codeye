---
title: "Install"
description: "Install codeye on macOS, Linux, or Windows."
---

# Install

codeye ships as a single static binary — no runtime, no CGO, no dependencies.

## Go install

The quickest way if you have Go 1.22+:

```bash
go install github.com/codeye/codeye/cmd/codeye@latest
```

The binary lands in `$GOPATH/bin` (usually `~/go/bin`). Make sure that's on your `$PATH`.

## Pre-built binaries

Download from the [Releases page](https://github.com/codeye/codeye/releases):

| Platform | File |
|----------|------|
| Linux x86-64 | `codeye_linux_amd64.tar.gz` |
| Linux arm64  | `codeye_linux_arm64.tar.gz` |
| macOS x86-64 | `codeye_darwin_amd64.tar.gz` |
| macOS arm64  | `codeye_darwin_arm64.tar.gz` |
| Windows x86-64 | `codeye_windows_amd64.zip` |

Extract and put the binary somewhere on your `PATH`.

## Homebrew (macOS/Linux)

```bash
brew install codeye/tap/codeye
```

## Verify

```bash
codeye --version
# codeye v0.1.0 (abc1234)
```

## Requirements

- `git` must be available on `PATH` (any version ≥ 2.0)
- No other runtime dependencies

## Shell completions

```bash
# bash
codeye completion bash > /etc/bash_completion.d/codeye

# zsh
codeye completion zsh > "${fpath[1]}/_codeye"

# fish
codeye completion fish > ~/.config/fish/completions/codeye.fish
```
