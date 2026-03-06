package output

import (
"fmt"
"io"
"strings"

"github.com/blu3ph4ntom/codeye/internal/scanner"
"github.com/dustin/go-humanize"
"github.com/muesli/termenv"
)

// TableRenderer renders scan results as a formatted table.
type TableRenderer struct{}

// langBrandColors maps language names to their GitHub Linguist / official brand hex colors.
// These are rendered via ANSI true-color — zero setup, works in every modern terminal.
var langBrandColors = map[string]string{
	// ── Systems ──────────────────────────────────────────
	"Go":            "#00ADD8", // Go gopher blue
	"Rust":          "#DEA584", // Rust official orange
	"C":             "#A8B9CC", // C logo blue-grey
	"C++":           "#F34B7D", // Linguist pink
	"C#":            "#178600", // Linguist green
	"Zig":           "#EC915C", // Zig orange
	"D":             "#BA595E",
	"Nim":           "#FFE953",
	"Assembly":      "#6E4C13",
	// ── Scripted ─────────────────────────────────────────
	"Python":        "#3572A5", // Python blue
	"Ruby":          "#CC342D", // Ruby red
	"PHP":           "#4F5D95", // PHP indigo
	"Perl":          "#0298C3",
	"Lua":           "#4080D0", // lightened from Linguist navy
	"R":             "#198CE7",
	"Julia":         "#9558B2",
	// ── JVM ──────────────────────────────────────────────
	"Java":          "#B07219", // Java orange-brown
	"Kotlin":        "#A97BFF", // Kotlin purple
	"Scala":         "#C22D40", // Scala red
	"Groovy":        "#E69F56",
	"Clojure":       "#DB5855",
	// ── JS family ────────────────────────────────────────
	"JavaScript":    "#F7DF1E", // JS yellow
	"TypeScript":    "#3178C6", // TS blue
	"CoffeeScript":  "#244776",
	// ── Mobile ───────────────────────────────────────────
	"Swift":         "#F05138", // Swift orange-red
	"Dart":          "#00B4AB",
	"Objective-C":   "#438EFF",
	// ── Functional ───────────────────────────────────────
	"Haskell":       "#5E5086",
	"Elixir":        "#6E4A7E",
	"Erlang":        "#B83998",
	"OCaml":         "#3BE133",
	"F#":            "#B845FC",
	"Elm":           "#60B5CC",
	// ── Web ──────────────────────────────────────────────
	"HTML":          "#E34C26", // HTML orange
	"CSS":           "#264DE4", // CSS blue
	"SCSS":          "#C6538C",
	"Sass":          "#A53B70",
	"Vue":           "#41B883",
	// ── Shell ────────────────────────────────────────────
	"Shell":         "#89E051",
	// ── Data / Config ────────────────────────────────────
	"JSON":          "#A8A8A8",
	"YAML":          "#F0C44C",
	"TOML":          "#C6733A",
	"XML":           "#0060AC",
	"CSV":           "#A8A8A8",
	// ── Docs ─────────────────────────────────────────────
	"Markdown":      "#5BA4CF",
	"Text":          "#A8A8A8",
	"LaTeX":         "#3D8B37",
	// ── DevOps / Build ───────────────────────────────────
	"Dockerfile":    "#2496ED", // Docker blue
	"Makefile":      "#6D9B3A",
	"SQL":           "#E38C00",
	"HCL":           "#844FBA",
	"Terraform":     "#7B42BC",
	"Protobuf":      "#4285F4",
	// ── Misc ─────────────────────────────────────────────
	"Vim Script":    "#199F4B",
	"Gitignore":     "#F54D27",
	"Gitattributes": "#F54D27",
	"Unknown":       "#6B7280",
}

// brandColorFor returns the brand hex color for a language, with a neutral fallback.
func brandColorFor(lang string) string {
	if c, ok := langBrandColors[lang]; ok {
		return c
	}
	lower := strings.ToLower(lang)
	for k, v := range langBrandColors {
		if strings.ToLower(k) == lower {
			return v
		}
	}
	return "#6B7280"
}

// nfIconFor returns the Nerd Font icon prefix when --nf is active; empty otherwise.
func nfIconFor(lang string, nerdFont bool) string {
	if !nerdFont {
		return ""
	}
	return IconForLang(lang) + " "
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
if opts.NerdFont {
n += 3 // icon glyph (2 wide) + space
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
// Build plain-text name (with optional NF icon) then pad — ANSI codes must NOT
// be inside the Sprintf format string or byte-width padding breaks alignment.
plainName := nfIconFor(l.Name, opts.NerdFont) + l.Name
paddedName := fmt.Sprintf("%-*s", maxLangLen, plainName)
// Apply brand color to the padded name so numbers stay left-aligned.
coloredName := termenv.String(paddedName).Foreground(p.p.Color(brandColorFor(l.Name))).Bold().String()
numPart := fmt.Sprintf("  %6s  %10s",
humanize.Comma(int64(l.Files)),
humanize.Comma(l.Code),
)
numPart += fmt.Sprintf("  %10s  %10s",
humanize.Comma(l.Comment),
humanize.Comma(l.Blank),
)
numPart += fmt.Sprintf("  %10s", humanize.Comma(lSum))
if opts.Pct {
numPart += fmt.Sprintf("  %5.1f%%", pct)
}
p.row(l.Name, " "+coloredName+numPart)
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

func (p *printer) row(lang, s string) {
	color := brandColorFor(lang)
	prefix := termenv.String("▌").Foreground(p.p.Color(color))
	fmt.Fprintln(p.w, prefix.String()+s)
}
