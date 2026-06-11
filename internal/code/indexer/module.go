package indexer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
			return mod, "go.mod", nil
		}
	}

	return "", "", fmt.Errorf("module directive not found in go.mod")
}
