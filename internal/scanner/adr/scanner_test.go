package adr

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/reponerve/reponerve/pkg/models"
)

func TestParseADR(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectedTitle  string
		expectedStatus string
	}{
		{
			name: "standard adr with status section",
			content: `# 1. Use SQLite

## Status

Accepted

## Context

Some context here...`,
			expectedTitle:  "1. Use SQLite",
			expectedStatus: "Accepted",
		},
		{
			name: "adr with inline status tag",
			content: `# Use SQLite Database
Status: Proposed

We want to decide on a local database.`,
			expectedTitle:  "Use SQLite Database",
			expectedStatus: "Proposed",
		},
		{
			name: "adr with status followed by another header",
			content: `# 5. Context Packs
## Status
Approved
## Context
Details...`,
			expectedTitle:  "5. Context Packs",
			expectedStatus: "Approved",
		},
		{
			name: "adr with no status section",
			content: `# Just a title
Some normal body without status.`,
			expectedTitle:  "Just a title",
			expectedStatus: "Accepted",
		},
		{
			name:           "adr with no title and no status",
			content:        `Some description text.`,
			expectedTitle:  "",
			expectedStatus: "Accepted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, status := ParseADR(tt.content)
			if title != tt.expectedTitle {
				t.Errorf("expected title %q, got %q", tt.expectedTitle, title)
			}
			if status != tt.expectedStatus {
				t.Errorf("expected status %q, got %q", tt.expectedStatus, status)
			}
		})
	}
}

func TestScanner_ArchitectureDocs(t *testing.T) {
	dir := t.TempDir()
	archDir := filepath.Join(dir, "docs", "architecture")
	if err := os.MkdirAll(archDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(archDir, "architecture-overview.md")
	if err := os.WriteFile(path, []byte("# Architecture Overview\n\nStatus: Accepted\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner()
	sources, err := scanner.Scan(context.Background(), &models.Repository{ID: "repo-1", Path: dir})
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("expected 1 architecture source, got %d", len(sources))
	}
	if sources[0].SourceType != "architecture_doc" {
		t.Fatalf("expected architecture_doc, got %q", sources[0].SourceType)
	}
	if sources[0].Reference != "docs/architecture/architecture-overview.md" {
		t.Fatalf("unexpected reference: %q", sources[0].Reference)
	}
}
