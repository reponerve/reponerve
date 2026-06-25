package indexer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// DiscoverModuleRoots returns Go module roots for a repository path.
func DiscoverModuleRoots(repoPath string) ([]moduleRoot, error) {
	return discoverModuleRoots(repoPath)
}

// ModulePathsForFiles maps changed file paths to Go module paths.
func ModulePathsForFiles(repoPath string, files []string) ([]string, error) {
	repoPath = filepath.Clean(repoPath)
	seen := map[string]struct{}{}
	var out []string
	for _, file := range files {
		mp, err := modulePathForFile(repoPath, file)
		if err != nil || mp == "" {
			continue
		}
		if _, ok := seen[mp]; ok {
			continue
		}
		seen[mp] = struct{}{}
		out = append(out, mp)
	}
	sort.Strings(out)
	return out, nil
}

func modulePathForFile(repoPath, file string) (string, error) {
	file = filepath.ToSlash(strings.TrimSpace(file))
	if file == "" {
		return "", nil
	}
	abs := file
	if !filepath.IsAbs(file) {
		abs = filepath.Join(repoPath, filepath.FromSlash(file))
	}
	abs = filepath.Clean(abs)
	dir := filepath.Dir(abs)
	repoPath = filepath.Clean(repoPath)
	for {
		if !strings.HasPrefix(dir, repoPath) {
			return "", nil
		}
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			mp, _, err := discoverModulePath(dir)
			return mp, err
		}
		if dir == repoPath {
			return "", nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}

// ChangedFiles returns paths from git working tree and last commit diff.
func ChangedFiles(repoPath string) ([]string, error) {
	seen := map[string]struct{}{}
	add := func(paths []string) {
		for _, p := range paths {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			seen[p] = struct{}{}
		}
	}

	for _, args := range [][]string{
		{"diff", "--name-only", "HEAD"},
		{"diff", "--name-only"},
		{"diff", "--cached", "--name-only"},
	} {
		out, err := gitOutput(repoPath, args...)
		if err != nil {
			return nil, err
		}
		if len(out) > 0 {
			add(strings.Split(string(out), "\n"))
		}
	}

	var files []string
	for p := range seen {
		files = append(files, p)
	}
	sort.Strings(files)
	return files, nil
}

func gitOutput(repoPath string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	return cmd.Output()
}

// FilterModuleRoots returns module roots matching requested module paths.
func FilterModuleRoots(roots []moduleRoot, modulePaths []string) ([]moduleRoot, error) {
	if len(modulePaths) == 0 {
		return roots, nil
	}
	want := map[string]struct{}{}
	for _, mp := range modulePaths {
		mp = strings.TrimSpace(mp)
		if mp == "" {
			continue
		}
		want[mp] = struct{}{}
	}
	if len(want) == 0 {
		return nil, fmt.Errorf("no module paths specified")
	}
	var filtered []moduleRoot
	for _, root := range roots {
		if _, ok := want[root.modulePath]; ok {
			filtered = append(filtered, root)
		}
	}
	if len(filtered) == 0 {
		return nil, fmt.Errorf("no matching modules found for: %s", strings.Join(modulePaths, ", "))
	}
	return filtered, nil
}

func fileUnderModuleRoots(repoPath string, relFile string, roots []moduleRoot) bool {
	abs := filepath.Join(repoPath, filepath.FromSlash(relFile))
	abs = filepath.Clean(abs)
	for _, root := range roots {
		rootPath := filepath.Clean(root.path)
		if abs == rootPath || strings.HasPrefix(abs, rootPath+string(filepath.Separator)) {
			return true
		}
	}
	return false
}
