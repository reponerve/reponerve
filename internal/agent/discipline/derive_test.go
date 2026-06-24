package discipline

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDerive_ADRAndCIAndGoLayout(t *testing.T) {
	repo := t.TempDir()
	adrDir := filepath.Join(repo, "docs", "adr")
	if err := os.MkdirAll(adrDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(adrDir, "0001-test.md"), []byte("# Test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	wfDir := filepath.Join(repo, ".github", "workflows")
	if err := os.MkdirAll(wfDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wfDir, "test.yml"), []byte("name: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "internal", "foo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module example.com\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	policy := Derive(context.Background(), DeriveInput{
		RepositoryID:   "repo-1",
		RepositoryPath: repo,
		ADRsIndexed:    1,
	})

	if policy.ADRDirectory != "docs/adr" {
		t.Fatalf("adr_directory=%q", policy.ADRDirectory)
	}
	if !policy.RequireADROnArchitecture {
		t.Fatal("expected require_adr_on_architecture")
	}
	if len(policy.CIWorkflowFiles) == 0 {
		t.Fatal("expected ci workflow files")
	}
	if policy.DominantLanguage != "go" {
		t.Fatalf("dominant_language=%q", policy.DominantLanguage)
	}
	if len(policy.LayerConventions) == 0 {
		t.Fatal("expected layer conventions")
	}
}

func TestWriteAndLoadPolicy(t *testing.T) {
	ws := t.TempDir()
	policy := Derive(context.Background(), DeriveInput{
		RepositoryID:   "repo-1",
		RepositoryPath: ws,
	})
	if err := WritePolicy(ws, policy); err != nil {
		t.Fatalf("WritePolicy: %v", err)
	}
	loaded, err := LoadPolicy(ws)
	if err != nil {
		t.Fatalf("LoadPolicy: %v", err)
	}
	if loaded == nil || loaded.RepositoryID != "repo-1" {
		t.Fatalf("loaded=%+v", loaded)
	}
}
