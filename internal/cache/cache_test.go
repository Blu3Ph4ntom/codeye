package cache_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codeye/codeye/internal/cache"
)

func TestCacheRoundtrip(t *testing.T) {
	dir := t.TempDir()
	c := cache.New(filepath.Join(dir, "cache"))

	data := []byte(`{"test":true}`)
	const repo = "/fake/repo"
	const tree = "abc123def456"

	// Cache miss
	_, err := c.Get(repo, tree, "")
	if err != cache.ErrMiss {
		t.Fatalf("expected ErrMiss, got %v", err)
	}

	// Put
	if err := c.Put(repo, tree, "", data); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Cache hit
	got, err := c.Get(repo, tree, "")
	if err != nil {
		t.Fatalf("Get after Put failed: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("got %q, want %q", got, data)
	}
}

func TestCacheWithFlagsHash(t *testing.T) {
	dir := t.TempDir()
	c := cache.New(filepath.Join(dir, "cache"))

	data1 := []byte(`{"format":"table"}`)
	data2 := []byte(`{"format":"json"}`)

	h1 := cache.FlagsHash("table")
	h2 := cache.FlagsHash("json")

	if err := c.Put("/repo", "sha1", h1, data1); err != nil {
		t.Fatal(err)
	}
	if err := c.Put("/repo", "sha1", h2, data2); err != nil {
		t.Fatal(err)
	}

	got1, _ := c.Get("/repo", "sha1", h1)
	got2, _ := c.Get("/repo", "sha1", h2)

	if string(got1) != string(data1) {
		t.Errorf("flags hash collision: got %q, want %q", got1, data1)
	}
	if string(got2) != string(data2) {
		t.Errorf("flags hash collision: got %q, want %q", got2, data2)
	}
}

func TestCacheClear(t *testing.T) {
	dir := t.TempDir()
	c := cache.New(filepath.Join(dir, "cache"))

	data := []byte(`{}`)
	if err := c.Put("/repo", "sha", "", data); err != nil {
		t.Fatal(err)
	}

	if err := c.Clear("/repo"); err != nil {
		t.Fatal(err)
	}

	_, err := c.Get("/repo", "sha", "")
	if err != cache.ErrMiss {
		t.Errorf("expected ErrMiss after clear, got %v", err)
	}
}

func TestCacheStatus(t *testing.T) {
	dir := t.TempDir()
	c := cache.New(filepath.Join(dir, "cache"))

	for i := 0; i < 5; i++ {
		if err := c.Put("/repo", string(rune('a'+i)), "", []byte(`{}`)); err != nil {
			t.Fatal(err)
		}
	}

	entries, size, err := c.Status()
	if err != nil {
		t.Fatal(err)
	}
	if entries != 5 {
		t.Errorf("expected 5 entries, got %d", entries)
	}
	if size == 0 {
		t.Error("expected non-zero size")
	}
	_ = os.Remove(dir)
}

func TestFlagsHash(t *testing.T) {
	h1 := cache.FlagsHash("a", "b", "c")
	h2 := cache.FlagsHash("a", "b", "c")
	h3 := cache.FlagsHash("a", "b", "d")

	if h1 != h2 {
		t.Error("same inputs should produce same hash")
	}
	if h1 == h3 {
		t.Error("different inputs should produce different hash")
	}
}
