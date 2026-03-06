package main

import (
"encoding/json"
"fmt"
"os"
"os/exec"
"runtime"
"strconv"
"strings"
"time"

"github.com/codeye/codeye/internal/blame"
"github.com/codeye/codeye/internal/cache"
"github.com/codeye/codeye/internal/config"
"github.com/codeye/codeye/internal/git"
"github.com/codeye/codeye/internal/history"
"github.com/codeye/codeye/internal/hotspot"
"github.com/codeye/codeye/internal/output"
"github.com/codeye/codeye/internal/scanner"
"github.com/dustin/go-humanize"
"github.com/spf13/cobra"
)

// Build-time variables injected via -ldflags.
var (
version   = "dev"
commit    = "unknown"
buildDate = "unknown"
)

func main() {
if err := newRootCmd().Execute(); err != nil {
fmt.Fprintln(os.Stderr, err)
os.Exit(1)
}
}

type runFlags struct {
branch          string
author          []string
since           string
until           string
pathFilter      []string
exclude         []string
lang            []string
noVendor        bool
noGenerated     bool
noTests         bool
minLines        int
format          string
sort            string
desc            bool
top             int
noColor         bool
noHeader        bool
wide            bool
compact         bool
pct             bool
progress        bool
nerdFont        bool
theme           string
history         bool
historyInterval string
historyLimit    int
blame           bool
hotspots        bool
contrib         bool
at              string
speedtest       bool
noCache         bool
cacheDir        string
workers         int
verbose         bool
dryRun          bool
configPath      string
}

func newRootCmd() *cobra.Command {
f := &runFlags{}

cmd := &cobra.Command{
Use:          "codeye [path]",
Short:        "eyes on your code, instantly",
Long:         "codeye — fastest CLI LoC scanner for git repos.\nSingle binary. Sub-100ms scans. Git porcelain only.",
Args:         cobra.MaximumNArgs(1),
SilenceUsage: true,
RunE: func(cmd *cobra.Command, args []string) error {
dir := "."
if len(args) > 0 {
dir = args[0]
}
cfg := flagsToConfig(f)
		// --nf flag enables Nerd Fonts; env var CODEYE_NERD_FONTS=1 also enables it (already in cfg from DefaultConfig).
		if cmd.Flags().Changed("nf") {
			cfg.NerdFont = f.nerdFont
		} else {
			// keep DefaultConfig value (env var may have set it)
		}
		return runScan(dir, cfg)
},
}

addFlags(cmd, f)
cmd.AddCommand(versionCmd(), doctorCmd(), cacheManageCmd(), langsCmd(), diffCmd(), completionCmd(cmd))
return cmd
}

func addFlags(cmd *cobra.Command, f *runFlags) {
// Scope
cmd.Flags().StringVarP(&f.branch, "branch", "b", "", "git branch/tag/SHA to scan")
cmd.Flags().StringArrayVarP(&f.author, "author", "a", nil, "filter by author email (glob)")
cmd.Flags().StringVar(&f.since, "since", "", "commits since date (e.g. 2024-01-01, 6months)")
cmd.Flags().StringVar(&f.until, "until", "", "commits until date")
cmd.Flags().StringArrayVarP(&f.pathFilter, "path-filter", "p", nil, "only scan files matching glob")
cmd.Flags().StringArrayVarP(&f.exclude, "exclude", "e", nil, "exclude files matching glob (stackable)")
cmd.Flags().StringArrayVarP(&f.lang, "lang", "l", nil, "only show these languages (comma or repeat)")
cmd.Flags().BoolVar(&f.noVendor, "no-vendor", false, "exclude vendor/node_modules/etc dirs")
cmd.Flags().BoolVar(&f.noGenerated, "no-generated", false, "exclude generated files (*.pb.go etc)")
cmd.Flags().BoolVar(&f.noTests, "no-tests", false, "exclude test files")
cmd.Flags().IntVar(&f.minLines, "min-lines", 0, "only show languages with >= N lines")
// Output
cmd.Flags().StringVarP(&f.format, "format", "f", "table", "output: table|bar|spark|json|ndjson|csv|badge|markdown|compact")
cmd.Flags().StringVarP(&f.sort, "sort", "s", "lines", "sort by: lines|files|code|blank|comment|lang")
cmd.Flags().BoolVar(&f.desc, "desc", true, "descending sort")
cmd.Flags().IntVarP(&f.top, "top", "n", 0, "show top N languages (0=all)")
cmd.Flags().BoolVar(&f.noColor, "no-color", false, "disable ANSI color")
cmd.Flags().BoolVar(&f.noHeader, "no-header", false, "suppress header/footer")
cmd.Flags().BoolVarP(&f.wide, "wide", "w", false, "show blank/comment columns")
cmd.Flags().BoolVar(&f.compact, "compact", false, "single-line summary")
	cmd.Flags().BoolVar(&f.pct, "pct", true, "show percentage column")
	cmd.Flags().BoolVar(&f.progress, "progress", false, "show live progress bar")
	cmd.Flags().BoolVar(&f.nerdFont, "nf", false, "use Nerd Font glyphs (requires a patched font: JetBrainsMono NF, FiraCode NF, etc.)")
	cmd.Flags().StringVar(&f.theme, "theme", "dark", "color theme: dark|light|mono")
// Analysis
cmd.Flags().BoolVarP(&f.history, "history", "H", false, "LoC growth over git history")
cmd.Flags().StringVar(&f.historyInterval, "history-interval", "week", "day|week|month|quarter|year")
cmd.Flags().IntVar(&f.historyLimit, "history-limit", 500, "max commits to walk")
cmd.Flags().BoolVar(&f.blame, "blame", false, "per-author line ownership")
cmd.Flags().BoolVar(&f.hotspots, "hotspots", false, "most-changed files (churn score)")
cmd.Flags().BoolVar(&f.contrib, "contrib", false, "contributor leaderboard")
cmd.Flags().StringVar(&f.at, "at", "", "scan repo at date (e.g. 2024-01-01)")
// Perf
cmd.Flags().BoolVar(&f.speedtest, "speedtest", false, "benchmark: 3 scans, report timing")
cmd.Flags().BoolVar(&f.noCache, "no-cache", false, "bypass cache")
cmd.Flags().StringVar(&f.cacheDir, "cache-dir", "", "override cache directory")
cmd.Flags().IntVar(&f.workers, "workers", 0, "goroutine pool size (default: GOMAXPROCS)")
// Meta
cmd.Flags().BoolVarP(&f.verbose, "verbose", "v", false, "debug logging to stderr")
cmd.Flags().BoolVar(&f.dryRun, "dry-run", false, "list files without scanning")
cmd.Flags().StringVarP(&f.configPath, "config", "c", "", "path to .codeye.toml")
}

func flagsToConfig(f *runFlags) *config.Config {
cfg := config.DefaultConfig()
cfg.Branch = f.branch
cfg.Author = f.author
cfg.Since = f.since
cfg.Until = f.until
cfg.PathFilter = f.pathFilter
cfg.Exclude = f.exclude
cfg.Lang = f.lang
cfg.NoVendor = f.noVendor
cfg.NoGenerated = f.noGenerated
cfg.NoTests = f.noTests
cfg.MinLines = f.minLines
cfg.Format = config.NormalizeFormat(f.format)
cfg.Sort = f.sort
cfg.Desc = f.desc
cfg.Top = f.top
cfg.NoColor = f.noColor || os.Getenv("NO_COLOR") != "" || os.Getenv("CODEYE_NO_COLOR") != ""
cfg.NoHeader = f.noHeader
cfg.Wide = f.wide
cfg.Compact = f.compact
cfg.Pct = f.pct
cfg.Progress = f.progress
	// NerdFont: don't override here — RunE applies cmd.Flags().Changed("nf") for explicit flag handling.
cfg.Theme = f.theme
cfg.History = f.history
cfg.HistoryInterval = f.historyInterval
cfg.HistoryLimit = f.historyLimit
cfg.Blame = f.blame
cfg.Hotspots = f.hotspots
cfg.Contrib = f.contrib
cfg.At = f.at
cfg.Speedtest = f.speedtest
cfg.NoCache = f.noCache
if f.cacheDir != "" {
cfg.CacheDir = f.cacheDir
}
if f.workers > 0 {
cfg.Workers = f.workers
}
cfg.Verbose = f.verbose
cfg.DryRun = f.dryRun
return cfg
}

func runScan(dir string, cfg *config.Config) error {
repo, err := git.Discover(dir)
if err != nil {
return fmt.Errorf("%s is not inside a git repository", dir)
}
if cfg.Verbose {
fmt.Fprintf(os.Stderr, "repo: %s\n", repo.Root)
}
switch {
case cfg.Speedtest:
return runSpeedtest(repo, cfg)
case cfg.History:
return runHistory(repo, cfg)
case cfg.Blame:
return runBlameMode(repo, cfg)
case cfg.Hotspots:
return runHotspotsMode(repo, cfg)
default:
return runStandardScan(repo, cfg, os.Stdout)
}
}

func runStandardScan(repo *git.Repo, cfg *config.Config, w *os.File) error {
start := time.Now()
ref := cfg.Branch

if cfg.At != "" {
sha, err := repo.CommitAtDate(cfg.At)
if err != nil {
return err
}
ref = sha
}

treeSHA, err := repo.TreeSHA(ref)
if err != nil {
return fmt.Errorf("resolving tree SHA: %w", err)
}

fh := cache.FlagsHash(
cfg.Format, cfg.Sort, cfg.Branch,
strings.Join(cfg.Exclude, ","),
strings.Join(cfg.Lang, ","),
strconv.FormatBool(cfg.NoVendor),
strconv.FormatBool(cfg.NoGenerated),
strconv.FormatBool(cfg.NoTests),
)

	c := cache.New(cfg.CacheDir)
	var result *scanner.ScanResult

	if !cfg.NoCache && !cfg.DryRun {
		if data, cerr := c.Get(repo.Root, treeSHA, fh); cerr == nil {
var sr scanner.ScanResult
if json.Unmarshal(data, &sr) == nil {
sr.Cached = true
sr.ScanMs = time.Since(start).Milliseconds()
result = &sr
}
}
}

if result == nil {
		files, err := repo.ListFiles(ref, cfg.Exclude)
		if err != nil {
			return fmt.Errorf("listing files: %w", err)
		}
if cfg.Verbose {
fmt.Fprintf(os.Stderr, "%d tracked files found\n", len(files))
}

scanOpts := scanner.ScanOpts{
RepoRoot:    repo.Root,
Ref:         ref,
Exclude:     cfg.Exclude,
Include:     cfg.PathFilter,
NoVendor:    cfg.NoVendor,
NoGenerated: cfg.NoGenerated,
NoTests:     cfg.NoTests,
LangFilter:  cfg.Lang,
MinLines:    cfg.MinLines,
Workers:     cfg.Workers,
DryRun:      cfg.DryRun,
}
files = scanner.FilterFiles(files, scanOpts)

if cfg.DryRun {
for _, f := range files {
fmt.Fprintln(w, f)
}
return nil
}

result, err = scanner.Scan(files, repo.Root, scanOpts)
if err != nil {
return fmt.Errorf("scan failed: %w", err)
}
result.ScanMs = time.Since(start).Milliseconds()
result.TreeSHA = treeSHA
result.Repo = repo.Root
if ref != "" {
result.Ref = ref
} else {
result.Ref = repo.CurrentBranch()
}

// Post-scan lang filter
if len(cfg.Lang) > 0 {
langSet := make(map[string]bool)
for _, l := range cfg.Lang {
langSet[strings.ToLower(l)] = true
}
var filtered []scanner.LangStats
for _, l := range result.Langs {
if langSet[strings.ToLower(l.Name)] {
filtered = append(filtered, l)
}
}
result.Langs = filtered
}

scanner.SortLangs(result.Langs, cfg.Sort, cfg.Desc)

if !cfg.NoCache {
if data, merr := json.Marshal(result); merr == nil {
_ = c.Put(repo.Root, treeSHA, fh, data)
}
}
}

renderer := output.Get(cfg.Format)
opts := output.RenderOpts{
NoColor:  cfg.NoColor,
NoHeader: cfg.NoHeader,
Wide:     cfg.Wide,
Compact:  cfg.Compact,
Pct:      cfg.Pct,
NerdFont: cfg.NerdFont,
Theme:    cfg.Theme,
Top:      cfg.Top,
Sort:     cfg.Sort,
Desc:     cfg.Desc,
}
return renderer.Render(w, result, opts)
}

func runSpeedtest(repo *git.Repo, cfg *config.Config) error {
fmt.Fprintln(os.Stderr, "speedtest: running 3 scans (no cache)...")
cfg.NoCache = true
devNull, _ := os.Open(os.DevNull)
defer devNull.Close()
var times [3]time.Duration
for i := 0; i < 3; i++ {
start := time.Now()
_ = runStandardScan(repo, cfg, devNull)
times[i] = time.Since(start)
fmt.Printf("  run %d: %dms\n", i+1, times[i].Milliseconds())
}
avg := (times[0] + times[1] + times[2]) / 3
fmt.Printf("  avg:   %dms\n", avg.Milliseconds())
return nil
}

func runHistory(repo *git.Repo, cfg *config.Config) error {
fmt.Fprintln(os.Stderr, "analyzing history...")
series, err := history.Analyze(repo, cfg.Branch, cfg.HistoryLimit,
cfg.HistoryInterval, cfg.Since, cfg.Until)
if err != nil {
return err
}
if len(series.Points) == 0 {
fmt.Println("no commits found in range")
return nil
}

vals := make([]int64, len(series.Points))
for i, p := range series.Points {
v := p.CumNet
if v < 0 {
v = 0
}
vals[i] = v
}

fmt.Printf("LoC history · %s buckets · %d data points\n\n",
cfg.HistoryInterval, len(series.Points))
fmt.Printf("start: %-10s   end: %-10s   net: %+d\n",
humanize.Comma(series.Start),
humanize.Comma(series.End),
series.End-series.Start,
)
fmt.Printf("spark: %s\n\n", output.Spark(vals))

show := 15
if len(series.Points) < show {
show = len(series.Points)
}
fmt.Printf("%-12s  %10s  %10s  %+10s  %12s\n",
"Date", "Added", "Removed", "Net", "Cumulative")
fmt.Println(strings.Repeat("─", 58))
for _, p := range series.Points[len(series.Points)-show:] {
fmt.Printf("%-12s  %10s  %10s  %+10d  %12s\n",
p.Date.Format("2006-01-02"),
humanize.Comma(p.Added),
humanize.Comma(p.Deleted),
p.Net,
humanize.Comma(p.CumNet),
)
}
return nil
}

func runBlameMode(repo *git.Repo, cfg *config.Config) error {
	fmt.Fprintln(os.Stderr, "running blame (may take a moment)...")
	files, err := repo.ListFiles(cfg.Branch, cfg.Exclude)
	if err != nil {
		return err
	}
files = scanner.FilterFiles(files, scanner.ScanOpts{
NoVendor: cfg.NoVendor, NoGenerated: cfg.NoGenerated, NoTests: cfg.NoTests,
})

result, err := blame.Analyze(repo, files, cfg.Branch, cfg.Workers)
if err != nil {
return err
}

fmt.Printf("%-40s  %12s  %5s  %6s\n", "Author", "Lines", "%", "Files")
fmt.Println(strings.Repeat("─", 68))
for i, a := range result.Authors {
if cfg.Top > 0 && i >= cfg.Top {
break
}
lbl := a.Email
if len(lbl) > 39 {
lbl = lbl[:36] + "..."
}
bar := strings.Repeat("█", int(a.Pct/2.5))
fmt.Printf("%-40s  %12s  %4.1f%%  %6d  %s\n",
lbl, humanize.Comma(a.Lines), a.Pct, a.Files, bar)
}
fmt.Printf("\ntotal: %s lines across %d authors\n",
humanize.Comma(result.Total), len(result.Authors))
return nil
}

func runHotspotsMode(repo *git.Repo, cfg *config.Config) error {
fmt.Fprintln(os.Stderr, "analyzing hotspots...")
result, err := hotspot.Analyze(repo.Root, cfg.Branch, cfg.HistoryLimit, cfg.Since)
if err != nil {
return err
}

top := 20
if cfg.Top > 0 {
top = cfg.Top
}
if len(result.Files) < top {
top = len(result.Files)
}

fmt.Printf("%-60s  %8s  %10s\n", "File", "Commits", "Score")
fmt.Println(strings.Repeat("─", 82))
for _, f := range result.Files[:top] {
path := f.Path
if len(path) > 58 {
path = "..." + path[len(path)-55:]
}
heat := ""
if f.Churn > 10000 {
heat = " 🔥"
}
fmt.Printf("%-60s  %8d  %10s%s\n",
path, f.Commits, humanize.Comma(f.Churn), heat)
}
return nil
}

func versionCmd() *cobra.Command {
return &cobra.Command{
Use:   "version",
Short: "print version information",
Run: func(cmd *cobra.Command, args []string) {
fmt.Printf("codeye %s\n", version)
fmt.Printf("  commit:  %s\n", commit)
fmt.Printf("  built:   %s\n", buildDate)
fmt.Printf("  runtime: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
},
}
}

func doctorCmd() *cobra.Command {
return &cobra.Command{
Use:   "doctor",
Short: "check environment for issues",
RunE: func(cmd *cobra.Command, args []string) error {
allOK := true
check := func(label, detail string, ok bool) {
if ok {
fmt.Printf("  ✓  %-18s  %s\n", label, detail)
} else {
fmt.Printf("  ✗  %-18s  %s\n", label, detail)
allOK = false
}
}

fmt.Println("codeye doctor")

p, err := exec.LookPath("git")
check("git binary", p, err == nil)

if err == nil {
out, _ := exec.Command("git", "--version").Output()
check("git version", strings.TrimSpace(string(out)), true)
}

_, rerr := git.Discover(".")
check("git repo (cwd)", ".", rerr == nil)

cdir := config.DefaultCacheDir()
check("cache dir", cdir, os.MkdirAll(cdir, 0o755) == nil)

fmt.Println()
if allOK {
fmt.Println("all checks passed ✓")
return nil
}
return fmt.Errorf("some checks failed")
},
}
}

func cacheManageCmd() *cobra.Command {
cmd := &cobra.Command{Use: "cache", Short: "manage the scan cache"}

cmd.AddCommand(&cobra.Command{
Use:   "status",
Short: "cache statistics",
RunE: func(cmd *cobra.Command, args []string) error {
c := cache.New(config.DefaultCacheDir())
n, sz, err := c.Status()
if err != nil {
return err
}
fmt.Printf("dir:     %s\nentries: %d\nsize:    %s\n",
config.DefaultCacheDir(), n, humanize.Bytes(uint64(sz)))
return nil
},
})

cmd.AddCommand(&cobra.Command{
Use:   "clear",
Short: "delete all cached results",
RunE: func(cmd *cobra.Command, args []string) error {
if err := cache.New(config.DefaultCacheDir()).ClearAll(); err != nil {
return err
}
fmt.Println("cache cleared")
return nil
},
})
return cmd
}

func langsCmd() *cobra.Command {
return &cobra.Command{
Use:   "langs",
Short: "list supported languages",
Run: func(cmd *cobra.Command, args []string) {
fmt.Println("codeye language registry — 150+ languages detected by extension, filename, and shebang")
fmt.Println("pass --lang <name> to filter output to a specific language")
},
}
}

func diffCmd() *cobra.Command {
return &cobra.Command{
Use:   "diff <ref1> <ref2>",
Short: "LoC delta between two refs",
Args:  cobra.ExactArgs(2),
RunE: func(cmd *cobra.Command, args []string) error {
repo, err := git.Discover(".")
if err != nil {
return err
}
		scan := func(ref string) (*scanner.ScanResult, error) {
			cfg := config.DefaultConfig()
			cfg.NoCache = true
			files, err := repo.ListFiles(ref, nil)
			if err != nil {
				return nil, err
			}
return scanner.Scan(files, repo.Root, scanner.ScanOpts{
RepoRoot: repo.Root, Ref: ref, Workers: cfg.Workers,
})
}
fmt.Fprintf(os.Stderr, "scanning %s...\n", args[0])
r1, err := scan(args[0])
if err != nil {
return err
}
fmt.Fprintf(os.Stderr, "scanning %s...\n", args[1])
r2, err := scan(args[1])
if err != nil {
return err
}

m1 := make(map[string]int64)
for _, l := range r1.Langs {
m1[l.Name] = l.Lines
}
m2 := make(map[string]int64)
for _, l := range r2.Langs {
m2[l.Name] = l.Lines
}
seen := make(map[string]bool)
for k := range m1 {
seen[k] = true
}
for k := range m2 {
seen[k] = true
}

fmt.Printf("\nLoC delta · %s → %s\n%s\n", args[0], args[1], strings.Repeat("─", 55))
fmt.Printf("%-20s  %10s  %10s  %10s\n", "Language", args[0], args[1], "Δ")
fmt.Println(strings.Repeat("─", 55))
var tb, ta int64
for lang := range seen {
b, a := m1[lang], m2[lang]
tb += b
ta += a
d := a - b
sym := "+"
if d < 0 {
sym = ""
}
fmt.Printf("%-20s  %10s  %10s  %s%s\n",
lang,
humanize.Comma(b),
humanize.Comma(a),
sym, humanize.Comma(d),
)
}
fmt.Println(strings.Repeat("─", 55))
nd := ta - tb
sym := "+"
if nd < 0 {
sym = ""
}
fmt.Printf("%-20s  %10s  %10s  %s%s\n",
"Total", humanize.Comma(tb), humanize.Comma(ta), sym, humanize.Comma(nd))
return nil
},
}
}

func completionCmd(root *cobra.Command) *cobra.Command {
return &cobra.Command{
Use:       "completion [bash|zsh|fish|powershell]",
Short:     "generate shell completion",
ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
Args:      cobra.ExactValidArgs(1),
RunE: func(cmd *cobra.Command, args []string) error {
switch args[0] {
case "bash":
return root.GenBashCompletion(os.Stdout)
case "zsh":
return root.GenZshCompletion(os.Stdout)
case "fish":
return root.GenFishCompletion(os.Stdout, true)
case "powershell":
return root.GenPowerShellCompletionWithDesc(os.Stdout)
}
return nil
},
}
}
