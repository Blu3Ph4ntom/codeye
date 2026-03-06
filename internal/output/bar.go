package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/blu3ph4ntom/codeye/internal/scanner"
	"github.com/dustin/go-humanize"
	"github.com/muesli/termenv"
)

// BarRenderer renders a horizontal bar chart.
type BarRenderer struct{}

const barWidth = 40

func (b *BarRenderer) Render(w io.Writer, result *scanner.ScanResult, opts RenderOpts) error {
	prof := termenv.ColorProfile()
	if opts.NoColor {
		prof = termenv.Ascii
	}

	langs := filteredLangs(result, opts)
	if len(langs) == 0 {
		fmt.Fprintln(w, "no files found")
		return nil
	}

	maxName := 8
	for _, l := range langs {
		if len(l.Name) > maxName {
			maxName = len(l.Name)
		}
	}

	maxLines := langs[0].Lines // sorted desc
	if maxLines == 0 {
		maxLines = 1
	}

	blocks := []string{"▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"}

	for _, l := range langs {
		pct := float64(l.Lines) / float64(result.Total.Lines) * 100
		barFrac := float64(l.Lines) / float64(maxLines) * float64(barWidth)
		full := int(barFrac)
		frac := barFrac - float64(full)

		bar := strings.Repeat("█", full)
		if frac >= 0.1 {
			idx := int(frac * 8)
			if idx > 7 {
				idx = 7
			}
			bar += blocks[idx]
		}

		color := brandColorFor(l.Name)
		styledBar := termenv.String(bar).Foreground(prof.Color(color)).String()

		fmt.Fprintf(w, " %-*s  %s%-*s  %5.1f%%  %s\n",
			maxName, l.Name,
			styledBar,
			barWidth-len(bar), "",
			pct,
			humanize.Comma(l.Lines),
		)
	}

	fmt.Fprintf(w, "\n total %s lines\n", humanize.Comma(result.Total.Lines))
	return nil
}
