// Package hotspot identifies high-churn files in a repository.
package hotspot

import (
"bytes"
"fmt"
"os"
"os/exec"
"sort"
"strings"
)

// FileHotspot holds churn data for a single file.
type FileHotspot struct {
Path    string
Commits int64
Added   int64
Deleted int64
Churn   int64 // score = commits × (added + deleted)
}

// Result holds the hotspot analysis results.
type Result struct {
Files []FileHotspot
}

// Analyze computes churn scores for all files in the repo.
func Analyze(repoRoot, ref string, limit int, since string) (*Result, error) {
args := []string{
"log", "--numstat",
"--pretty=format:COMMIT %H",
}
if limit > 0 {
args = append(args, fmt.Sprintf("-n%d", limit))
}
if since != "" {
args = append(args, "--since="+since)
}
if ref != "" {
args = append(args, ref)
}

cmd := exec.Command("git", args...)
cmd.Dir = repoRoot
cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_OPTIONAL_LOCKS=0")
out, err := cmd.Output()
if err != nil {
return nil, fmt.Errorf("git log: %w", err)
}

type fileStats struct {
commits int64
added   int64
deleted int64
}
files := make(map[string]*fileStats)
inCommit := false

for _, rawLine := range bytes.Split(out, []byte("\n")) {
line := string(rawLine)
if len(line) == 0 {
continue
}
if strings.HasPrefix(line, "COMMIT ") {
inCommit = true
continue
}
if !inCommit {
continue
}
parts := strings.SplitN(line, "\t", 3)
if len(parts) < 3 {
continue
}
var added, deleted int64
fmt.Sscanf(parts[0], "%d", &added)
fmt.Sscanf(parts[1], "%d", &deleted)
fname := strings.TrimSpace(parts[2])
if fname == "" {
continue
}
if _, ok := files[fname]; !ok {
files[fname] = &fileStats{}
}
files[fname].commits++
files[fname].added += added
files[fname].deleted += deleted
}

hotspots := make([]FileHotspot, 0, len(files))
for path, s := range files {
churn := s.commits * (s.added + s.deleted)
hotspots = append(hotspots, FileHotspot{
Path:    path,
Commits: s.commits,
Added:   s.added,
Deleted: s.deleted,
Churn:   churn,
})
}

sort.Slice(hotspots, func(i, j int) bool {
return hotspots[i].Churn > hotspots[j].Churn
})

return &Result{Files: hotspots}, nil
}
