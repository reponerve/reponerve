package sqlite_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/reponerve/reponerve/internal/storage"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

func TestMemorySearchStore_RebuildAndSearch(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-memory-search-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	db, err := sqlite.Open(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := migrations.RunUp(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	store := sqlite.NewMemorySearchStore(db)
	ctx := context.Background()
	repoID := "repo_test"

	docs := []storage.MemorySearchDocument{
		{
			MemoryID:     "evt_1",
			RepositoryID: repoID,
			EntityType:   "EVENT",
			Title:        "Expose Repository Intelligence Through MCP Tools",
			Content:      "MCP integration",
		},
		{
			MemoryID:     "evt_2",
			RepositoryID: repoID,
			EntityType:   "EVENT",
			Title:        "Update Copyright Year",
			Content:      "LICENSE README",
		},
	}

	if err := store.Rebuild(ctx, repoID, docs); err != nil {
		t.Fatalf("rebuild failed: %v", err)
	}

	hits, err := store.Search(ctx, repoID, []string{"MCP"}, "EVENT")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if hits[0].MemoryID != "evt_1" {
		t.Fatalf("expected evt_1, got %s", hits[0].MemoryID)
	}

	filtered, err := store.Search(ctx, repoID, []string{"MCP"}, "DECISION")
	if err != nil {
		t.Fatalf("filtered search failed: %v", err)
	}
	if len(filtered) != 0 {
		t.Fatalf("expected 0 decision hits, got %d", len(filtered))
	}
}

func TestMemorySearchStore_HyphenatedTermsDoNotError(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-memory-search-hyphen-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	db, err := sqlite.Open(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := migrations.RunUp(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	store := sqlite.NewMemorySearchStore(db)
	ctx := context.Background()
	repoID := "repo_test"

	docs := []storage.MemorySearchDocument{
		{
			MemoryID:     "fact_1",
			RepositoryID: repoID,
			EntityType:   "FACT",
			Title:        "Metadata Panel DEPENDS_ON user-service",
			Content:      "user service integration boundary",
		},
	}
	if err := store.Rebuild(ctx, repoID, docs); err != nil {
		t.Fatalf("rebuild failed: %v", err)
	}

	hits, err := store.Search(ctx, repoID, []string{"user-service"}, "")
	if err != nil {
		t.Fatalf("hyphenated search should not error: %v", err)
	}
	if len(hits) == 0 {
		t.Fatalf("expected hits for hyphenated query")
	}
}
