// Package output provides renderers for scan results.
package output

import (
	"io"

	"github.com/codeye/codeye/internal/scanner"
)

// RenderOpts carries display preferences to renderers.
type RenderOpts struct {
	NoColor   bool
	NoHeader  bool
	Wide      bool
	Compact   bool
	Pct       bool
	Emoji     bool
	NerdFont  bool // use Nerd Font glyphs instead of emoji (requires Nerd Font in terminal)
	Theme     string
	Top       int
	Sort      string
	Desc      bool
}

// Renderer is the interface all output formats implement.
type Renderer interface {
	Render(w io.Writer, result *scanner.ScanResult, opts RenderOpts) error
}

// Get returns a renderer for the given format name.
// Falls back to TableRenderer for unknown formats.
func Get(format string) Renderer {
	switch format {
	case "bar":
		return &BarRenderer{}
	case "spark", "sparkline":
		return &SparkRenderer{}
	case "json":
		return &JSONRenderer{}
	case "ndjson":
		return &NDJSONRenderer{}
	case "csv":
		return &CSVRenderer{}
	case "badge":
		return &BadgeRenderer{}
	case "markdown":
		return &MarkdownRenderer{}
	case "compact":
		return &CompactRenderer{}
	default:
		return &TableRenderer{}
	}
}

// ValidFormats lists all valid --format values.
var ValidFormats = []string{
	"table", "bar", "spark", "json", "ndjson", "csv", "badge", "markdown", "compact",
}
