package codeowners

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Rule represents a single rule mapping a file/directory pattern to owner identities.
type Rule struct {
	Pattern string   `json:"pattern"`
	Owners  []string `json:"owners"`
}

// FindAndParse searches for CODEOWNERS in standard repo locations and parses it.
func FindAndParse(repoPath string) ([]Rule, error) {
	locations := []string{
		filepath.Join(repoPath, ".github", "CODEOWNERS"),
		filepath.Join(repoPath, "CODEOWNERS"),
		filepath.Join(repoPath, "docs", "CODEOWNERS"),
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return ParseFile(loc)
		}
	}
	return nil, nil
}

// ParseFile parses a given CODEOWNERS file into a slice of Rules.
func ParseFile(path string) ([]Rule, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var rules []Rule
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pattern := fields[0]
		owners := fields[1:]
		rules = append(rules, Rule{
			Pattern: pattern,
			Owners:  owners,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return rules, nil
}

// Matches checks if a relative path matches a CODEOWNERS pattern.
func Matches(pattern, path string) bool {
	pattern = filepath.ToSlash(pattern)
	path = filepath.ToSlash(path)

	if pattern == "*" {
		return true
	}
	trimmedPattern := strings.Trim(pattern, "/")
	trimmedPath := strings.Trim(path, "/")

	if strings.HasPrefix(trimmedPath, trimmedPattern) {
		return true
	}

	matched, err := filepath.Match(pattern, path)
	if err == nil && matched {
		return true
	}
	return false
}
