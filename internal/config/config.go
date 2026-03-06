// Package config handles .codeye.toml and environment variable resolution.
package config

import (
	"os"
	"path/filepath"
	"runtime"
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
	Progress  bool
	Emoji     bool
	NerdFont  bool // use Nerd Font terminal glyphs (requires patched font)
	Theme     string

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
	Speedtest  bool
	NoCache    bool
	CacheDir   string
	Workers    int
	Profile    string
	Trace      string

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
		NerdFont:        os.Getenv("CODEYE_NERD_FONTS") == "1" || os.Getenv("NERD_FONTS") == "1",
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

// LoadFile loads config from a .codeye.toml file if present.
func LoadFile(path string) error {
	viper.SetConfigFile(path)
	viper.SetConfigType("toml")
	return viper.ReadInConfig()
}

// LoadAuto searches for .codeye.toml starting from dir, walking up to filesystem root.
func LoadAuto(dir string) (string, error) {
	for {
		candidate := filepath.Join(dir, ".codeye.toml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, LoadFile(candidate)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", nil
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
	"target/",       // Rust/Maven
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
