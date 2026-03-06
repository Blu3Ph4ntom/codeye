// Package history provides LoC growth analysis via git log --numstat.
package history

import (
	"fmt"
	"sort"
	"time"

	"github.com/blu3ph4ntom/codeye/internal/git"
)

// DataPoint is a single point in time with LoC added/removed.
type DataPoint struct {
	Date    time.Time
	Added   int64
	Deleted int64
	Net     int64
	CumNet  int64 // cumulative
}

// Series is a time-ordered sequence of LoC data points.
type Series struct {
	Points []DataPoint
	Peak   int64
	Start  int64
	End    int64
}

// Analyze walks git log --numstat and builds a LoC growth time series.
func Analyze(repo *git.Repo, ref string, limit int, interval string, since, until string) (*Series, error) {
	entries, err := repo.Log(ref, limit, since, until)
	if err != nil {
		return nil, err
	}

	// Bucket entries by interval
	buckets := make(map[string]*DataPoint)
	var bucketOrder []string

	for _, e := range entries {
		t, err := time.Parse(time.RFC3339, e.Date)
		if err != nil {
			continue
		}

		key := bucketKey(t, interval)
		if _, ok := buckets[key]; !ok {
			buckets[key] = &DataPoint{Date: bucketDate(t, interval)}
			bucketOrder = append(bucketOrder, key)
		}
		buckets[key].Added += e.Added
		buckets[key].Deleted += e.Deleted
		buckets[key].Net += e.Added - e.Deleted
	}

	// Sort by date
	sort.Strings(bucketOrder)

	points := make([]DataPoint, 0, len(bucketOrder))
	var cum int64
	for _, key := range bucketOrder {
		dp := buckets[key]
		cum += dp.Net
		dp.CumNet = cum
		points = append(points, *dp)
	}

	s := &Series{Points: points}
	if len(points) > 0 {
		s.Start = points[0].CumNet
		s.End = points[len(points)-1].CumNet
		for _, p := range points {
			if p.CumNet > s.Peak {
				s.Peak = p.CumNet
			}
		}
	}
	return s, nil
}

func bucketKey(t time.Time, interval string) string {
	switch interval {
	case "day":
		return t.Format("2006-01-02")
	case "month":
		return t.Format("2006-01")
	case "quarter":
		y, m, _ := t.Date()
		q := (int(m)-1)/3 + 1
		return fmt.Sprintf("%d-Q%d", y, q)
	case "year":
		return t.Format("2006")
	default: // week
		y, w := t.ISOWeek()
		return fmt.Sprintf("%04d-W%02d", y, w)
	}
}

func bucketDate(t time.Time, interval string) time.Time {
	switch interval {
	case "day":
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	case "month":
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	case "year":
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	default: // week
		// Start of ISO week (Monday)
		weekday := int(t.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := t.AddDate(0, 0, -(weekday - 1))
		return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)
	}
}
