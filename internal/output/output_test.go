package output_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/blu3ph4ntom/codeye/internal/output"
	"github.com/blu3ph4ntom/codeye/internal/scanner"
)

func makeResult() *scanner.ScanResult {
	return &scanner.ScanResult{
		Repo:   "/test/repo",
		Ref:    "main",
		ScanMs: 42,
		Files:  10,
		Total: scanner.LangStats{
			Name:    "Total",
			Files:   10,
			Code:    1000,
			Blank:   100,
			Comment: 50,
			Lines:   1150,
		},
		Langs: []scanner.LangStats{
			{Name: "Go", Files: 7, Code: 800, Blank: 80, Comment: 40, Lines: 920},
			{Name: "Python", Files: 3, Code: 200, Blank: 20, Comment: 10, Lines: 230},
		},
	}
}

func noColor() output.RenderOpts {
	return output.RenderOpts{NoColor: true}
}

func TestTableRenderer(t *testing.T) {
	r := &output.TableRenderer{}
	var buf bytes.Buffer
	if err := r.Render(&buf, makeResult(), noColor()); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	// Check key content
	if !strings.Contains(out, "Go") {
		t.Error("output missing Go")
	}
	if !strings.Contains(out, "Python") {
		t.Error("output missing Python")
	}
	if !strings.Contains(out, "Total") {
		t.Error("output missing Total row")
	}
	if !strings.Contains(out, "42ms") {
		t.Error("output missing scan time")
	}
}

func TestTableRendererCompact(t *testing.T) {
	r := &output.TableRenderer{}
	var buf bytes.Buffer
	opts := noColor()
	opts.Compact = true
	if err := r.Render(&buf, makeResult(), opts); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.Count(out, "\n") > 2 {
		t.Errorf("compact should be ~1 line, got %d lines", strings.Count(out, "\n"))
	}
}

func TestJSONRenderer(t *testing.T) {
	r := &output.JSONRenderer{}
	var buf bytes.Buffer
	if err := r.Render(&buf, makeResult(), noColor()); err != nil {
		t.Fatal(err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, buf.String())
	}
	if out["repo"] != "/test/repo" {
		t.Errorf("repo mismatch: %v", out["repo"])
	}
	langs, ok := out["languages"].([]interface{})
	if !ok {
		t.Fatal("languages not an array")
	}
	if len(langs) != 2 {
		t.Errorf("expected 2 languages, got %d", len(langs))
	}
}

func TestNDJSONRenderer(t *testing.T) {
	r := &output.NDJSONRenderer{}
	var buf bytes.Buffer
	if err := r.Render(&buf, makeResult(), noColor()); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 NDJSON lines, got %d", len(lines))
	}
	for i, line := range lines {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Errorf("line %d is invalid JSON: %v", i, err)
		}
	}
}

func TestCSVRenderer(t *testing.T) {
	r := &output.CSVRenderer{}
	var buf bytes.Buffer
	if err := r.Render(&buf, makeResult(), noColor()); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	// header + 2 langs + total = 4 lines
	if len(lines) != 4 {
		t.Errorf("expected 4 CSV lines, got %d: %q", len(lines), out)
	}
	if !strings.HasPrefix(lines[0], "language,") {
		t.Errorf("first line should be header, got: %q", lines[0])
	}
}

func TestBarRenderer(t *testing.T) {
	r := &output.BarRenderer{}
	var buf bytes.Buffer
	if err := r.Render(&buf, makeResult(), noColor()); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Go") {
		t.Error("bar output missing Go")
	}
	if !strings.Contains(out, "█") {
		t.Error("bar output missing bar character")
	}
}

func TestBadgeRenderer(t *testing.T) {
	r := &output.BadgeRenderer{}
	var buf bytes.Buffer
	if err := r.Render(&buf, makeResult(), noColor()); err != nil {
		t.Fatal(err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("badge is invalid JSON: %v", err)
	}
	if out["label"] != "lines of code" {
		t.Errorf("wrong label: %v", out["label"])
	}
	if out["schemaVersion"].(float64) != 1 {
		t.Errorf("wrong schemaVersion: %v", out["schemaVersion"])
	}
}

func TestGet(t *testing.T) {
	formats := []string{"table", "bar", "spark", "json", "ndjson", "csv", "badge", "unknown"}
	for _, f := range formats {
		r := output.Get(f)
		if r == nil {
			t.Errorf("Get(%q) returned nil", f)
		}
	}
}

func TestSparkline(t *testing.T) {
	vals := []int64{0, 10, 20, 50, 100, 50, 30}
	spark := output.Spark(vals)
	if len([]rune(spark)) != len(vals) {
		t.Errorf("spark length %d, want %d", len([]rune(spark)), len(vals))
	}
	// First char should be smallest (▁)
	runes := []rune(spark)
	if runes[0] != '▁' {
		t.Errorf("first spark char should be ▁, got %c", runes[0])
	}
	// Last char should be max (█)
	if runes[4] != '█' {
		t.Errorf("peak spark char should be █, got %c", runes[4])
	}
}

func TestTopFilter(t *testing.T) {
	r := &output.TableRenderer{}
	var buf bytes.Buffer
	opts := noColor()
	opts.Top = 1
	result := makeResult()
	if err := r.Render(&buf, result, opts); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.Contains(out, "Python") && !strings.Contains(out, "Total") {
		t.Error("with top=1, Python should be filtered out")
	}
}
