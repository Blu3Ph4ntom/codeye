package output

import (
	"fmt"
	"io"

	"github.com/codeye/codeye/internal/scanner"
	"github.com/dustin/go-humanize"
)

// SparkRenderer renders sparklines for history mode.
// In standard scan mode, it falls back to a compact bar summary.
type SparkRenderer struct{}

// sparkChars are the unicode block elements used for sparklines.
var sparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// Spark generates a sparkline string from a slice of values.
func Spark(values []int64) string {
	if len(values) == 0 {
		return ""
	}
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	if max == 0 {
		max = 1
	}
	out := make([]rune, len(values))
	for i, v := range values {
		idx := int(float64(v)/float64(max)*float64(len(sparkChars)-1) + 0.5)
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sparkChars) {
			idx = len(sparkChars) - 1
		}
		out[i] = sparkChars[idx]
	}
	return string(out)
}

func (s *SparkRenderer) Render(w io.Writer, result *scanner.ScanResult, opts RenderOpts) error {
	langs := filteredLangs(result, opts)
	maxName := 8
	for _, l := range langs {
		if len(l.Name) > maxName {
			maxName = len(l.Name)
		}
	}

	fmt.Fprintf(w, " %-*s   %s   %s\n", maxName, "Language", "Spark", "Lines")
	fmt.Fprintln(w, " "+repeatStr("─", maxName+30))
	for _, l := range langs {
		spark := Spark([]int64{l.Code, l.Blank, l.Comment})
		fmt.Fprintf(w, " %-*s   %s   %s\n",
			maxName, l.Name,
			spark,
			humanize.Comma(l.Lines),
		)
	}
	return nil
}

func repeatStr(s string, n int) string {
	out := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		out = append(out, s...)
	}
	return string(out)
}
