# Contributing to codeye

Thanks for contributing.

## Development setup

Requirements:

- Go 1.22+
- Git 2.x
- Hugo 0.156+ if you are changing the website

Clone the repo and run:

```sh
make test
make build
```

For website work:

```sh
make web
```

## Change expectations

- Keep changes focused. Split unrelated fixes into separate pull requests.
- Add or update tests when behavior changes.
- Update the docs when flags, config, install paths, or output formats change.
- Prefer preserving backwards compatibility for CLI flags and config keys unless there is a strong reason not to.

## Coding notes

- `codeye` is a pure Go CLI. Avoid introducing CGO or background services.
- Keep the fast path fast. Watch for unnecessary allocations, extra subprocesses, or cache invalidation mistakes.
- Favor simple subprocess-based git integration over adding new heavy dependencies.

## Before opening a pull request

Run as much of this as applies:

```sh
go test ./...
go vet ./...
hugo --minify --source web
```

Include in the PR description:

- what changed
- why it changed
- how you validated it
- any compatibility or release impact

## Reporting bugs

Please include:

- operating system and architecture
- `codeye version`
- `git --version`
- the command you ran
- the expected result
- the actual result

If the bug involves a specific repository layout, attach a minimal reproduction when possible.
