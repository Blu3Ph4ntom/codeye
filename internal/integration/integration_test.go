// Package integration contains end-to-end tests that run the codeye binary
// against real git repositories. Tests skip when the binary is not found or
// the test environment is missing git.
package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var (
	buildOnce sync.Once
	builtBin  string
	buildErr  error
	buildOut  []byte
)

// binaryPath returns the path to a freshly built codeye binary for the current test run.
func binaryPath(t *testing.T) string {
	t.Helper()

	// Walk up from this file to the repo root.
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// testdata/integration/<file> → repo root is ../../..
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")

	buildOnce.Do(func() {
		tmpDir, err := os.MkdirTemp("", "codeye-integration-*")
		if err != nil {
			buildErr = err
			return
		}
		bin := filepath.Join(tmpDir, "codeye")
		if runtime.GOOS == "windows" {
			bin += ".exe"
		}
		cmd := exec.Command("go", "build", "-o", bin, "./cmd/codeye")
		cmd.Dir = repoRoot
		cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
		if out, err := cmd.CombinedOutput(); err != nil {
			buildErr = err
			buildOut = out
			return
		}
		builtBin = bin
	})
	if buildErr != nil {
		t.Fatalf("build failed: %v\n%s", buildErr, buildOut)
	}
	return builtBin
}

// requireGit skips the test if git is not available.
func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

// repoRoot returns the codeye repository root (used as a real git repo fixture).
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

// run executes codeye with the given arguments in dir and returns stdout.
func run(t *testing.T, dir, binary string, args ...string) string {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		stderr := ""
		if ee, ok := err.(*exec.ExitError); ok {
			stderr = string(ee.Stderr)
		}
		t.Fatalf("codeye %v: %v\nstderr: %s", args, err, stderr)
	}
	return strings.TrimSpace(string(out))
}

// ─── Tests ───────────────────────────────────────────────────────────────────

func TestVersion(t *testing.T) {
	bin := binaryPath(t)
	out := run(t, t.TempDir(), bin, "version")
	if !strings.Contains(out, "codeye") {
		t.Errorf("version output missing 'codeye': %q", out)
	}
}

func TestScanJSON(t *testing.T) {
	requireGit(t)
	bin := binaryPath(t)
	root := repoRoot(t)

	out := run(t, root, bin, "--format=json", ".")
	if out == "" {
		t.Fatal("empty output")
	}

	var result struct {
		Languages []struct {
			Name  string `json:"language"`
			Lines int64  `json:"lines"`
		} `json:"languages"`
		Total struct {
			Lines int64 `json:"lines"`
		} `json:"total"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(result.Languages) == 0 {
		t.Error("no languages in result")
	}
	if result.Total.Lines == 0 {
		t.Error("total lines is 0")
	}
	t.Logf("scanned %d languages, %d total lines", len(result.Languages), result.Total.Lines)
}

func TestScanTable(t *testing.T) {
	requireGit(t)
	bin := binaryPath(t)
	root := repoRoot(t)

	out := run(t, root, bin, "--no-color", ".")
	if !strings.Contains(out, "Go") {
		t.Errorf("expected 'Go' in table output, got:\n%s", out)
	}
	if !strings.Contains(out, "Code") {
		t.Errorf("expected 'Code' header in table output, got:\n%s", out)
	}
}

func TestScanCSV(t *testing.T) {
	requireGit(t)
	bin := binaryPath(t)
	root := repoRoot(t)

	out := run(t, root, bin, "--format=csv", ".")
	lines := strings.Split(out, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 CSV lines, got %d", len(lines))
	}
	header := lines[0]
	if !strings.Contains(header, "language") {
		t.Errorf("CSV header missing 'language': %q", header)
	}
}

func TestScanBar(t *testing.T) {
	requireGit(t)
	bin := binaryPath(t)
	root := repoRoot(t)

	out := run(t, root, bin, "--format=bar", ".")
	if !strings.Contains(out, "█") && !strings.Contains(out, "▏") {
		t.Errorf("expected bar chart chars in output:\n%s", out)
	}
}

func TestTopN(t *testing.T) {
	requireGit(t)
	bin := binaryPath(t)
	root := repoRoot(t)

	out := run(t, root, bin, "--format=json", "--top=1", ".")
	var result struct {
		Languages []struct {
			Name string `json:"language"`
		} `json:"languages"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(result.Languages) != 1 {
		t.Errorf("expected 1 language with --top=1, got %d", len(result.Languages))
	}
}

func TestLangCSVFilter(t *testing.T) {
	requireGit(t)
	bin := binaryPath(t)
	root := repoRoot(t)

	out := run(t, root, bin, "--format=json", "--lang=Go,Markdown", ".")
	var result struct {
		Languages []struct {
			Name string `json:"name"`
		} `json:"languages"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(result.Languages) != 2 {
		t.Fatalf("expected 2 filtered languages, got %d", len(result.Languages))
	}
	names := map[string]bool{}
	for _, lang := range result.Languages {
		names[lang.Name] = true
	}
	if !names["Go"] || !names["Markdown"] {
		t.Fatalf("unexpected filtered languages: %#v", result.Languages)
	}
}

func TestConfigFileApplied(t *testing.T) {
	requireGit(t)
	bin := binaryPath(t)
	root := repoRoot(t)

	configDir := t.TempDir()
	configPath := filepath.Join(configDir, ".codeye.toml")
	configBody := "format = \"json\"\nlang = [\"Go\"]\nno_color = true\n"
	if err := os.WriteFile(configPath, []byte(configBody), 0o644); err != nil {
		t.Fatal(err)
	}

	out := run(t, root, bin, "--config", configPath, ".")
	var result struct {
		Languages []struct {
			Name string `json:"name"`
		} `json:"languages"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(result.Languages) != 1 || result.Languages[0].Name != "Go" {
		t.Fatalf("unexpected languages from config filter: %#v", result.Languages)
	}
}

func TestDoctor(t *testing.T) {
	requireGit(t)
	bin := binaryPath(t)
	root := repoRoot(t)

	cmd := exec.Command(bin, "doctor")
	cmd.Dir = root
	out, _ := cmd.CombinedOutput() // doctor may exit non-zero if some checks fail
	if !strings.Contains(string(out), "git") {
		t.Errorf("doctor output missing git info:\n%s", out)
	}
}

func TestLangsCommand(t *testing.T) {
	bin := binaryPath(t)
	out := run(t, t.TempDir(), bin, "langs")
	if !strings.Contains(out, "language") {
		t.Errorf("langs output missing language info: %q", out)
	}
}

func TestCacheStatus(t *testing.T) {
	requireGit(t)
	bin := binaryPath(t)
	root := repoRoot(t)

	out := run(t, root, bin, "cache", "status")
	// Just check it does not crash; format may vary.
	_ = out
}

func TestSortByCode(t *testing.T) {
	requireGit(t)
	bin := binaryPath(t)
	root := repoRoot(t)

	out := run(t, root, bin, "--format=json", "--sort=code", ".")
	var result struct {
		Languages []struct {
			Code int64 `json:"code"`
		} `json:"languages"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	for i := 1; i < len(result.Languages); i++ {
		if result.Languages[i].Code > result.Languages[i-1].Code {
			t.Errorf("results not sorted by code desc at index %d", i)
		}
	}
}
