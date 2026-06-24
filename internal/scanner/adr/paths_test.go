package adr

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/reponerve/reponerve/pkg/models"
)

func TestResolveDocumentPaths_MergesConfigWithDefaults(t *testing.T) {
	paths := ResolveDocumentPaths([]DocumentPath{
		{Path: "docs/decisions", Kind: DocumentKindADR},
	})
	seen := make(map[string]struct{})
	for _, p := range paths {
		key := p.Path + "|" + string(p.Kind)
		if _, ok := seen[key]; ok {
			t.Fatalf("duplicate path %s", key)
		}
		seen[key] = struct{}{}
	}
	if _, ok := seen["docs/decisions|adr"]; !ok {
		t.Fatal("expected configured path docs/decisions")
	}
	if _, ok := seen["docs/adr|adr"]; !ok {
		t.Fatal("expected default path docs/adr")
	}
}

func TestScanner_CustomDocumentPath(t *testing.T) {
	repo := t.TempDir()
	customDir := filepath.Join(repo, "architecture", "records")
	if err := os.MkdirAll(customDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "# Custom ADR\n\nStatus: Accepted\n"
	if err := os.WriteFile(filepath.Join(customDir, "001-custom.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner(DocumentPath{Path: "architecture/records", Kind: DocumentKindADR})
	sources, err := scanner.Scan(context.Background(), &models.Repository{ID: "repo-1", Path: repo})
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(sources))
	}
	if sources[0].Title != "Custom ADR" {
		t.Fatalf("unexpected title %q", sources[0].Title)
	}
}
