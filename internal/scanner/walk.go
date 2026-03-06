package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

// WalkDir finds all files in a directory that should be scanned.
// It is used as a fallback when git is not available or the dir is not a repo.
func WalkDir(root string, exclude []string) ([]string, error) {
	var files []string
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip hidden dirs and common noise
			if rel != "." && (strings.HasPrefix(info.Name(), ".") || info.Name() == "node_modules" || info.Name() == "vendor") {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		files = append(files, rel)
		return nil
	})

	return files, err
}
