package indexer

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var skipDirNames = map[string]bool{
	".git":       true,
	"vendor":     true,
	".reponerve": true,
	"bin":        true,
	"node_modules": true,
}

func listGoFiles(repoPath string) ([]string, error) {
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

		if !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}

		if isGeneratedFile(path) {
			return nil
		}

		rel, err := filepath.Rel(repoPath, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}

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
