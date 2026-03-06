package scanner_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/blu3ph4ntom/codeye/internal/scanner"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"main.go", "Go"},
		{"lib.rs", "Rust"},
		{"app.py", "Python"},
		{"index.ts", "TypeScript"},
		{"index.js", "JavaScript"},
		{"app.tsx", "TSX"},
		{"styles.css", "CSS"},
		{"layout.scss", "SCSS"},
		{"index.html", "HTML"},
		{"config.yaml", "YAML"},
		{"config.yml", "YAML"},
		{"Makefile", "Makefile"},
		{"Dockerfile", "Dockerfile"},
		{"go.mod", "Go"},
		{"README.md", "Markdown"},
		{"setup.py", "Python"},
		{"Gemfile", "Ruby"},
		{"unknown.xyz", "Unknown"},
		{".gitignore", "Gitignore"}, // recognized as Gitignore language
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := scanner.DetectLanguage(tt.path, nil)
			if got != tt.want {
				t.Errorf("DetectLanguage(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestShebangDetection(t *testing.T) {
	tests := []struct {
		path    string
		shebang []byte
		want    string
	}{
		{"script", []byte("#!/usr/bin/env python3\ncode"), "Python"},
		{"run", []byte("#!/bin/bash\ncode"), "Shell"},
		{"server", []byte("#!/usr/bin/env node\ncode"), "JavaScript"},
		{"task", []byte("#!/usr/bin/env ruby\ncode"), "Ruby"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := scanner.DetectLanguage(tt.path, tt.shebang)
			if got != tt.want {
				t.Errorf("DetectLanguage with shebang %q = %q, want %q",
					tt.path, got, tt.want)
			}
		})
	}
}

func TestCountLines(t *testing.T) {
	// Create a temp Go file and scan it
	dir := t.TempDir()
	initGitRepo(t, dir)

	src := `package main

// Package comment
import "fmt"

/*
 * Block comment
 */
func main() {
    fmt.Println("hello")
}
`
	err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Stage file for git ls-files to pick it up
	runCmd(t, dir, "git", "add", ".")

	files := []string{"main.go"}
	opts := scanner.ScanOpts{RepoRoot: dir, Workers: 1}
	result, err := scanner.Scan(files, dir, opts)
	if err != nil {
		t.Fatal(err)
	}

	if result.Total.Lines == 0 {
		t.Error("expected non-zero total lines")
	}

	var goStats *scanner.LangStats
	for _, l := range result.Langs {
		if l.Name == "Go" {
			ls := l
			goStats = &ls
			break
		}
	}
	if goStats == nil {
		t.Fatal("Go not detected")
	}
	if goStats.Comment == 0 {
		t.Error("expected comment lines > 0")
	}
	if goStats.Blank == 0 {
		t.Error("expected blank lines > 0")
	}
	if goStats.Code == 0 {
		t.Error("expected code lines > 0")
	}
}

func TestFilterFiles(t *testing.T) {
	files := []string{
		"main.go",
		"main_test.go",
		"vendor/foo/bar.go",
		"node_modules/pkg/index.js",
		"src/app.ts",
		"generated.pb.go",
	}

	t.Run("no-vendor", func(t *testing.T) {
		opts := scanner.ScanOpts{NoVendor: true}
		got := scanner.FilterFiles(files, opts)
		for _, f := range got {
			if f == "vendor/foo/bar.go" || f == "node_modules/pkg/index.js" {
				t.Errorf("vendor file not excluded: %s", f)
			}
		}
	})

	t.Run("no-tests", func(t *testing.T) {
		opts := scanner.ScanOpts{NoTests: true}
		got := scanner.FilterFiles(files, opts)
		for _, f := range got {
			if f == "main_test.go" {
				t.Errorf("test file not excluded: %s", f)
			}
		}
	})

	t.Run("no-generated", func(t *testing.T) {
		opts := scanner.ScanOpts{NoGenerated: true}
		got := scanner.FilterFiles(files, opts)
		for _, f := range got {
			if f == "generated.pb.go" {
				t.Errorf("generated file not excluded: %s", f)
			}
		}
	})

	t.Run("exclude glob", func(t *testing.T) {
		opts := scanner.ScanOpts{Exclude: []string{"*.ts"}}
		got := scanner.FilterFiles(files, opts)
		for _, f := range got {
			if f == "src/app.ts" {
				t.Errorf("excluded file not filtered: %s", f)
			}
		}
	})
}

func TestSortLangs(t *testing.T) {
	langs := []scanner.LangStats{
		{Name: "Python", Lines: 100},
		{Name: "Go", Lines: 500},
		{Name: "Rust", Lines: 300},
	}

	scanner.SortLangs(langs, "lines", true)
	if langs[0].Name != "Go" {
		t.Errorf("expected Go first, got %s", langs[0].Name)
	}

	scanner.SortLangs(langs, "lines", false)
	if langs[0].Name != "Python" {
		t.Errorf("expected Python first (asc), got %s", langs[0].Name)
	}

	scanner.SortLangs(langs, "lang", true)
	if langs[0].Name != "Rust" {
		t.Errorf("expected Rust first (lang desc), got %s", langs[0].Name)
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	runCmd(t, dir, "git", "init", "-b", "main")
	runCmd(t, dir, "git", "config", "user.email", "test@test.com")
	runCmd(t, dir, "git", "config", "user.name", "test")
}

func runCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	// Use os/exec directly
	cmd := newCmd(dir, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Logf("cmd %v: %s", args, out)
	}
}
