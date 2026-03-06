package output

import (
	"fmt"
	"io"

	"github.com/blu3ph4ntom/codeye/internal/scanner"
)

// MarkdownRenderer renders a summary as a markdown table.
type MarkdownRenderer struct{}

func (r *MarkdownRenderer) Render(w io.Writer, result *scanner.ScanResult, opts RenderOpts) error {
	fmt.Fprintf(w, "## LoC Scan Result\n\n")
	fmt.Fprintf(w, "| Language | Files | Code | Total |\n")
	fmt.Fprintf(w, "| :--- | ---: | ---: | ---: |\n")

	for i, l := range result.Langs {
		if opts.Top > 0 && i >= opts.Top {
			break
		}
		fmt.Fprintf(w, "| %s | %d | %d | %d |\n",
			l.Name, l.Files, l.Code, l.Lines)
	}

	fmt.Fprintf(w, "| **Total** | **%d** | **%d** | **%d** |\n\n",
		result.Total.Files, result.Total.Code, result.Total.Lines)

	fmt.Fprintf(w, "*⚡ %dms · tree %s*\n", result.ScanMs, result.TreeSHA[:8])
	return nil
}
