// Package git provides wrappers around git porcelain commands.
// All operations run git as a subprocess; no CGO, no libgit2.
package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repo represents a git repository.
type Repo struct {
	Root   string // absolute path to the repo root
	GitDir string // path to .git dir or file
}

// Discover finds the git repository root starting from dir.
// Equivalent to `git rev-parse --show-toplevel`.
func Discover(dir string) (*Repo, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("abs path: %w", err)
	}

	out, err := runGit(abs, "rev-parse", "--show-toplevel", "--absolute-git-dir")
	if err != nil {
		return nil, fmt.Errorf("not a git repository: %w", err)
	}

	parts := strings.SplitN(strings.TrimSpace(out), "\n", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("unexpected git output: %q", out)
	}

	return &Repo{
		Root:   strings.TrimSpace(parts[0]),
		GitDir: strings.TrimSpace(parts[1]),
	}, nil
}

// HEAD returns the current HEAD commit SHA.
func (r *Repo) HEAD() (string, error) {
	out, err := runGit(r.Root, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("rev-parse HEAD: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// TreeSHA returns the tree SHA for a given ref (or HEAD if ref is empty).
func (r *Repo) TreeSHA(ref string) (string, error) {
	if ref == "" {
		ref = "HEAD"
	}
	out, err := runGit(r.Root, "rev-parse", ref+"^{tree}")
	if err != nil {
		return "", fmt.Errorf("tree SHA for %s: %w", ref, err)
	}
	return strings.TrimSpace(out), nil
}

// CurrentBranch returns the current branch name, or "HEAD" if detached.
func (r *Repo) CurrentBranch() string {
	out, err := runGit(r.Root, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "HEAD"
	}
	return strings.TrimSpace(out)
}

// LastCommitTime returns the relative time of the last commit (e.g. "3 hours ago").
func (r *Repo) LastCommitTime() string {
	out, err := runGit(r.Root, "log", "-1", "--format=%ar")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// ListFiles returns the list of git-tracked files for the given ref.
// If ref is empty, uses the index (working tree tracked files).
// Uses -z for null-byte separation to handle filenames with spaces/newlines.
func (r *Repo) ListFiles(ref string, extraArgs ...string) ([]string, error) {
	args := []string{"ls-files", "-z"}
	if ref != "" {
		args = append(args, "--with-tree="+ref)
	}
	args = append(args, extraArgs...)

	out, err := runGitRaw(r.Root, args...)
	if err != nil {
		return nil, fmt.Errorf("ls-files: %w", err)
	}

	if len(out) == 0 {
		return nil, nil
	}

	// Split on null bytes
	parts := bytes.Split(out, []byte{0})
	files := make([]string, 0, len(parts))
	for _, p := range parts {
		s := string(p)
		if s != "" {
			files = append(files, s)
		}
	}
	return files, nil
}

// LogEntry is a single entry from git log --numstat.
type LogEntry struct {
	Hash    string
	Author  string
	Email   string
	Date    string // ISO 8601
	Subject string
	Added   int64
	Deleted int64
}

// Log returns log entries with numstat for the given ref range.
func (r *Repo) Log(ref string, limit int, since, until string) ([]LogEntry, error) {
	args := []string{
		"log",
		"--numstat",
		"--pretty=format:COMMIT %H %ae %aI %s",
	}
	if limit > 0 {
		args = append(args, fmt.Sprintf("-n%d", limit))
	}
	if since != "" {
		args = append(args, "--since="+since)
	}
	if until != "" {
		args = append(args, "--until="+until)
	}
	if ref != "" {
		args = append(args, ref)
	}

	out, err := runGit(r.Root, args...)
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	return parseNumstat(out), nil
}

// Shortlog returns per-author commit and line counts.
func (r *Repo) Shortlog(ref string) ([]AuthorStat, error) {
	args := []string{"shortlog", "-sne"}
	if ref != "" {
		args = append(args, ref)
	}
	out, err := runGit(r.Root, args...)
	if err != nil {
		return nil, fmt.Errorf("shortlog: %w", err)
	}
	return parseShortlog(out), nil
}

// BlameLines runs git blame --porcelain for a file and returns per-author line counts.
func (r *Repo) BlameLines(file, ref string) (map[string]int64, error) {
	args := []string{"blame", "--porcelain"}
	if ref != "" {
		args = append(args, ref)
	}
	args = append(args, "--", file)

	out, err := runGit(r.Root, args...)
	if err != nil {
		return nil, nil // file may not exist at ref
	}

	return parseBlame(out), nil
}

// RevParse resolves a ref to a full commit SHA.
func (r *Repo) RevParse(ref string) (string, error) {
	out, err := runGit(r.Root, "rev-parse", "--verify", ref)
	if err != nil {
		return "", fmt.Errorf("rev-parse %s: %w", ref, err)
	}
	return strings.TrimSpace(out), nil
}

// CommitAtDate finds the closest commit SHA at or before the given date.
func (r *Repo) CommitAtDate(date string) (string, error) {
	out, err := runGit(r.Root, "rev-list", "-n1", "--before="+date, "HEAD")
	if err != nil {
		return "", fmt.Errorf("commit at %s: %w", date, err)
	}
	sha := strings.TrimSpace(out)
	if sha == "" {
		return "", fmt.Errorf("no commit found before %s", date)
	}
	return sha, nil
}

// Branches returns the list of local branch names.
func (r *Repo) Branches() ([]string, error) {
	out, err := runGit(r.Root, "branch", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}
	var branches []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			branches = append(branches, line)
		}
	}
	return branches, nil
}

// IsClean returns true if the working tree and index have no uncommitted changes.
func (r *Repo) IsClean() (bool, error) {
	out, err := runGit(r.Root, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) == "", nil
}

// runGit runs a git command and returns stdout as string.
func runGit(dir string, args ...string) (string, error) {
	out, err := runGitRaw(dir, args...)
	return string(out), err
}

// runGitRaw runs a git command and returns stdout as bytes.
func runGitRaw(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_OPTIONAL_LOCKS=0",
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, stderr.String())
	}
	return out, nil
}

// AuthorStat holds shortlog stats for one author.
type AuthorStat struct {
	Name    string
	Email   string
	Commits int64
}

func parseShortlog(out string) []AuthorStat {
	var stats []AuthorStat
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var commits int64
		var rest string
		fmt.Sscanf(line, "%d\t%[^\n]", &commits, &rest)
		// rest is "Name <email>"
		name, email := parseNameEmail(rest)
		stats = append(stats, AuthorStat{Name: name, Email: email, Commits: commits})
	}
	return stats
}

func parseNameEmail(s string) (name, email string) {
	s = strings.TrimSpace(s)
	start := strings.LastIndex(s, "<")
	end := strings.LastIndex(s, ">")
	if start >= 0 && end > start {
		name = strings.TrimSpace(s[:start])
		email = s[start+1 : end]
		return
	}
	return s, ""
}

func parseNumstat(out string) []LogEntry {
	var entries []LogEntry
	var current *LogEntry

	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "COMMIT ") {
			if current != nil {
				entries = append(entries, *current)
			}
			parts := strings.SplitN(line[7:], " ", 4)
			current = &LogEntry{}
			if len(parts) >= 1 {
				current.Hash = parts[0]
			}
			if len(parts) >= 2 {
				// May have author name+ email
				nameEmail := parts[1]
				current.Author, current.Email = parseNameEmail(nameEmail)
				if current.Author == "" {
					current.Email = nameEmail
				}
			}
			if len(parts) >= 3 {
				current.Date = parts[2]
			}
			if len(parts) >= 4 {
				current.Subject = parts[3]
			}
			continue
		}
		if current == nil {
			continue
		}
		// numstat lines: "added\tdeleted\tfile"
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) == 3 {
			var added, deleted int64
			fmt.Sscanf(parts[0], "%d", &added)
			fmt.Sscanf(parts[1], "%d", &deleted)
			current.Added += added
			current.Deleted += deleted
		}
	}
	if current != nil {
		entries = append(entries, *current)
	}
	return entries
}

func parseBlame(out string) map[string]int64 {
	counts := make(map[string]int64)
	emailMap := make(map[string]string) // hash -> email

	lines := strings.Split(out, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			continue
		}

		// A line starting with a tab is a code line.
		if strings.HasPrefix(line, "\t") {
			// This shouldn't happen before we see a commit info line,
			// but we'll handle it by just skipping.
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 4 && len(parts[0]) >= 7 {
			// This is likely a commit info line: <sha> <src> <dst> <n>
			hash := parts[0]
			email, ok := emailMap[hash]

			// Scan forward for metadata (author-mail) and then the tabbed code line
			for j := i + 1; j < len(lines); j++ {
				inner := lines[j]
				if strings.HasPrefix(inner, "author-mail ") && !ok {
					email = strings.TrimPrefix(inner, "author-mail ")
					email = strings.TrimSpace(strings.Trim(email, "<>"))
					emailMap[hash] = email
					ok = true
				} else if strings.HasPrefix(inner, "\t") {
					// Found the code line for this commit block
					if ok && email != "" {
						counts[email]++
					}
					i = j // Advance main loop
					break
				} else if len(inner) > 40 && !strings.Contains(inner[:8], " ") {
					// Hit next commit block without finding tab? (Shouldn't happen in porcelain)
					i = j - 1
					break
				}
			}
		}
	}
	return counts
}
