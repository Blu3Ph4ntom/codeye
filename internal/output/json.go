package output

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/codeye/codeye/internal/scanner"
)

// JSONRenderer renders full scan result as JSON.
type JSONRenderer struct{}

type jsonOutput struct {
	Repo      string        `json:"repo"`
	Ref       string        `json:"ref"`
	TreeSHA   string        `json:"tree_sha,omitempty"`
	ScanMs    int64         `json:"scan_ms"`
	Cached    bool          `json:"cached"`
	ScannedAt time.Time     `json:"scanned_at"`
	Total     jsonLang      `json:"total"`
	Languages []jsonLang    `json:"languages"`
}

type jsonLang struct {
	Name    string   `json:"name"`
	Files   int      `json:"files"`
	Code    int64    `json:"code"`
	Blank   int64    `json:"blank"`
	Comment int64    `json:"comment"`
	Lines   int64    `json:"lines"`
	Pct     float64  `json:"pct,omitempty"`
}

func (j *JSONRenderer) Render(w io.Writer, result *scanner.ScanResult, opts RenderOpts) error {
	langs := filteredLangs(result, opts)
	jLangs := make([]jsonLang, len(langs))
	for i, l := range langs {
		pct := 0.0
		if result.Total.Lines > 0 {
			pct = float64(l.Lines) / float64(result.Total.Lines) * 100
		}
		jLangs[i] = jsonLang{
			Name: l.Name, Files: l.Files,
			Code: l.Code, Blank: l.Blank,
			Comment: l.Comment, Lines: l.Lines,
			Pct: pct,
		}
	}
	out := jsonOutput{
		Repo:      result.Repo,
		Ref:       result.Ref,
		TreeSHA:   result.TreeSHA,
		ScanMs:    result.ScanMs,
		Cached:    result.Cached,
		ScannedAt: time.Now().UTC(),
		Total: jsonLang{
			Name: "Total", Files: result.Total.Files,
			Code: result.Total.Code, Blank: result.Total.Blank,
			Comment: result.Total.Comment, Lines: result.Total.Lines,
		},
		Languages: jLangs,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

// NDJSONRenderer renders one JSON object per language, newline-delimited.
type NDJSONRenderer struct{}

func (n *NDJSONRenderer) Render(w io.Writer, result *scanner.ScanResult, opts RenderOpts) error {
	langs := filteredLangs(result, opts)
	for _, l := range langs {
		pct := 0.0
		if result.Total.Lines > 0 {
			pct = float64(l.Lines) / float64(result.Total.Lines) * 100
		}
		obj := jsonLang{
			Name: l.Name, Files: l.Files,
			Code: l.Code, Blank: l.Blank,
			Comment: l.Comment, Lines: l.Lines,
			Pct: pct,
		}
		data, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, string(data))
	}
	return nil
}
