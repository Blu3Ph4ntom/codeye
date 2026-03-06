package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/codeye/codeye/internal/scanner"
)

// CompactRenderer renders a single line summary of the scan results.
type CompactRenderer struct{}

func (r *CompactRenderer) Render(w io.Writer, result *scanner.ScanResult, opts RenderOpts) error {
	var names []string
	for i, l := range result.Langs {
		if i >= 3 {
			break
		}
		names = append(names, l.Name)
	}

	fmt.Fprintf(w, "codeye · %d lines · %d files · %s · %dms\n",
		result.Total.Lines,
		result.Total.Files,
		strings.Join(names, " "),
		result.ScanMs,
	)
	return nil
}
