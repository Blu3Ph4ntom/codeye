// Package config handles .codeye.toml and environment variable resolution.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// Config is the resolved configuration for a codeye run.
type Config struct {
	// Scope & filtering
	Branch      string
	Ref         string
	Author      []string
	Since       string
	Until       string
	PathFilter  []string
	Exclude     []string
	Lang        []string
	NoVendor    bool
	NoGenerated bool
	NoTests     bool
	MinLines    int
	Depth       int

	// Output & display
	Format   string
	Sort     string
	Desc     bool
	Top      int
	NoColor  bool
	NoHeader bool
	Wide     bool
	Compact  bool
	Pct      bool
	Progress bool
	NerdFont bool // use Nerd Font terminal glyphs (requires patched font)
	Theme    string

	// Analysis modes
	History         bool
	HistoryInterval string
	HistoryLimit    int
	Blame           bool
	BlameLang       bool
	Hotspots        bool
	Complexity      bool
	Duplication     bool
	DeadCode        bool
	Contrib         bool
	LangTrend       bool
	At              string

	// Performance & caching
	Speedtest bool
	NoCache   bool
	CacheDir  string
	Workers   int
	Profile   string
	Trace     string

	// Meta
	Verbose bool
	DryRun  bool
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Format:          "table",
		Sort:            "lines",
		Top:             0, // 0 = show all
		Theme:           "dark",
		NerdFont:        os.Getenv("CODEYE_NERD_FONTS") == "1" || os.Getenv("NERD_FONTS") == "1", // opt-in; set CODEYE_NERD_FONTS=1 or use --nf flag
		HistoryInterval: "week",
		HistoryLimit:    500,
		Workers:         runtime.GOMAXPROCS(0),
		CacheDir:        DefaultCacheDir(),
	}
}

// DefaultCacheDir returns the OS-appropriate cache directory.
func DefaultCacheDir() string {
	if d := os.Getenv("CODEYE_CACHE_DIR"); d != "" {
		return d
	}
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "codeye")
	}
	if home, err := os.UserHomeDir(); err == nil {
		// macOS: ~/Library/Caches/codeye
		// Linux: ~/.cache/codeye
		// Windows: %LOCALAPPDATA%\codeye\cache
		switch runtime.GOOS {
		case "darwin":
			return filepath.Join(home, "Library", "Caches", "codeye")
		case "windows":
			if local := os.Getenv("LOCALAPPDATA"); local != "" {
				return filepath.Join(local, "codeye", "cache")
			}
			return filepath.Join(home, "AppData", "Local", "codeye", "cache")
		default:
			return filepath.Join(home, ".cache", "codeye")
		}
	}
	return filepath.Join(os.TempDir(), "codeye-cache")
}

// Resolve loads configuration from defaults, optional home/project config files,
// and CODEYE_* environment variables. CLI flags should be applied afterward.
func Resolve(explicitPath, dir string) (*Config, string, error) {
	cfg := DefaultConfig()
	v := viper.New()
	v.SetConfigType("toml")
	v.SetEnvPrefix("CODEYE")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	if homePath, ok := userConfigPath(); ok {
		if err := mergeConfigFile(v, homePath, false); err != nil {
			return nil, "", err
		}
	}

	configPath := explicitPath
	if configPath == "" {
		configPath = os.Getenv("CODEYE_CONFIG")
	}
	if configPath == "" {
		configPath = findProjectConfig(dir)
	}
	if err := mergeConfigFile(v, configPath, explicitPath != "" || os.Getenv("CODEYE_CONFIG") != ""); err != nil {
		return nil, "", err
	}

	applyResolvedValues(cfg, v)
	if os.Getenv("NO_COLOR") != "" {
		cfg.NoColor = true
	}

	return cfg, configPath, nil
}

func userConfigPath() (string, bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", false
	}
	path := filepath.Join(home, ".codeye.toml")
	if _, err := os.Stat(path); err == nil {
		return path, true
	}
	return "", false
}

func findProjectConfig(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	for {
		candidate := filepath.Join(abs, ".codeye.toml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return ""
		}
		abs = parent
	}
}

func mergeConfigFile(v *viper.Viper, path string, required bool) error {
	if path == "" {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		if required {
			return fmt.Errorf("loading config %q: %w", path, err)
		}
		return nil
	}
	defer f.Close()

	if err := v.MergeConfig(f); err != nil {
		return fmt.Errorf("parsing config %q: %w", path, err)
	}
	return nil
}

func applyResolvedValues(cfg *Config, v *viper.Viper) {
	cfg.Branch = stringValue(v, "branch", cfg.Branch)
	cfg.Ref = stringValue(v, "ref", cfg.Ref)
	cfg.Author = csvListValue(v, "author", cfg.Author)
	cfg.Since = stringValue(v, "since", cfg.Since)
	cfg.Until = stringValue(v, "until", cfg.Until)
	cfg.PathFilter = csvListValue(v, "path_filter", cfg.PathFilter)
	cfg.Exclude = csvListValue(v, "exclude", cfg.Exclude)
	cfg.Lang = csvListValue(v, "lang", cfg.Lang)
	cfg.NoVendor = boolValue(v, "no_vendor", cfg.NoVendor)
	cfg.NoGenerated = boolValue(v, "no_generated", cfg.NoGenerated)
	cfg.NoTests = boolValue(v, "no_tests", cfg.NoTests)
	cfg.MinLines = intValue(v, "min_lines", cfg.MinLines)
	cfg.Depth = intValue(v, "depth", cfg.Depth)
	cfg.Format = NormalizeFormat(stringValue(v, "format", cfg.Format))
	cfg.Sort = stringValue(v, "sort", cfg.Sort)
	cfg.Desc = boolValue(v, "desc", cfg.Desc)
	cfg.Top = intValue(v, "top", cfg.Top)
	cfg.NoColor = boolValue(v, "no_color", cfg.NoColor)
	cfg.NoHeader = boolValue(v, "no_header", cfg.NoHeader)
	cfg.Wide = boolValue(v, "wide", cfg.Wide)
	cfg.Compact = boolValue(v, "compact", cfg.Compact)
	cfg.Pct = boolValue(v, "pct", cfg.Pct)
	cfg.Progress = boolValue(v, "progress", cfg.Progress)
	cfg.NerdFont = boolValue(v, "nerd_font", cfg.NerdFont) || boolValue(v, "nerd_fonts", cfg.NerdFont)
	cfg.Theme = stringValue(v, "theme", cfg.Theme)
	cfg.History = boolValue(v, "history", cfg.History)
	cfg.HistoryInterval = stringValue(v, "history_interval", cfg.HistoryInterval)
	cfg.HistoryLimit = intValue(v, "history_limit", cfg.HistoryLimit)
	cfg.Blame = boolValue(v, "blame", cfg.Blame)
	cfg.Hotspots = boolValue(v, "hotspots", cfg.Hotspots)
	cfg.Contrib = boolValue(v, "contrib", cfg.Contrib)
	cfg.At = stringValue(v, "at", cfg.At)
	cfg.Speedtest = boolValue(v, "speedtest", cfg.Speedtest)
	cfg.NoCache = boolValue(v, "no_cache", cfg.NoCache)
	cfg.CacheDir = stringValue(v, "cache_dir", cfg.CacheDir)
	cfg.Workers = intValue(v, "workers", cfg.Workers)
	cfg.Verbose = boolValue(v, "verbose", cfg.Verbose)
	cfg.DryRun = boolValue(v, "dry_run", cfg.DryRun)
}

func stringValue(v *viper.Viper, key, fallback string) string {
	if !v.IsSet(key) {
		return fallback
	}
	value := strings.TrimSpace(v.GetString(key))
	if value == "" {
		return fallback
	}
	return value
}

func boolValue(v *viper.Viper, key string, fallback bool) bool {
	if !v.IsSet(key) {
		return fallback
	}
	raw := strings.TrimSpace(v.GetString(key))
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return parsed
}

func intValue(v *viper.Viper, key string, fallback int) int {
	if !v.IsSet(key) {
		return fallback
	}
	raw := strings.TrimSpace(v.GetString(key))
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return parsed
}

func csvListValue(v *viper.Viper, key string, fallback []string) []string {
	if !v.IsSet(key) {
		return fallback
	}
	items := normalizeList(v.GetStringSlice(key))
	if len(items) > 0 {
		return items
	}
	return normalizeList([]string{v.GetString(key)})
}

func normalizeList(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}

// VendorDirs is the list of directories auto-excluded by --no-vendor.
var VendorDirs = []string{
	"vendor/",
	"node_modules/",
	"bower_components/",
	".yarn/",
	"Pods/",
	"Carthage/",
	"__pycache__/",
	".venv/",
	"venv/",
	"env/",
	".env/",
	"target/", // Rust/Maven
	"build/",
	"dist/",
	"out/",
	".gradle/",
	".mvn/",
}

// GeneratedPatterns is the list of file patterns auto-excluded by --no-generated.
var GeneratedPatterns = []string{
	"*.pb.go",
	"*.pb.gw.go",
	"*.gen.go",
	"*_generated.go",
	"*.min.js",
	"*.min.css",
	"*-lock.json",
	"*.lock",
	"go.sum",
}

// TestPatterns is the list of file patterns auto-excluded by --no-tests.
var TestPatterns = []string{
	"*_test.go",
	"*.spec.ts",
	"*.spec.js",
	"*.test.ts",
	"*.test.js",
	"*.spec.py",
	"*_test.py",
	"*_spec.rb",
	"test_*.py",
}

// NormalizeFormat lowercases and validates the format string.
func NormalizeFormat(f string) string {
	return strings.ToLower(strings.TrimSpace(f))
}
