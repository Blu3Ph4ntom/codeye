package scanner

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// LangStats holds counts for a single language.
type LangStats struct {
	Name    string
	Files   int
	Code    int64
	Blank   int64
	Comment int64
	Lines   int64
}

// Total returns total lines.
func (s *LangStats) Total() int64 {
	return s.Code + s.Blank + s.Comment
}

// ScanResult is the aggregated output of a scan.
type ScanResult struct {
	Repo    string
	Ref     string
	TreeSHA string
	ScanMs  int64
	Cached  bool
	Files   int
	Total   LangStats
	Langs   []LangStats // sorted by Lines desc
}

// ScanOpts controls scanning behavior.
type ScanOpts struct {
	RepoRoot    string
	Ref         string   // git ref to scan (default: working tree)
	Exclude     []string // glob patterns to exclude
	Include     []string // glob patterns to include (empty = all)
	NoVendor    bool
	NoGenerated bool
	NoTests     bool
	LangFilter  []string // only these languages
	MinLines    int
	Workers     int
	DryRun      bool // list files but don't count
}

// fileJob is a unit of work sent to a worker goroutine.
type fileJob struct {
	path string
}

// fileResult is the result from a worker goroutine.
type fileResult struct {
	lang    string
	code    int64
	blank   int64
	comment int64
	err     error
}

// Scan performs a full LoC scan of the files returned by git ls-files.
// It uses a goroutine pool for parallel file reading.
func Scan(files []string, root string, opts ScanOpts) (*ScanResult, error) {
	workers := opts.Workers
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}

	jobs := make(chan fileJob, workers*4)
	results := make(chan fileResult, workers*4)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				res := processFile(filepath.Join(root, job.path))
				results <- res
			}
		}()
	}

	// Close results when all workers done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Feed jobs
	go func() {
		for _, f := range files {
			jobs <- fileJob{path: f}
		}
		close(jobs)
	}()

	// Aggregate results
	langMap := make(map[string]*LangStats)
	totalFiles := 0
	for res := range results {
		if res.err != nil {
			continue // skip unreadable files silently
		}
		totalFiles++
		ls, ok := langMap[res.lang]
		if !ok {
			ls = &LangStats{Name: res.lang}
			langMap[res.lang] = ls
		}
		ls.Files++
		ls.Code += res.code
		ls.Blank += res.blank
		ls.Comment += res.comment
		ls.Lines = ls.Code + ls.Blank + ls.Comment
	}

	// Filter by min lines
	if opts.MinLines > 0 {
		for k, v := range langMap {
			if v.Lines < int64(opts.MinLines) {
				delete(langMap, k)
			}
		}
	}

	// Build result
	sr := &ScanResult{
		Repo:  root,
		Ref:   opts.Ref,
		Files: totalFiles,
	}

	var totalCode, totalBlank, totalComment int64
	for _, ls := range langMap {
		sr.Langs = append(sr.Langs, *ls)
		totalCode += ls.Code
		totalBlank += ls.Blank
		totalComment += ls.Comment
	}

	sr.Total = LangStats{
		Name:    "Total",
		Files:   totalFiles,
		Code:    totalCode,
		Blank:   totalBlank,
		Comment: totalComment,
		Lines:   totalCode + totalBlank + totalComment,
	}

	SortLangs(sr.Langs, "lines", true)
	return sr, nil
}

// processFile reads a single file and counts its lines.
func processFile(path string) fileResult {
	// Read first 512 bytes for shebang detection, then full file
	f, err := os.Open(path)
	if err != nil {
		return fileResult{err: err}
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return fileResult{err: err}
	}

	// Skip very large files (>50MB) — probably binary or data
	if fi.Size() > 50*1024*1024 {
		lang := DetectLanguage(path, nil)
		return fileResult{lang: lang, code: 0, blank: 0, comment: 0}
	}

	// Read entire file
	buf := make([]byte, fi.Size())
	n, err := f.Read(buf)
	if err != nil && n == 0 {
		return fileResult{err: fmt.Errorf("read %s: %w", path, err)}
	}
	buf = buf[:n]

	// Skip binary files (null bytes in first 8kb)
	checkLen := 8192
	if len(buf) < checkLen {
		checkLen = len(buf)
	}
	if bytes.IndexByte(buf[:checkLen], 0) >= 0 {
		return fileResult{lang: "Binary"}
	}

	// Detect language using first 128 bytes for shebang
	shebangBuf := buf
	if len(shebangBuf) > 128 {
		shebangBuf = shebangBuf[:128]
	}
	lang := DetectLanguage(path, shebangBuf)
	def := GetLangDef(lang)

	code, blank, comment := countLines(buf, def)
	return fileResult{
		lang:    lang,
		code:    code,
		blank:   blank,
		comment: comment,
	}
}

// countLines counts code, blank, and comment lines in buf.
func countLines(buf []byte, def LangDef) (code, blank, comment int64) {
	lines := bytes.Split(buf, []byte("\n"))
	// Remove trailing empty element from files ending with newline
	if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}

	inBlock := false
	blockEndIdx := -1

	for _, rawLine := range lines {
		line := bytes.TrimSpace(rawLine)

		// Blank line
		if len(line) == 0 {
			blank++
			continue
		}

		lineStr := string(line)

		// Block comment tracking
		if inBlock {
			comment++
			// Check for block end on this line
			if blockEndIdx < len(def.BlockEnd) {
				end := def.BlockEnd[blockEndIdx]
				if strings.Contains(lineStr, end) {
					inBlock = false
				}
			}
			continue
		}

		// Check for block comment or docstring start
		blockStarted := false
		for i, bs := range def.BlockStart {
			if strings.HasPrefix(lineStr, bs) {
				comment++
				inBlock = true
				blockEndIdx = i
				// Check if block ends on same line
				endIdx := strings.Index(lineStr[len(bs):], def.BlockEnd[i])
				if endIdx >= 0 {
					inBlock = false
				}
				blockStarted = true
				break
			}
		}
		if blockStarted {
			continue
		}

		// Check docstrings
		for i, ds := range def.DocstringStart {
			if strings.HasPrefix(lineStr, ds) {
				comment++
				rest := lineStr[len(ds):]
				// Check if docstring ends on same line (with content)
				endToken := def.DocstringEnd[i]
				if strings.Contains(rest, endToken) {
					// Single-line docstring
				} else {
					inBlock = true
					blockEndIdx = -1
					// Use docstring end as block end (simple approximation)
					// We'll handle this by looking for the end token
					_ = endToken
				}
				blockStarted = true
				break
			}
		}
		if blockStarted {
			continue
		}

		// Check for line comment
		isComment := false
		for _, lc := range def.LineComment {
			if strings.HasPrefix(lineStr, lc) {
				comment++
				isComment = true
				break
			}
		}
		if !isComment {
			code++
		}
	}
	return
}

// FilterFiles applies exclusion/inclusion rules to a list of file paths.
func FilterFiles(files []string, opts ScanOpts) []string {
	out := make([]string, 0, len(files))
	for _, f := range files {
		if shouldExclude(f, opts) {
			continue
		}
		out = append(out, f)
	}
	return out
}

func shouldExclude(path string, opts ScanOpts) bool {
	normalized := filepath.ToSlash(path)

	if opts.NoVendor {
		for _, v := range vendorPrefixes {
			if strings.HasPrefix(normalized, v) || strings.Contains(normalized, "/"+v) {
				return true
			}
		}
	}

	if opts.NoGenerated {
		base := filepath.Base(path)
		for _, pat := range generatedSuffixes {
			if matchGlob(pat, base) {
				return true
			}
		}
	}

	if opts.NoTests {
		base := filepath.Base(path)
		for _, pat := range testPatterns {
			if matchGlob(pat, base) {
				return true
			}
		}
	}

	for _, pat := range opts.Exclude {
		if matchGlob(pat, normalized) || matchGlob(pat, filepath.Base(path)) {
			return true
		}
	}

	if len(opts.Include) > 0 {
		for _, pat := range opts.Include {
			if matchGlob(pat, normalized) {
				return false
			}
		}
		return true
	}

	return false
}

// matchGlob is a simple glob matcher supporting * and ? wildcards.
func matchGlob(pattern, name string) bool {
	matched, _ := filepath.Match(pattern, name)
	return matched
}

var vendorPrefixes = []string{
	"vendor/", "node_modules/", "bower_components/",
	".yarn/", "Pods/", "Carthage/",
	"__pycache__/", ".venv/", "venv/", "env/", ".env/",
	"target/", ".gradle/",
}

var generatedSuffixes = []string{
	"*.pb.go", "*.pb.gw.go", "*.gen.go", "*_generated.go",
	"*.min.js", "*.min.css",
}

var testPatterns = []string{
	"*_test.go", "*.spec.ts", "*.spec.js", "*.test.ts", "*.test.js",
	"*_test.py", "test_*.py", "*_spec.rb",
}
