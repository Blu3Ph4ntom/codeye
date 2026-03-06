package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/codeye/codeye/internal/git"
)

func setupRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Logf("cmd %v: %s", args, out)
		}
	}
	run("git", "init", "-b", "main")
	run("git", "config", "user.email", "test@test.com")
	run("git", "config", "user.name", "Tester")
	err := os.WriteFile(filepath.Join(dir, "hello.go"), []byte("package main\n"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	run("git", "add", ".")
	run("git", "commit", "-m", "init")
	return dir
}

func TestDiscover(t *testing.T) {
	dir := setupRepo(t)
	repo, err := git.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if repo.Root == "" {
		t.Error("repo root is empty")
	}
}

func TestDiscoverFromSubdir(t *testing.T) {
	dir := setupRepo(t)
	sub := filepath.Join(dir, "a", "b", "c")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	repo, err := git.Discover(sub)
	if err != nil {
		t.Fatalf("Discover from subdir: %v", err)
	}
	if repo.Root == "" {
		t.Error("repo root is empty")
	}
}

func TestDiscoverNotRepo(t *testing.T) {
	dir := t.TempDir()
	_, err := git.Discover(dir)
	if err == nil {
		t.Error("expected error for non-repo dir")
	}
}

func TestHead(t *testing.T) {
	dir := setupRepo(t)
	repo, _ := git.Discover(dir)
	sha, err := repo.HEAD()
	if err != nil {
		t.Fatalf("HEAD: %v", err)
	}
	if len(sha) != 40 {
		t.Errorf("HEAD SHA length %d, want 40: %q", len(sha), sha)
	}
}

func TestTreeSHA(t *testing.T) {
	dir := setupRepo(t)
	repo, _ := git.Discover(dir)
	sha, err := repo.TreeSHA("")
	if err != nil {
		t.Fatalf("TreeSHA: %v", err)
	}
	if len(sha) != 40 {
		t.Errorf("tree SHA length %d, want 40", len(sha))
	}
}

func TestListFiles(t *testing.T) {
	dir := setupRepo(t)
	repo, _ := git.Discover(dir)

	files, err := repo.ListFiles("", nil)
	if err != nil {
		t.Fatalf("ListFiles: %v", err)
	}
	if len(files) == 0 {
		t.Error("expected at least one file")
	}
	found := false
	for _, f := range files {
		if f == "hello.go" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("hello.go not in file list: %v", files)
	}
}

func TestCurrentBranch(t *testing.T) {
	dir := setupRepo(t)
	repo, _ := git.Discover(dir)
	branch := repo.CurrentBranch()
	if branch == "" {
		t.Error("empty branch name")
	}
}

func TestIsClean(t *testing.T) {
	dir := setupRepo(t)
	repo, _ := git.Discover(dir)

	clean, err := repo.IsClean()
	if err != nil {
		t.Fatal(err)
	}
	if !clean {
		t.Error("expected clean repo after commit")
	}

	// make it dirty
	_ = os.WriteFile(filepath.Join(dir, "new.go"), []byte("package x\n"), 0o644)
	clean, err = repo.IsClean()
	if err != nil {
		t.Fatal(err)
	}
	if clean {
		t.Error("expected dirty after writing untracked file")
	}
}
