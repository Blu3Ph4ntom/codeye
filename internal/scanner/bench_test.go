package scanner_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/codeye/codeye/internal/scanner"
)

// generateFiles creates n synthetic source files of the given language extension
// in a temp directory, each containing lineCount lines of code.
func generateFiles(b *testing.B, n, lineCount int, ext string) (dir string, files []string) {
	b.Helper()
	dir = b.TempDir()

	line := "x := x + 1 // computation\n"
	body := ""
	for i := 0; i < lineCount; i++ {
		body += line
	}
	content := []byte("package bench\n\n" + body)

	for i := 0; i < n; i++ {
		name := filepath.Join(dir, fmt.Sprintf("file_%04d%s", i, ext))
		if err := os.WriteFile(name, content, 0o644); err != nil {
			b.Fatal(err)
		}
		files = append(files, name)
	}
	return dir, files
}

// BenchmarkScan_100files_500lines — small repo baseline
func BenchmarkScan_100files_500lines(b *testing.B) {
	dir, files := generateFiles(b, 100, 500, ".go")
	opts := scanner.ScanOpts{Workers: 8}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := scanner.Scan(files, dir, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkScan_1000files_500lines — medium repo
func BenchmarkScan_1000files_500lines(b *testing.B) {
	dir, files := generateFiles(b, 1000, 500, ".go")
	opts := scanner.ScanOpts{Workers: 8}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := scanner.Scan(files, dir, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkScan_5000files_500lines — large repo
func BenchmarkScan_5000files_500lines(b *testing.B) {
	dir, files := generateFiles(b, 5000, 500, ".go")
	opts := scanner.ScanOpts{Workers: 8}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := scanner.Scan(files, dir, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkScan_1000files_5000lines — large files
func BenchmarkScan_1000files_5000lines(b *testing.B) {
	dir, files := generateFiles(b, 1000, 5000, ".go")
	opts := scanner.ScanOpts{Workers: 8}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := scanner.Scan(files, dir, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkScan_Workers — worker count scaling
func BenchmarkScan_Workers(b *testing.B) {
	dir, files := generateFiles(b, 500, 500, ".go")
	for _, w := range []int{1, 2, 4, 8, 16, 32} {
		b.Run(fmt.Sprintf("workers=%d", w), func(b *testing.B) {
			opts := scanner.ScanOpts{Workers: w}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := scanner.Scan(files, dir, opts)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkDetectLanguage — language detection microbenchmark
func BenchmarkDetectLanguage(b *testing.B) {
	cases := []string{
		"main.go", "index.ts", "app.py", "server.rs", "Main.java",
		"style.css", "index.html", "config.yaml", "Makefile", "Dockerfile",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, name := range cases {
			scanner.DetectLanguage(name, nil)
		}
	}
}

// BenchmarkScan_MixedLanguages — real-world polyglot repo pattern
func BenchmarkScan_MixedLanguages(b *testing.B) {
	dir := b.TempDir()
	exts := []string{".go", ".ts", ".py", ".js", ".md", ".yaml", ".json", ".sh"}
	var files []string
	for _, ext := range exts {
		d, ff := generateFiles(b, 50, 300, ext)
		_ = d
		// rewrite paths into single dir
		for _, f := range ff {
			base := filepath.Base(f)
			dst := filepath.Join(dir, ext[1:]+"_"+base)
			data, _ := os.ReadFile(f)
			_ = os.WriteFile(dst, data, 0o644)
			files = append(files, dst)
		}
	}
	opts := scanner.ScanOpts{Workers: 8}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := scanner.Scan(files, dir, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}
