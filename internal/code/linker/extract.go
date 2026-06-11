package linker

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	goFilePathPattern = regexp.MustCompile(`(?:^|[\s"'(,])([a-zA-Z0-9][\w./-]*\.go)`)
	packagePathPattern = regexp.MustCompile(`(?:^|[\s"'(,])((?:internal|cmd|pkg)/[\w./-]+)`)
)

type textMatch struct {
	Value string
	Field string
}

func extractGoFilePaths(text, field string) []textMatch {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	seen := make(map[string]struct{})
	var matches []textMatch
	for _, sub := range goFilePathPattern.FindAllStringSubmatch(text, -1) {
		if len(sub) < 2 {
			continue
		}
		path := normalizePath(sub[1])
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		matches = append(matches, textMatch{Value: path, Field: field})
	}
	return matches
}

func extractPackagePaths(text, field string) []textMatch {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	seen := make(map[string]struct{})
	var matches []textMatch
	for _, sub := range packagePathPattern.FindAllStringSubmatch(text, -1) {
		if len(sub) < 2 {
			continue
		}
		path := normalizePath(sub[1])
		if strings.HasSuffix(path, ".go") {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		matches = append(matches, textMatch{Value: path, Field: field})
	}
	return matches
}

func extractQualifiedSymbols(text, field string, qualifiedNames []string) []textMatch {
	if strings.TrimSpace(text) == "" || len(qualifiedNames) == 0 {
		return nil
	}
	names := append([]string(nil), qualifiedNames...)
	sort.Slice(names, func(i, j int) bool {
		return len(names[i]) > len(names[j])
	})

	seen := make(map[string]struct{})
	var matches []textMatch
	lowerText := strings.ToLower(text)
	for _, name := range names {
		if name == "" {
			continue
		}
		lowerName := strings.ToLower(name)
		idx := strings.Index(lowerText, lowerName)
		if idx < 0 {
			continue
		}
		if !symbolBoundary(text, idx, len(name)) {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		matches = append(matches, textMatch{Value: name, Field: field})
	}
	return matches
}

func symbolBoundary(text string, idx, length int) bool {
	if idx > 0 {
		prev := text[idx-1]
		if isIdentChar(prev) {
			return false
		}
	}
	end := idx + length
	if end < len(text) && isIdentChar(text[end]) {
		return false
	}
	return true
}

func isIdentChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_' || b == '.'
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, `"'`)
	path = filepath.ToSlash(path)
	path = strings.TrimPrefix(path, "./")
	return path
}
