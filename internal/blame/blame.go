// Package blame provides per-author LoC ownership via git blame.
package blame

import (
	"sort"
	"sync"

	"github.com/blu3ph4ntom/codeye/internal/git"
)

// AuthorOwnership holds line ownership stats for one author.
type AuthorOwnership struct {
	Email string
	Name  string
	Lines int64
	Files int
	Pct   float64
}

// Result is the aggregate blame result for a repo.
type Result struct {
	Authors []AuthorOwnership
	Total   int64
}

// Analyze runs git blame on all provided files in parallel
// and aggregates line ownership per author email.
func Analyze(repo *git.Repo, files []string, ref string, workers int) (*Result, error) {
	if workers <= 0 {
		workers = 8
	}

	type job struct{ file string }
	type blameResult struct {
		counts map[string]int64
		file   string
	}

	jobs := make(chan job, workers*2)
	results := make(chan blameResult, workers*2)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				counts, _ := repo.BlameLines(j.file, ref)
				results <- blameResult{counts: counts, file: j.file}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	go func() {
		for _, f := range files {
			jobs <- job{file: f}
		}
		close(jobs)
	}()

	// Aggregate
	emailLines := make(map[string]int64)
	emailFiles := make(map[string]int)
	for res := range results {
		for email, count := range res.counts {
			emailLines[email] += count
			emailFiles[email]++
		}
	}

	var total int64
	for _, v := range emailLines {
		total += v
	}

	authors := make([]AuthorOwnership, 0, len(emailLines))
	for email, lines := range emailLines {
		pct := 0.0
		if total > 0 {
			pct = float64(lines) / float64(total) * 100
		}
		authors = append(authors, AuthorOwnership{
			Email: email,
			Lines: lines,
			Files: emailFiles[email],
			Pct:   pct,
		})
	}

	sort.Slice(authors, func(i, j int) bool {
		return authors[i].Lines > authors[j].Lines
	})

	return &Result{Authors: authors, Total: total}, nil
}
