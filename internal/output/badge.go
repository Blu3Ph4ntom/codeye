package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/codeye/codeye/internal/scanner"
	"github.com/dustin/go-humanize"
)

// BadgeRenderer renders a shields.io-compatible JSON endpoint response.
type BadgeRenderer struct{}

type badgeOutput struct {
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
	Color         string `json:"color"`
}

func (b *BadgeRenderer) Render(w io.Writer, result *scanner.ScanResult, opts RenderOpts) error {
	total := result.Total.Lines
	msg := humanize.SI(float64(total), "")
	if msg == "" {
		msg = fmt.Sprintf("%d", total)
	}

	// Trim trailing space from SI formatter
	for len(msg) > 0 && msg[len(msg)-1] == ' ' {
		msg = msg[:len(msg)-1]
	}

	color := badgeColor(total)
	out := badgeOutput{
		SchemaVersion: 1,
		Label:         "lines of code",
		Message:       msg,
		Color:         color,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func badgeColor(lines int64) string {
	switch {
	case lines < 1000:
		return "green"
	case lines < 10000:
		return "brightgreen"
	case lines < 100000:
		return "blue"
	case lines < 1000000:
		return "orange"
	default:
		return "red"
	}
}
