package adr

import (
	"os"
	"path/filepath"
	"strings"
)

// DocumentKind classifies ingested markdown directories.
type DocumentKind string

const (
	DocumentKindADR             DocumentKind = "adr"
	DocumentKindArchitectureDoc DocumentKind = "architecture_doc"
)

// DocumentPath is one repository-relative directory scanned for markdown sources.
type DocumentPath struct {
	Path string       `mapstructure:"path" json:"path" yaml:"path"`
	Kind DocumentKind `mapstructure:"kind" json:"kind" yaml:"kind"`
}

// DefaultDocumentPaths are scanned when config does not override paths.
func DefaultDocumentPaths() []DocumentPath {
	return []DocumentPath{
		{Path: "docs/adr", Kind: DocumentKindADR},
		{Path: "docs/adrs", Kind: DocumentKindADR},
		{Path: "docs/decisions", Kind: DocumentKindADR},
		{Path: "docs/rfc", Kind: DocumentKindADR},
		{Path: "adr", Kind: DocumentKindADR},
		{Path: "adrs", Kind: DocumentKindADR},
		{Path: "docs/architecture", Kind: DocumentKindArchitectureDoc},
	}
}

// ResolveDocumentPaths merges configured paths with defaults (config first, then defaults, deduped).
func ResolveDocumentPaths(configured []DocumentPath) []DocumentPath {
	seen := make(map[string]struct{})
	var out []DocumentPath
	appendPath := func(p DocumentPath) {
		p.Path = normalizeRelPath(p.Path)
		if p.Path == "" {
			return
		}
		if p.Kind == "" {
			p.Kind = DocumentKindADR
		}
		key := p.Path + "|" + string(p.Kind)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, p)
	}
	for _, p := range configured {
		appendPath(p)
	}
	for _, p := range DefaultDocumentPaths() {
		appendPath(p)
	}
	return out
}

func normalizeRelPath(path string) string {
	path = filepath.ToSlash(strings.TrimSpace(path))
	path = strings.Trim(path, "/")
	if path == "." || path == ".." {
		return ""
	}
	return path
}

func (p DocumentPath) scanTarget(repoPath string) scanTarget {
	rel := normalizeRelPath(p.Path)
	switch p.Kind {
	case DocumentKindArchitectureDoc:
		return scanTarget{
			dir:        filepath.Join(repoPath, rel),
			sourceType: "architecture_doc",
			idPrefix:   "archdoc_",
		}
	default:
		return scanTarget{
			dir:        filepath.Join(repoPath, rel),
			sourceType: "adr",
			idPrefix:   "adr_",
		}
	}
}

// PrimaryADRDirectory returns the first ADR path that exists with markdown.
func PrimaryADRDirectory(repoPath string, paths []DocumentPath) string {
	for _, p := range paths {
		if p.Kind != DocumentKindADR {
			continue
		}
		dir := filepath.Join(repoPath, normalizeRelPath(p.Path))
		if hasMarkdownFiles(dir) {
			return normalizeRelPath(p.Path)
		}
	}
	return ""
}

func hasMarkdownFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
			return true
		}
	}
	return false
}
