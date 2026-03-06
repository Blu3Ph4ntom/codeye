package output

import (
"fmt"
"io"
"strings"

"github.com/codeye/codeye/internal/scanner"
"github.com/dustin/go-humanize"
"github.com/muesli/termenv"
)

// TableRenderer renders scan results as a formatted table.
type TableRenderer struct{}

// langNerdIcons maps language names to Nerd Font v3 glyph + trailing space.
//
// Codepoints come from the Devicons (nf-dev-*), seti-ui (nf-seti-*),
// and Material Design (nf-md-*) subsets bundled in Nerd Fonts v3.
// These are the same icons VS Code extensions like "vscode-icons" and
// "Material Icon Theme" display for each file type.
// Enable with --nf flag or CODEYE_NERD_FONTS=1 env var.
var langNerdIcons = map[string]string{
	// ── Systems ──────────────────────────────────────────
	"Go":            "\ue627 ", // nf-dev-go
	"Rust":          "\ue7a8 ", // nf-seti-rust
	"C":             "\ue61e ", // nf-custom-c
	"C++":           "\ue61d ", // nf-custom-cpp
	"C#":            "\uf031b ", // nf-md-language_csharp
	"Zig":           "\ue6a9 ", // nf-seti-zig
	// ── Scripted ─────────────────────────────────────────
	"Python":        "\ue73c ", // nf-dev-python
	"Ruby":          "\ue739 ", // nf-dev-ruby
	"PHP":           "\ue73d ", // nf-dev-php
	"Perl":          "\ue769 ", // nf-dev-perl
	"Lua":           "\ue620 ", // nf-seti-lua
	// ── JVM ──────────────────────────────────────────────
	"Java":          "\ue738 ", // nf-dev-java
	"Kotlin":        "\ue70e ", // nf-seti-kotlin
	"Scala":         "\ue737 ", // nf-dev-scala
	"Groovy":        "\ue775 ", // nf-dev-groovy
	"Clojure":       "\ue76a ", // nf-dev-clojure
	// ── Typed JS family ──────────────────────────────────
	"JavaScript":    "\ue74e ", // nf-dev-javascript
	"TypeScript":    "\ue628 ", // nf-seti-typescript
	"CoffeeScript":  "\ue751 ", // nf-dev-coffeescript
	// ── Mobile ───────────────────────────────────────────
	"Swift":         "\ue755 ", // nf-dev-swift
	"Dart":          "\ue798 ", // nf-dev-dart
	"Objective-C":   "\ue61e ", // nf-custom-c (ObjC uses C icon)
	// ── Functional ───────────────────────────────────────
	"Haskell":       "\ue61f ", // nf-dev-haskell
	"Elixir":        "\ue62d ", // nf-dev-elixir
	"Erlang":        "\ue7b1 ", // nf-dev-erlang
	"OCaml":         "\ue67a ", // nf-seti-ocaml
	"F#":            "\ue7a7 ", // nf-dev-fsharp
	// ── Web ──────────────────────────────────────────────
	"HTML":          "\ue736 ", // nf-dev-html5
	"CSS":           "\ue749 ", // nf-dev-css3
	"SCSS":          "\ue749 ", // nf-dev-css3
	"Sass":          "\ue74b ", // nf-dev-sass
	"Vue":           "\ue6a0 ", // nf-dev-vue
	// ── Shell ────────────────────────────────────────────
	"Shell":         "\ue691 ", // nf-dev-terminal (bash)
	// ── Data / Config ────────────────────────────────────
	"JSON":          "\ue60b ", // nf-seti-json
	"YAML":          "\ue6d2 ", // nf-seti-yaml
	"TOML":          "\ue6b2 ", // nf-seti-config (gear)
	"XML":           "\ue619 ", // nf-seti-xml
	"CSV":           "\uf1c3 ", // nf-fa-file_excel_o
	// ── Docs ─────────────────────────────────────────────
	"Markdown":      "\ue73e ", // nf-dev-markdown
	"Text":          "\uf15c ", // nf-fa-file_text_o
	"LaTeX":         "\ue612 ", // nf-seti-latex
	// ── DevOps / Build ───────────────────────────────────
	"Dockerfile":    "\uf308 ", // nf-dev-docker
	"Makefile":      "\ue779 ", // nf-dev-cmake
	"SQL":           "\uf1c0 ", // nf-fa-database
	// ── Git ──────────────────────────────────────────────
	"Gitignore":     "\ue725 ", // nf-dev-git
	"Gitattributes": "\ue725 ", // nf-dev-git
	// ── Fallback ─────────────────────────────────────────
	"Unknown":       "\uf15b ", // nf-fa-file
}

// langEmoji maps language names to emoji icons.
// These are plain Unicode emoji — no special fonts required.
// Chosen to have a real semantic connection to the language or ecosystem.
var langEmoji = map[string]string{
	// ── Systems ──────────────────────────────────────────
	"Go":            "🐹 ", // Gopher — Go's official mascot
	"Rust":          "🦀 ", // Ferris — Rust's mascot
	"C":             "⚙️  ", // gear — systems-level
	"C++":           "⚙️  ", // gear — systems-level
	"C#":            "🎵 ", // C# is literally a musical note
	"Zig":           "⚡ ", // lightning — Zig's speed focus
	// ── Scripted ─────────────────────────────────────────
	"Python":        "🐍 ", // snake — Python's mascot
	"Ruby":          "💎 ", // ruby gem
	"PHP":           "🐘 ", // Ellie the elephant — PHP's mascot
	"Perl":          "🧅 ", // onion — Perl's logo shape
	"Lua":           "🌙 ", // moon — Lua means "moon" in Portuguese
	// ── JVM ──────────────────────────────────────────────
	"Java":          "☕ ", // coffee — Java's origin story
	"Kotlin":        "🎯 ", // target — Kotlin island shape
	"Scala":         "🔴 ", // Scala's signature red
	"Groovy":        "🎸 ", // groovy
	"Clojure":       "🔵 ", // Clojure logo is a blue circle
	// ── JS family ────────────────────────────────────────
	"JavaScript":    "🟡 ", // JS logo is yellow
	"TypeScript":    "📘 ", // TypeScript docs / blue branding
	"CoffeeScript":  "☕ ", // CoffeeScript — coffee
	// ── Mobile ───────────────────────────────────────────
	"Swift":         "🦅 ", // swift bird
	"Dart":          "🎯 ", // dartboard
	"Objective-C":   "⚙️  ", // systems / legacy
	// ── Functional ───────────────────────────────────────
	"Haskell":       "λ  ", // lambda — Haskell is pure lambda calculus
	"Elixir":        "💜 ", // purple — Elixir/Phoenix branding
	"Erlang":        "📡 ", // telecoms origins
	"OCaml":         "🐫 ", // camel — OCaml's logo
	"F#":            "🎵 ", // F# is literally a musical note
	// ── Web ──────────────────────────────────────────────
	"HTML":          "🌐 ", // globe — web
	"CSS":           "🎨 ", // palette — styling
	"SCSS":          "🎨 ",
	"Sass":          "🎨 ",
	"Vue":           "💚 ", // Vue's green branding
	// ── Shell ────────────────────────────────────────────
	"Shell":         "🐚 ", // shell — it is literally a shell
	// ── Data / Config ────────────────────────────────────
	"JSON":          "{ } ", // JSON braces
	"YAML":          "📄 ", // config file
	"TOML":          "📄 ", // config file
	"XML":           "📄 ", // markup file
	"CSV":           "📊 ", // spreadsheet/chart
	// ── Docs ─────────────────────────────────────────────
	"Markdown":      "📝 ", // notepad — writing
	"Text":          "📄 ", // document
	"LaTeX":         "📐 ", // math / typesetting
	// ── DevOps / Build ───────────────────────────────────
	"Dockerfile":    "🐳 ", // Moby the whale — Docker's mascot
	"Makefile":      "🔧 ", // wrench — build tool
	"SQL":           "🗄️  ", // file cabinet — database
	// ── Git ──────────────────────────────────────────────
	"Gitignore":     "🚫 ", // no sign — ignoring files
	"Gitattributes": "⚙️  ", // settings gear
	// ── Fallback ─────────────────────────────────────────
	"Unknown":       "📁 ", // generic folder/file
}

// iconFor returns the icon prefix for a language given the render mode.
// If nerdFont=true, uses Nerd Font glyphs; otherwise emoji; if !use, empty string.
func iconFor(lang string, use, nerdFont bool) string {
	if !use {
		return ""
	}
	iconMap := langEmoji
	if nerdFont {
		iconMap = langNerdIcons
	}
	if icon, ok := iconMap[lang]; ok {
		return icon
	}
	// case-insensitive fallback
	lower := strings.ToLower(lang)
	for k, v := range iconMap {
		if strings.ToLower(k) == lower {
			return v
		}
	}
	if nerdFont {
		return "\uf15b " // fa-file generic
	}
	return "❓ "
}

// Render writes the table to w.
func (t *TableRenderer) Render(w io.Writer, result *scanner.ScanResult, opts RenderOpts) error {
p := newPrinter(w, opts)

langs := filteredLangs(result, opts)
if len(langs) == 0 {
fmt.Fprintln(w, "no files found")
return nil
}

ref := result.Ref
if ref == "" {
ref = "HEAD"
}

// Total lines (sum of all breakdown)
totalSum := result.Total.Total()

// Header
if !opts.Compact {
cacheStr := ""
if result.Cached {
cacheStr = " (cached)"
}
scanInfo := fmt.Sprintf("codeye · %s · %d files · %dms%s",
ref, result.Files, result.ScanMs, cacheStr)
p.header(scanInfo)
}

if opts.Compact {
fmt.Fprintf(w, "%s: %s lines across %d files (%d languages)\n",
ref,
humanize.Comma(totalSum),
result.Files,
len(langs),
)
return nil
}

// Column widths
maxLangLen := 8 // "Language"
for _, l := range langs {
n := len(l.Name)
if opts.Emoji {
n += 3
}
if n > maxLangLen {
maxLangLen = n
}
}

// Table header
if !opts.NoHeader {
hdr := fmt.Sprintf(" %-*s  %6s  %10s", maxLangLen, "Language", "Files", "Code")
hdr += fmt.Sprintf("  %10s  %10s", "Comments", "Blanks")
hdr += fmt.Sprintf("  %10s", "Total")
if opts.Pct {
hdr += fmt.Sprintf("  %6s", "%")
}

width := len(hdr) + 2
sep := strings.Repeat("─", width)
p.dim(sep)
p.bold(hdr)
p.dim(sep)
}

// Data rows
for _, l := range langs {
pct := 0.0
lSum := l.Total()
if totalSum > 0 {
pct = float64(lSum) / float64(totalSum) * 100
}
name := iconFor(l.Name, opts.Emoji, opts.NerdFont) + l.Name
row := fmt.Sprintf(" %-*s  %6s  %10s",
maxLangLen, name,
humanize.Comma(int64(l.Files)),
humanize.Comma(l.Code),
)
row += fmt.Sprintf("  %10s  %10s",
humanize.Comma(l.Comment),
humanize.Comma(l.Blank),
)
row += fmt.Sprintf("  %10s", humanize.Comma(lSum))
if opts.Pct {
row += fmt.Sprintf("  %5.1f%%", pct)
}
p.row(l.Name, row)
}

// Total row
if !opts.NoHeader {
// Calculate exact separator width
dummyHdr := fmt.Sprintf(" %-*s  %6s  %10s", maxLangLen, "Language", "Files", "Code")
dummyHdr += fmt.Sprintf("  %10s  %10s", "Comments", "Blanks")
dummyHdr += fmt.Sprintf("  %10s", "Total")
if opts.Pct { dummyHdr += fmt.Sprintf("  %6s", "%") }
sep := strings.Repeat("─", len(dummyHdr)+2)

p.dim(sep)
totRow := fmt.Sprintf(" %-*s  %6s  %10s",
maxLangLen, "Total",
humanize.Comma(int64(result.Total.Files)),
humanize.Comma(result.Total.Code),
)
totRow += fmt.Sprintf("  %10s  %10s",
humanize.Comma(result.Total.Comment),
humanize.Comma(result.Total.Blank),
)
totRow += fmt.Sprintf("  %10s", humanize.Comma(totalSum))
if opts.Pct {
totRow += fmt.Sprintf("  %5.1f%%", 100.0)
}
p.bold(totRow)
p.dim(sep)
}

// Footer
cacheTag := "⚡ cache hit"
if !result.Cached {
cacheTag = "❄️ cache miss"
}
footer := fmt.Sprintf(" %dms · %s", result.ScanMs, cacheTag)
if result.TreeSHA != "" {
footer += fmt.Sprintf(" · 🌳 %.8s", result.TreeSHA)
}
p.dim(footer)

return nil
}

// filteredLangs returns langs respecting the Top option.
func filteredLangs(result *scanner.ScanResult, opts RenderOpts) []scanner.LangStats {
langs := result.Langs
if opts.Top > 0 && len(langs) > opts.Top {
langs = langs[:opts.Top]
}
return langs
}

// printer abstracts color output.
type printer struct {
w    io.Writer
p    termenv.Profile
opts RenderOpts
}

func newPrinter(w io.Writer, opts RenderOpts) *printer {
prof := termenv.ColorProfile()
if opts.NoColor {
prof = termenv.Ascii
}
return &printer{w: w, p: prof, opts: opts}
}

func (p *printer) write(s string) {
fmt.Fprintln(p.w, s)
}

func (p *printer) header(s string) {
style := termenv.String(s).Bold().Foreground(p.p.Color("#7C3AED"))
fmt.Fprintln(p.w, style)
}

func (p *printer) dim(s string) {
style := termenv.String(s).Faint()
fmt.Fprintln(p.w, style)
}

func (p *printer) bold(s string) {
style := termenv.String(s).Bold()
fmt.Fprintln(p.w, style)
}

var langColors = []string{
"#7C3AED", "#2563EB", "#059669", "#D97706", "#DC2626",
"#0891B2", "#7C3AED", "#BE185D", "#865DFF", "#16A34A",
}

func (p *printer) row(lang, s string) {
h := fnv32(lang)
color := langColors[h%uint32(len(langColors))]
prefix := termenv.String("▌").Foreground(p.p.Color(color))
fmt.Fprintln(p.w, prefix.String()+s)
}

func fnv32(s string) uint32 {
var h uint32 = 2166136261
for i := 0; i < len(s); i++ {
h ^= uint32(s[i])
h *= 16777619
}
return h
}
