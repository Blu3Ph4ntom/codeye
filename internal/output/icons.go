package output

import "strings"

// FileIcon resolves a Nerd Font v3 icon glyph for a file extension or exact filename.
//
// Codepoints are sourced from the same icon set used by eza, lsd, and yazi —
// which in turn mirrors what VS Code's "Material Icon Theme" and "vscode-icons"
// extensions display in the file explorer.
//
// The priority chain is:
//   1. Exact filename match (case-insensitive): Makefile, Dockerfile, go.mod, …
//   2. Extension match (case-insensitive): .go, .py, .rs, …
//   3. Fallback generic file icon
//
// Reference icon sources:
//   eza:  https://github.com/eza-community/eza/blob/main/src/output/icons.rs
//   lsd:  https://github.com/lsd-rs/lsd/blob/develop/src/icon.rs
//   yazi: https://github.com/sxyazi/yazi/blob/main/yazi-config/preset/theme.toml

// NF holds all Nerd Font v3 glyph strings used for file icons.
// Names follow the nerd-fonts cheat-sheet naming convention.
var NF = struct {
	// Languages
	Go, Rust, Python, Ruby, PHP, Perl, Lua,
	Java, Kotlin, Scala, Groovy, Clojure,
	JavaScript, TypeScript, JSX, TSX, CoffeeScript,
	Swift, Dart, ObjectiveC,
	C, CPP, CSharp, Zig,
	Haskell, Elixir, Erlang, OCaml, FSharp,
	Crystal, Nim, Julia, R, Vlang, Gleam,
	// Web
	HTML, CSS, SCSS, Sass, Vue, Svelte, Angular,
	// Shell
	Shell, Fish, PowerShell,
	// Data / Config
	JSON, YAML, TOML, XML, CSV, NDJSON, HCL,
	// Docs
	Markdown, Rst, LaTeX, Asciidoc,
	// DB
	SQL, SQLite,
	// Container / DevOps
	Docker, Kubernetes, Terraform, Ansible,
	// Build
	Makefile, Cmake, Gradle, Bazel, Ninja,
	// VCS
	Git, GitIgnore,
	// Lock files / manifests
	Package, Cargo, Gemfile, Pipfile, NpmLock, PnpmLock,
	// Config / env
	Env, EditorConfig, PrettierRC,
	// Binary / archives
	Archive, Binary,
	// Text / generic
	Text, Cert, License,
	// Generic
	File, FilledFile string
}{
	// ── Languages ────────────────────────────────────────────────────────
	Go:           "\ue627", // nf-dev-go
	Rust:         "\ue7a8", // nf-seti-rust
	Python:       "\ue73c", // nf-dev-python
	Ruby:         "\ue739", // nf-dev-ruby
	PHP:          "\ue73d", // nf-dev-php
	Perl:         "\ue769", // nf-dev-perl
	Lua:          "\ue620", // nf-seti-lua
	Java:         "\ue738", // nf-dev-java
	Kotlin:       "\ue70e", // nf-seti-kotlin
	Scala:        "\ue737", // nf-dev-scala
	Groovy:       "\ue775", // nf-dev-groovy
	Clojure:      "\ue76a", // nf-dev-clojure
	JavaScript:   "\ue74e", // nf-dev-javascript
	TypeScript:   "\ue628", // nf-seti-typescript
	JSX:          "\ue7ba", // nf-dev-react
	TSX:          "\ue7ba", // nf-dev-react (TSX = React)
	CoffeeScript: "\ue751", // nf-dev-coffeescript
	Swift:        "\ue755", // nf-dev-swift
	Dart:         "\ue798", // nf-dev-dart
	ObjectiveC:   "\ue61e", // nf-custom-c (ObjC inherits)
	C:            "\ue61e", // nf-custom-c
	CPP:          "\ue61d", // nf-custom-cpp
	CSharp:       "\uf031b", // nf-md-language_csharp
	Zig:          "\ue6a9", // nf-seti-zig
	Haskell:      "\ue61f", // nf-dev-haskell
	Elixir:       "\ue62d", // nf-dev-elixir
	Erlang:       "\ue7b1", // nf-dev-erlang
	OCaml:        "\ue67a", // nf-seti-ocaml
	FSharp:       "\ue7a7", // nf-dev-fsharp
	Crystal:      "\ue7a3", // nf-custom-crystal
	Nim:          "\ue677", // nf-seti-nim
	Julia:        "\ue624", // nf-seti-julia
	R:            "\uf25d", // nf-fa-registered (R logo approximation)
	Vlang:        "\ue6b1", // nf-seti-v
	Gleam:        "\ue6a9", // placeholder: seti-zig shape

	// ── Web ──────────────────────────────────────────────────────────────
	HTML:    "\ue736", // nf-dev-html5
	CSS:     "\ue749", // nf-dev-css3
	SCSS:    "\ue603", // nf-seti-sass
	Sass:    "\ue74b", // nf-dev-sass
	Vue:     "\ue6a0", // nf-dev-vue
	Svelte:  "\ue697", // nf-dev-svelte
	Angular: "\ue753", // nf-dev-angular

	// ── Shell ────────────────────────────────────────────────────────────
	Shell:      "\ue691", // nf-dev-terminal
	Fish:       "\ue691", // nf-dev-terminal
	PowerShell: "\uebc7", // nf-md-powershell

	// ── Data / Config ────────────────────────────────────────────────────
	JSON:   "\ue60b", // nf-seti-json
	YAML:   "\ue6d2", // nf-seti-yaml
	TOML:   "\ue6b2", // nf-seti-config
	XML:    "\ue619", // nf-seti-xml
	CSV:    "\uf1c3", // nf-fa-file_excel_o
	NDJSON: "\ue60b", // nf-seti-json
	HCL:    "\ue61b", // nf-seti-terraform

	// ── Docs ─────────────────────────────────────────────────────────────
	Markdown: "\ue73e", // nf-dev-markdown
	Rst:      "\uf15c", // nf-fa-file_text_o
	LaTeX:    "\ue612", // nf-seti-latex
	Asciidoc: "\uf15c", // nf-fa-file_text_o

	// ── DB ───────────────────────────────────────────────────────────────
	SQL:    "\uf1c0", // nf-fa-database
	SQLite: "\ue7c4", // nf-dev-sqlite

	// ── Container / DevOps ───────────────────────────────────────────────
	Docker:     "\uf308", // nf-dev-docker
	Kubernetes: "\ue7b2", // nf-dev-kubernetes (helm/k8s)
	Terraform:  "\ue61b", // nf-seti-terraform
	Ansible:    "\ue61b", // nf-seti-config

	// ── Build ────────────────────────────────────────────────────────────
	Makefile: "\ue779", // nf-dev-cmake
	Cmake:    "\ue779", // nf-dev-cmake
	Gradle:   "\ue660", // nf-dev-gradle
	Bazel:    "\ue63a", // nf-seti-bazel
	Ninja:    "\ue779", // nf-dev-cmake (no dedicated ninja icon)

	// ── VCS ──────────────────────────────────────────────────────────────
	Git:       "\ue725", // nf-dev-git
	GitIgnore: "\ue725", // nf-dev-git

	// ── Manifests / Lock files ────────────────────────────────────────────
	Package: "\ue6b4", // nf-seti-npm / package
	Cargo:   "\ue7a8", // nf-seti-rust
	Gemfile: "\ue739", // nf-dev-ruby
	Pipfile: "\ue73c", // nf-dev-python
	NpmLock: "\ue6b4", // nf-seti-npm
	PnpmLock: "\ue6b4", // nf-seti-npm

	// ── Config / env ─────────────────────────────────────────────────────
	Env:          "\uf462", // nf-oct-key
	EditorConfig: "\ue614", // nf-seti-editorconfig
	PrettierRC:   "\ue6b3", // nf-seti-prettier

	// ── Binary / archives ────────────────────────────────────────────────
	Archive: "\uf1c6", // nf-fa-file_archive_o
	Binary:  "\uf471", // nf-oct-file_binary

	// ── Text / generic ───────────────────────────────────────────────────
	Text:    "\uf15c", // nf-fa-file_text_o
	Cert:    "\uf0a3", // nf-fa-certificate
	License: "\uf0a3", // nf-fa-certificate

	// ── Generic fallback ─────────────────────────────────────────────────
	File:       "\uf15b", // nf-fa-file_o
	FilledFile: "\uf15c", // nf-fa-file_text_o
}

// extIcons maps lowercase file extensions (with dot) to Nerd Font glyphs.
// Mirrors VS Code Material Icon Theme / vscode-icons coverage.
var extIcons = map[string]string{
	// Go
	".go":   NF.Go,
	// Rust
	".rs":   NF.Rust,
	// C family
	".c":    NF.C,   ".h":   NF.C,
	".cc":   NF.CPP, ".cpp": NF.CPP, ".cxx": NF.CPP,
	".hh":   NF.CPP, ".hpp": NF.CPP, ".hxx": NF.CPP,
	// C#
	".cs":   NF.CSharp, ".csx": NF.CSharp,
	// D
	".d":    NF.File,
	// Zig
	".zig":  NF.Zig,
	// Python
	".py":   NF.Python, ".pyw": NF.Python, ".pyi": NF.Python,
	// Ruby
	".rb":      NF.Ruby, ".rake": NF.Ruby, ".gemspec": NF.Ruby,
	".rbw":     NF.Ruby,
	// PHP
	".php":  NF.PHP, ".phtml": NF.PHP, ".php4": NF.PHP, ".php5": NF.PHP,
	// Perl
	".pl":   NF.Perl, ".pm": NF.Perl, ".pod": NF.Perl,
	// Lua
	".lua":  NF.Lua,
	// Java
	".java": NF.Java, ".class": NF.Java, ".jar": NF.Java,
	// Kotlin
	".kt":   NF.Kotlin, ".kts": NF.Kotlin,
	// Scala
	".scala": NF.Scala, ".sc": NF.Scala,
	// Groovy
	".groovy": NF.Groovy, ".gradle": NF.Gradle,
	// Clojure
	".clj":   NF.Clojure, ".cljs": NF.Clojure,
	".cljc":  NF.Clojure, ".edn":  NF.Clojure,
	// JavaScript
	".js":    NF.JavaScript, ".mjs": NF.JavaScript, ".cjs": NF.JavaScript,
	".jsx":   NF.JSX,
	// TypeScript
	".ts":    NF.TypeScript, ".mts": NF.TypeScript, ".cts": NF.TypeScript,
	".tsx":   NF.TSX,
	// CoffeeScript
	".coffee": NF.CoffeeScript,
	// Swift
	".swift": NF.Swift,
	// Objective-C
	".m":    NF.ObjectiveC, ".mm": NF.ObjectiveC,
	// Dart
	".dart": NF.Dart,
	// Elixir
	".ex":   NF.Elixir, ".exs": NF.Elixir, ".eex": NF.Elixir, ".leex": NF.Elixir, ".heex": NF.Elixir,
	// Erlang
	".erl":  NF.Erlang, ".hrl": NF.Erlang,
	// Haskell
	".hs":   NF.Haskell, ".lhs": NF.Haskell,
	// F#
	".fs":   NF.FSharp, ".fsi": NF.FSharp, ".fsx": NF.FSharp,
	// OCaml
	".ml":   NF.OCaml, ".mli": NF.OCaml,
	// Crystal
	".cr":   NF.Crystal,
	// Nim
	".nim":  NF.Nim, ".nims": NF.Nim,
	// Julia
	".jl":   NF.Julia,
	// R
	".r":    NF.R,
	// Lua already above
	// V
	".v":    NF.Vlang, ".vsh": NF.Vlang,
	// Shell
	".sh":   NF.Shell, ".bash": NF.Shell, ".zsh": NF.Shell,
	".ksh":  NF.Shell, ".csh": NF.Shell, ".tcsh": NF.Shell,
	".fish": NF.Fish,
	".ps1":  NF.PowerShell, ".psm1": NF.PowerShell, ".psd1": NF.PowerShell,
	// Web
	".html": NF.HTML, ".htm": NF.HTML, ".xhtml": NF.HTML,
	".vue":  NF.Vue,
	".svelte": NF.Svelte,
	".css":  NF.CSS,
	".scss": NF.SCSS,
	".sass": NF.Sass,
	".less": NF.CSS,
	// Data / Config
	".json":     NF.JSON, ".jsonc": NF.JSON, ".json5": NF.JSON,
	".yaml":     NF.YAML, ".yml": NF.YAML,
	".toml":     NF.TOML,
	".xml":      NF.XML, ".xsl": NF.XML, ".xslt": NF.XML,
	".csv":      NF.CSV, ".tsv": NF.CSV,
	".ndjson":   NF.NDJSON,
	".hcl":      NF.HCL, ".tf": NF.Terraform, ".tfvars": NF.Terraform,
	".ini":      NF.TOML, ".cfg": NF.TOML, ".conf": NF.TOML,
	".properties": NF.TOML,
	".env":      NF.Env,
	// Docs
	".md":   NF.Markdown, ".mdx": NF.Markdown, ".markdown": NF.Markdown,
	".rst":  NF.Rst,
	".tex":  NF.LaTeX, ".sty": NF.LaTeX, ".cls": NF.LaTeX,
	".adoc": NF.Asciidoc, ".asciidoc": NF.Asciidoc,
	// DB
	".sql":    NF.SQL, ".mysql": NF.SQL, ".pgsql": NF.SQL,
	".sqlite": NF.SQLite, ".db": NF.SQLite,
	// Archives
	".zip":  NF.Archive, ".tar": NF.Archive, ".gz": NF.Archive,
	".bz2":  NF.Archive, ".xz":  NF.Archive, ".zst": NF.Archive,
	".7z":   NF.Archive, ".rar": NF.Archive,
	".tgz":  NF.Archive, ".tbz2": NF.Archive,
	// Certs
	".pem":  NF.Cert, ".crt": NF.Cert, ".cer": NF.Cert,
	".key":  NF.Cert, ".p12": NF.Cert, ".pfx": NF.Cert,
	// Text
	".txt":  NF.Text, ".text": NF.Text, ".log": NF.Text,
	".diff": NF.Text, ".patch": NF.Text,
	// Binary
	".o": NF.Binary, ".so": NF.Binary, ".a": NF.Binary,
	".dll": NF.Binary, ".exe": NF.Binary, ".dylib": NF.Binary,
	".wasm": NF.Binary,
	// Proto
	".proto": NF.File,
	// Nix
	".nix": NF.File,
	// Dockerfile variations
	".dockerfile": NF.Docker,
}

// nameIcons maps exact filenames (lowercased) to Nerd Font glyphs.
// Checked before extension lookup. Matches what VS Code icon themes do.
var nameIcons = map[string]string{
	// Docker
	"dockerfile":           NF.Docker,
	"docker-compose.yml":   NF.Docker,
	"docker-compose.yaml":  NF.Docker,
	".dockerignore":        NF.Docker,

	// Go
	"go.mod":  NF.Go,
	"go.sum":  NF.Go,
	"go.work": NF.Go,
	"go.work.sum": NF.Go,

	// Rust
	"cargo.toml":  NF.Cargo,
	"cargo.lock":  NF.Cargo,

	// Node / JS
	"package.json":      NF.Package,
	"package-lock.json": NF.NpmLock,
	"pnpm-lock.yaml":    NF.PnpmLock,
	"yarn.lock":         NF.Package,
	".npmrc":            NF.Package,
	".nvmrc":            NF.Package,
	".node-version":     NF.Package,
	"tsconfig.json":     NF.TypeScript,
	"jsconfig.json":     NF.JavaScript,
	"webpack.config.js": NF.JavaScript,
	"vite.config.ts":    NF.TypeScript,
	"vite.config.js":    NF.JavaScript,
	"rollup.config.js":  NF.JavaScript,
	"babel.config.js":   NF.JavaScript,
	".babelrc":          NF.JavaScript,
	"next.config.js":    NF.JavaScript,
	"next.config.ts":    NF.TypeScript,
	"nuxt.config.ts":    NF.Vue,

	// Python
	"requirements.txt":     NF.Python,
	"requirements-dev.txt": NF.Python,
	"pipfile":              NF.Pipfile,
	"pipfile.lock":         NF.Pipfile,
	"pyproject.toml":       NF.Python,
	"setup.py":             NF.Python,
	"setup.cfg":            NF.Python,
	"tox.ini":              NF.Python,

	// Ruby
	"gemfile":      NF.Gemfile,
	"gemfile.lock": NF.Gemfile,
	"rakefile":     NF.Ruby,
	".ruby-version": NF.Ruby,

	// PHP
	"composer.json": NF.PHP,
	"composer.lock": NF.PHP,

	// Make / Build
	"makefile":    NF.Makefile,
	"makefile.in": NF.Makefile,
	"gnumakefile": NF.Makefile,
	"cmakecache.txt": NF.Cmake,
	"cmakelists.txt": NF.Cmake,
	"build.gradle":  NF.Gradle,
	"build.gradle.kts": NF.Gradle,
	"settings.gradle": NF.Gradle,
	"build.bazel":   NF.Bazel,
	"workspace":     NF.Bazel,
	"build.ninja":   NF.Ninja,

	// Terraform
	"main.tf":   NF.Terraform,
	".terraform": NF.Terraform,

	// Git
	".gitignore":     NF.GitIgnore,
	".gitattributes": NF.Git,
	".gitmodules":    NF.Git,
	".gitconfig":     NF.Git,
	".mailmap":       NF.Git,

	// CI/CD
	".travis.yml":       NF.YAML,
	".github":           NF.Git,
	"jenkinsfile":       NF.File,
	".circleci":         NF.File,

	// Config / Editor
	".editorconfig": NF.EditorConfig,
	".prettierrc":   NF.PrettierRC,
	".prettierrc.js": NF.PrettierRC,
	".prettierrc.json": NF.PrettierRC,
	".eslintrc":     NF.JavaScript,
	".eslintrc.js":  NF.JavaScript,
	".eslintrc.json": NF.JavaScript,
	".eslintignore": NF.JavaScript,
	".stylelintrc":  NF.CSS,
	"tailwind.config.js": NF.CSS,
	"tailwind.config.ts": NF.CSS,
	".golangci.yml": NF.Go,
	".goreleaser.yml": NF.Go,
	".goreleaser.yaml": NF.Go,

	// Env
	".env":          NF.Env,
	".env.local":    NF.Env,
	".env.example":  NF.Env,
	".env.test":     NF.Env,
	".env.production": NF.Env,

	// Licenses / Readme
	"license":          NF.License,
	"license.md":       NF.License,
	"license.txt":      NF.License,
	"licence":          NF.License,
	"copying":          NF.License,
	"readme":           NF.Markdown,
	"readme.md":        NF.Markdown,
	"readme.rst":       NF.Markdown,
	"changelog":        NF.Markdown,
	"changelog.md":     NF.Markdown,
	"changelog.txt":    NF.Markdown,
	"authors":          NF.Text,
	"contributors":     NF.Text,
	"codeowners":       NF.Git,

	// Kubernetes
	"helmfile.yaml": NF.Kubernetes,

	// Proto
	"buf.yaml": NF.File,
	"buf.gen.yaml": NF.File,
}

// primaryExt maps language name → canonical extension for icon lookup.
// Used when only a language name is available (e.g., table renderer).
var langPrimaryExt = map[string]string{
	"Go":           ".go",
	"Rust":         ".rs",
	"Python":       ".py",
	"Ruby":         ".rb",
	"PHP":          ".php",
	"Perl":         ".pl",
	"Lua":          ".lua",
	"Java":         ".java",
	"Kotlin":       ".kt",
	"Scala":        ".scala",
	"Groovy":       ".groovy",
	"Clojure":      ".clj",
	"JavaScript":   ".js",
	"JSX":          ".jsx",
	"TypeScript":   ".ts",
	"TSX":          ".tsx",
	"CoffeeScript": ".coffee",
	"Swift":        ".swift",
	"Dart":         ".dart",
	"Objective-C":  ".m",
	"C":            ".c",
	"C++":          ".cpp",
	"C#":           ".cs",
	"Zig":          ".zig",
	"Haskell":      ".hs",
	"Elixir":       ".ex",
	"Erlang":       ".erl",
	"OCaml":        ".ml",
	"F#":           ".fs",
	"Crystal":      ".cr",
	"Nim":          ".nim",
	"Julia":        ".jl",
	"R":            ".r",
	"V":            ".v",
	"Gleam":        ".gleam",
	"HTML":         ".html",
	"CSS":          ".css",
	"SCSS":         ".scss",
	"Sass":         ".sass",
	"Vue":          ".vue",
	"Svelte":       ".svelte",
	"Shell":        ".sh",
	"Fish":         ".fish",
	"PowerShell":   ".ps1",
	"JSON":         ".json",
	"YAML":         ".yaml",
	"TOML":         ".toml",
	"XML":          ".xml",
	"CSV":          ".csv",
	"SQL":          ".sql",
	"SQLite":       ".sqlite",
	"Markdown":     ".md",
	"LaTeX":        ".tex",
	"Dockerfile":   "dockerfile",    // exact name → nameIcons
	"Makefile":     "makefile",      // exact name → nameIcons
	"Gitignore":    ".gitignore",    // exact name → nameIcons
	"Gitattributes": ".gitattributes",
	"Text":         ".txt",
}

// IconForExt returns the Nerd Font glyph for a raw filename or extension.
// filename should be the base name (e.g., "main.go", ".gitignore", "Makefile").
func IconForExt(filename string) string {
	lower := strings.ToLower(filename)

	// 1. Exact filename match
	if g, ok := nameIcons[lower]; ok {
		return g
	}

	// 2. Extension match
	if idx := strings.LastIndexByte(lower, '.'); idx >= 0 {
		ext := lower[idx:]
		if g, ok := extIcons[ext]; ok {
			return g
		}
	}

	// 3. Plain filename with no extension — try nameIcons again without dot
	// (already covered above)

	return NF.File
}

// IconForLang returns the Nerd Font glyph for a language name, using the
// canonical extension as an intermediate lookup for maximum accuracy.
func IconForLang(lang string) string {
	// First: if the language has a canonical extension, look it up by extension
	if ext, ok := langPrimaryExt[lang]; ok {
		return IconForExt(ext)
	}
	// If ext is an exact filename (no leading dot), look up nameIcons
	// already handled by IconForExt above
	return NF.File
}
