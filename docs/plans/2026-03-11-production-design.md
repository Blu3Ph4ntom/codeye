# codeye production-readiness design

Date: 2026-03-11

## Goal

Move `codeye` from a promising repo into a shippable product surface:

- make the documented CLI behavior match the binary
- ensure release metadata points at the real repository
- deploy the public website through GitHub Pages
- bind the site to `codeye.bluephantom.dev`

## Scope

### CLI

- activate layered config loading for `.codeye.toml`, `CODEYE_*`, and CLI flags
- preserve compatibility for older docs with alias flags where practical
- fix documented comma-separated `--lang` filters
- add tests for config resolution and language filtering

### Release pipeline

- keep GoReleaser focused on artifacts that can ship from this repository today
- remove package-manager publishing steps that depend on external repos or generated assets not present in the repo
- keep release output aligned to the real GitHub namespace

### Website

- deploy the Hugo site with a dedicated GitHub Pages workflow
- commit a `CNAME` file for the custom domain
- update metadata, copy, install paths, and docs so the public site matches the current binary

## Non-goals

- DNS provider changes outside the repository
- creating external distribution repos such as a Homebrew tap
- adding telemetry, auth, or a hosted backend
