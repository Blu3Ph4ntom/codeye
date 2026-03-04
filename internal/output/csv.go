package output

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/codeye/codeye/internal/scanner"
)

// CSVRenderer renders scan results as CSV.
type CSVRenderer struct{}

func (c *CSVRenderer) Render(w io.Writer, result *scanner.ScanResult, opts RenderOpts) error {
	cw := csv.NewWriter(w)

	header := []string{"language", "files", "code", "blank", "comment", "total"}
	if opts.Pct {
		header = append(header, "pct")
	}
	if err := cw.Write(header); err != nil {
		return err
	}

	langs := filteredLangs(result, opts)
	for _, l := range langs {
		row := []string{
			l.Name,
			fmt.Sprintf("%d", l.Files),
			fmt.Sprintf("%d", l.Code),
			fmt.Sprintf("%d", l.Blank),
			fmt.Sprintf("%d", l.Comment),
			fmt.Sprintf("%d", l.Lines),
		}
		if opts.Pct {
			pct := 0.0
			if result.Total.Lines > 0 {
				pct = float64(l.Lines) / float64(result.Total.Lines) * 100
			}
			row = append(row, fmt.Sprintf("%.2f", pct))
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}

	// Total row
	total := result.Total
	totRow := []string{
		"Total",
		fmt.Sprintf("%d", total.Files),
		fmt.Sprintf("%d", total.Code),
		fmt.Sprintf("%d", total.Blank),
		fmt.Sprintf("%d", total.Comment),
		fmt.Sprintf("%d", total.Lines),
	}
	if opts.Pct {
		totRow = append(totRow, "100.00")
	}
	if err := cw.Write(totRow); err != nil {
		return err
	}

	cw.Flush()
	return cw.Error()
}
