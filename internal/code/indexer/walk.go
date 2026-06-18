package indexer

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/reponerve/reponerve/internal/code/lang"
)

var skipDirNames = map[string]bool{
	".git":         true,
	"vendor":       true,
	".reponerve":   true,
	"bin":          true,
	"node_modules": true,
	// Common build output directories — contain generated/minified files that
	// produce false symbol collisions and add no indexing value.
	"build":       true,
	"dist":        true,
	"out":         true,
	"target":      true,
	".next":       true,
	".nuxt":       true,
	".output":     true,
	".cache":      true,
	"__pycache__": true,
	"_build":      true,
	"coverage":    true,
	// Agent/tooling directories — not project source; pollute plan/ask context.
	".agents":       true,
	"graphify-venv": true,
}

func listGoFiles(repoPath string) ([]string, error) {
	return listFilesByPredicate(repoPath, func(rel string) bool {
		return lang.Detect(rel) == lang.Go && lang.IsIndexable(rel)
	})
}

func listMultiLangFiles(repoPath string) ([]string, error) {
	return listFilesByPredicate(repoPath, func(rel string) bool {
		detected := lang.Detect(rel)
		return detected != "" && detected != lang.Go && lang.IsIndexable(rel)
	})
}

func listAllIndexableFiles(repoPath string) ([]string, error) {
	return listFilesByPredicate(repoPath, lang.IsIndexable)
}

func listFilesByPredicate(repoPath string, keep func(rel string) bool) ([]string, error) {
	repoPath = filepath.Clean(repoPath)
	var files []string

	err := filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if skipDirNames[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(repoPath, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !keep(rel) {
			return nil
		}
		if lang.Detect(rel) == lang.Go && isGeneratedFile(path) {
			return nil
		}
		files = append(files, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func isGeneratedFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for i := 0; i < 5 && scanner.Scan(); i++ {
		if strings.Contains(scanner.Text(), "Code generated") {
			return true
		}
	}
	return false
}
