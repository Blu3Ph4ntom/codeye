package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveMergesHomeProjectEnv(t *testing.T) {
	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	project := filepath.Join(tmp, "repo")
	nested := filepath.Join(project, "nested")

	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	homeConfig := "format = \"compact\"\nworkers = 4\nno_color = true\n"
	if err := os.WriteFile(filepath.Join(home, ".codeye.toml"), []byte(homeConfig), 0o644); err != nil {
		t.Fatal(err)
	}

	projectConfig := "format = \"json\"\nlang = [\"Go\", \"Markdown\"]\nworkers = 8\n"
	if err := os.WriteFile(filepath.Join(project, ".codeye.toml"), []byte(projectConfig), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("CODEYE_FORMAT", "csv")
	t.Setenv("CODEYE_WORKERS", "12")
	t.Setenv("NO_COLOR", "1")

	cfg, source, err := Resolve("", nested)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	if source != filepath.Join(project, ".codeye.toml") {
		t.Fatalf("unexpected config source: %q", source)
	}
	if cfg.Format != "csv" {
		t.Fatalf("expected env override format=csv, got %q", cfg.Format)
	}
	if cfg.Workers != 12 {
		t.Fatalf("expected env override workers=12, got %d", cfg.Workers)
	}
	if !cfg.NoColor {
		t.Fatal("expected NO_COLOR to force no_color")
	}
	if len(cfg.Lang) != 2 || cfg.Lang[0] != "Go" || cfg.Lang[1] != "Markdown" {
		t.Fatalf("unexpected lang filter: %#v", cfg.Lang)
	}
}

func TestResolveExplicitConfigRequired(t *testing.T) {
	tmp := t.TempDir()
	_, _, err := Resolve(filepath.Join(tmp, "missing.toml"), tmp)
	if err == nil {
		t.Fatal("expected missing explicit config to fail")
	}
}
