---
title: "Install"
description: "Install codeye on Linux, macOS, or Windows."
---

# Install

`codeye` ships as a single static binary. You need `git` on your `PATH`; nothing else is required at runtime.

## Go install

Best for developers already using Go 1.22+:

```bash
go install github.com/blu3ph4ntom/codeye/cmd/codeye@latest
```

The binary is installed into your Go bin directory, usually `~/go/bin` on Unix-like systems.

## Hosted install scripts

The GitHub Pages site hosts the same installer scripts tracked in the repository:

```bash
# Linux / macOS
curl -sSfL https://codeye.bluephantom.dev/install.sh | sh
```

```powershell
# Windows PowerShell
iex (irm https://codeye.bluephantom.dev/install.ps1)
```

## Release archives

Prebuilt binaries are published on the [GitHub releases page](https://github.com/blu3ph4ntom/codeye/releases).

Current archive names follow this pattern:

| Platform | Example archive |
|----------|-----------------|
| Linux amd64 | `codeye_0.1.0_linux_amd64.tar.gz` |
| Linux arm64 | `codeye_0.1.0_linux_arm64.tar.gz` |
| macOS amd64 | `codeye_0.1.0_darwin_amd64.tar.gz` |
| macOS arm64 | `codeye_0.1.0_darwin_arm64.tar.gz` |
| Windows amd64 | `codeye_0.1.0_windows_amd64.zip` |

Extract the archive and move `codeye` or `codeye.exe` onto your `PATH`.

## Verify

```bash
codeye version
codeye doctor
```

`codeye doctor` checks for the git binary, the current repo context, and the cache directory.

## Shell completion

```bash
# bash
codeye completion bash > ~/.bash_completion.d/codeye

# zsh
codeye completion zsh > "${fpath[1]}/_codeye"

# fish
codeye completion fish > ~/.config/fish/completions/codeye.fish
```
