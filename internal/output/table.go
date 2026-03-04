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

var langEmoji = map[string]string{
	"Go":         "🐹",
	"Rust":       "🦀",
	"Python":     "🐍",
	"JavaScript": "🟨",
	"TypeScript": "🔷",
	"Ruby":       "💎",
	"Java":       "☕",
	"Kotlin":     "🟣",
	"Swift":      "🧡",
	"C":          "🔵",
	"C++":        "🔵",
	"C#":         "🟢",
	"PHP":        "🐘",
	"Shell":      "🐚",
	"HTML":       "🌐",
	"CSS":        "🎨",
	"Markdown":   "📝",
	"SQL":        "🗄️",
	"Dart":       "🎯",
	"Scala":      "🔴",
	"Elixir":     "💜",
	"Haskell":    "🟤",
}

func emojiFor(lang string, use bool) string {
	if !use {
		return ""
	}
	if e, ok := langEmoji[lang]; ok {
		return e + " "
	}
	return "   "
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
		// Single-line summary
		fmt.Fprintf(w, "%s: %s lines across %d files (%d languages)\n",
			ref,
			humanize.Comma(result.Total.Lines),
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
		sep := strings.Repeat("─", maxLangLen+2+7+10+10+10+12+7)
		p.dim(sep)
		hdr := fmt.Sprintf(" %-*s  %6s  %9s", maxLangLen, "Language", "Files", "Code")
		if opts.Wide {
			hdr += fmt.Sprintf("  %8s  %8s", "Blank", "Comment")
		}
		hdr += fmt.Sprintf("  %9s", "Total")
		if opts.Pct {
			hdr += fmt.Sprintf("  %6s", "%")
		}
		p.bold(hdr)
		p.dim(sep)
	}

	// Data rows
	total := result.Total
	for _, l := range langs {
		pct := 0.0
		if total.Lines > 0 {
			pct = float64(l.Lines) / float64(total.Lines) * 100
		}
		name := emojiFor(l.Name, opts.Emoji) + l.Name
		row := fmt.Sprintf(" %-*s  %6s  %9s",
			maxLangLen, name,
			humanize.Comma(int64(l.Files)),
			humanize.Comma(l.Code),
		)
		if opts.Wide {
			row += fmt.Sprintf("  %8s  %8s",
				humanize.Comma(l.Blank),
				humanize.Comma(l.Comment),
			)
		}
		row += fmt.Sprintf("  %9s", humanize.Comma(l.Lines))
		if opts.Pct {
			row += fmt.Sprintf("  %5.1f%%", pct)
		}
		p.row(l.Name, row)
	}

	// Total row
	if !opts.NoHeader {
		sep := strings.Repeat("─", maxLangLen+2+7+10+10+10+12+7)
		p.dim(sep)
		totRow := fmt.Sprintf(" %-*s  %6s  %9s",
			maxLangLen, "Total",
			humanize.Comma(int64(total.Files)),
			humanize.Comma(total.Code),
		)
		if opts.Wide {
			totRow += fmt.Sprintf("  %8s  %8s",
				humanize.Comma(total.Blank),
				humanize.Comma(total.Comment),
			)
		}
		totRow += fmt.Sprintf("  %9s", humanize.Comma(total.Lines))
		if opts.Pct {
			totRow += fmt.Sprintf("  %5.1f%%", 100.0)
		}
		p.bold(totRow)
		p.dim(sep)
	}

	// Footer
	cacheTag := "cache miss"
	if result.Cached {
		cacheTag = "cached"
	}
	footer := fmt.Sprintf(" ⚡ %dms · %s", result.ScanMs, cacheTag)
	if result.TreeSHA != "" {
		footer += fmt.Sprintf(" · tree %.8s", result.TreeSHA)
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
	w      io.Writer
	p      termenv.Profile
	opts   RenderOpts
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

func (p *printer) styled(s string, style termenv.Style) {
	fmt.Fprintln(p.w, style.Styled(s))
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
	// Cycle through palette colors per language
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
