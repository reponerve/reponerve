package indexer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type moduleRoot struct {
	path       string
	modulePath string
	goModFile  string
}

func discoverModuleRoots(repoPath string) ([]moduleRoot, error) {
	workPath := filepath.Join(repoPath, "go.work")
	if _, err := os.Stat(workPath); err == nil {
		return parseGoWork(repoPath, workPath)
	}

	modulePath, goModFile, err := discoverModulePath(repoPath)
	if err != nil {
		return nil, err
	}
	return []moduleRoot{{path: repoPath, modulePath: modulePath, goModFile: goModFile}}, nil
}

func discoverModulePath(repoPath string) (string, string, error) {
	goModPath := filepath.Join(repoPath, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return "", "", fmt.Errorf("read go.mod: %w", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			mod := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			mod = strings.Trim(mod, `"`)
			if mod == "" {
				return "", "", fmt.Errorf("empty module path in go.mod")
			}
			rel, _ := filepath.Rel(repoPath, goModPath)
			return mod, filepath.ToSlash(rel), nil
		}
	}

	return "", "", fmt.Errorf("module directive not found in go.mod")
}

func parseGoWork(repoPath, workPath string) ([]moduleRoot, error) {
	data, err := os.ReadFile(workPath)
	if err != nil {
		return nil, fmt.Errorf("read go.work: %w", err)
	}

	var roots []moduleRoot
	inUse := false
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "use (" {
			inUse = true
			continue
		}
		if inUse && line == ")" {
			break
		}
		if !inUse {
			continue
		}
		usePath := strings.Trim(strings.TrimSpace(line), `"`)
		if usePath == "" {
			continue
		}
		moduleDir := filepath.Clean(filepath.Join(repoPath, usePath))
		modulePath, goModFile, err := discoverModulePath(moduleDir)
		if err != nil {
			continue
		}
		roots = append(roots, moduleRoot{
			path:       moduleDir,
			modulePath: modulePath,
			goModFile:  filepath.ToSlash(filepath.Join(usePath, goModFile)),
		})
	}

	if len(roots) == 0 {
		return nil, fmt.Errorf("no modules found in go.work")
	}
	return roots, nil
}
