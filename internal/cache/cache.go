// Package cache provides a content-addressable cache for scan results.
// Keys are {repo_hash}/{tree_sha} — stable across runs if the tree is unchanged.
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ErrMiss is returned on a cache miss.
var ErrMiss = errors.New("cache miss")

// Entry is the cached result stored on disk.
type Entry struct {
	Version  int             `json:"v"`
	CachedAt time.Time       `json:"cached_at"`
	Data     json.RawMessage `json:"data"`
}

const currentVersion = 2

// Cache manages a directory of cached scan results.
type Cache struct {
	Dir string
}

// New creates a Cache backed by dir. The directory is created if it doesn't exist.
func New(dir string) *Cache {
	return &Cache{Dir: dir}
}

// Get retrieves a cached entry for the given repo root and tree SHA.
// Returns ErrMiss if not found or stale.
func (c *Cache) Get(repoRoot, treeSHA, flagsHash string) ([]byte, error) {
	path := c.entryPath(repoRoot, treeSHA, flagsHash)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, ErrMiss
	}
	if err != nil {
		return nil, fmt.Errorf("cache read: %w", err)
	}

	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, ErrMiss // corrupt entry — treat as miss
	}
	if entry.Version != currentVersion {
		return nil, ErrMiss
	}
	return entry.Data, nil
}

// Put stores a result in the cache.
func (c *Cache) Put(repoRoot, treeSHA, flagsHash string, result []byte) error {
	path := c.entryPath(repoRoot, treeSHA, flagsHash)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("cache mkdir: %w", err)
	}

	entry := Entry{
		Version:  currentVersion,
		CachedAt: time.Now().UTC(),
		Data:     json.RawMessage(result),
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("cache marshal: %w", err)
	}

	// Write atomically via temp file
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("cache write: %w", err)
	}
	return os.Rename(tmp, path)
}

// Clear removes all cache entries for a given repo root.
func (c *Cache) Clear(repoRoot string) error {
	dir := filepath.Join(c.Dir, repoHash(repoRoot))
	if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cache clear: %w", err)
	}
	return nil
}

// ClearAll removes the entire cache directory.
func (c *Cache) ClearAll() error {
	return os.RemoveAll(c.Dir)
}

// Status returns statistics about the cache.
func (c *Cache) Status() (entries int, sizeBytes int64, err error) {
	err = filepath.Walk(c.Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			entries++
			sizeBytes += info.Size()
		}
		return nil
	})
	return
}

func (c *Cache) entryPath(repoRoot, treeSHA, flagsHash string) string {
	rh := repoHash(repoRoot)
	key := treeSHA
	if flagsHash != "" {
		key += "-" + flagsHash
	}
	return filepath.Join(c.Dir, rh, key+".json")
}

func repoHash(root string) string {
	h := sha256.Sum256([]byte(root))
	return hex.EncodeToString(h[:8])
}

// FlagsHash produces a short hash of the scan options that affect output.
func FlagsHash(flags ...string) string {
	h := sha256.New()
	for _, f := range flags {
		h.Write([]byte(f))
		h.Write([]byte{0})
	}
	sum := h.Sum(nil)
	return hex.EncodeToString(sum[:4])
}
