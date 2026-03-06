package output

import "testing"

func TestIconForExt(t *testing.T) {
	cases := []struct {
		filename string
		wantNot  string // should NOT return the generic file icon
	}{
		// Exact filenames
		{"Makefile", NF.File},
		{"Dockerfile", NF.File},
		{".gitignore", NF.File},
		{"go.mod", NF.File},
		{"go.sum", NF.File},
		{"package.json", NF.File},
		{"Cargo.toml", NF.File},
		{".goreleaser.yaml", NF.File},
		{".golangci.yml", NF.File},
		{"tsconfig.json", NF.File},
		// Extensions
		{"main.go", NF.File},
		{"lib.rs", NF.File},
		{"app.py", NF.File},
		{"index.ts", NF.File},
		{"component.tsx", NF.File},
		{"style.css", NF.File},
		{"index.html", NF.File},
		{"config.yaml", NF.File},
		{"config.toml", NF.File},
		{"data.json", NF.File},
		{"schema.sql", NF.File},
		{"README.md", NF.File},
		{"script.sh", NF.File},
		{"Main.java", NF.File},
		{"App.kt", NF.File},
		{"main.swift", NF.File},
	}

	for _, tc := range cases {
		got := IconForExt(tc.filename)
		if got == tc.wantNot {
			t.Errorf("IconForExt(%q) = generic file icon, want a specific icon", tc.filename)
		}
	}
}

func TestIconForLang(t *testing.T) {
	langs := []string{
		"Go", "Rust", "Python", "JavaScript", "TypeScript",
		"Java", "Kotlin", "Swift", "Ruby", "PHP",
		"HTML", "CSS", "Shell", "Markdown", "YAML",
		"JSON", "TOML", "Dockerfile", "Makefile",
		"Gitignore", "Gitattributes",
	}
	for _, lang := range langs {
		got := IconForLang(lang)
		if got == NF.File {
			t.Errorf("IconForLang(%q) = generic file icon, want a specific icon", lang)
		}
	}
}
