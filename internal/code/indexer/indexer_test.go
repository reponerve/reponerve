package indexer_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/reponerve/reponerve/internal/code/indexer"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

func TestIndexer_SampleModule(t *testing.T) {
	repoPath := filepath.Join("testdata", "samplemodule")
	absRepo, err := filepath.Abs(repoPath)
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}

	tempDir, err := os.MkdirTemp("", "reponerve-code-indexer-*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	db, err := sqlite.Open(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := migrations.RunUp(db); err != nil {
		t.Fatalf("migrations: %v", err)
	}

	idx := indexer.New(
		sqlite.NewSQLiteCodeEntityStore(db),
		sqlite.NewSQLiteCodeRelationshipStore(db),
		sqlite.NewSQLiteRepositoryCodeRelationshipStore(db),
		sqlite.NewSQLiteCodeIndexStateStore(db),
	)

	repoID := "repo_sample"
	_, err = db.Exec(`INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())`,
		repoID, "sample", absRepo, "main")
	if err != nil {
		t.Fatalf("insert repository: %v", err)
	}

	if err := idx.Index(context.Background(), repoID, absRepo); err != nil {
		t.Fatalf("index failed: %v", err)
	}

	var entityCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM code_entities WHERE repository_id = ?`, repoID).Scan(&entityCount); err != nil {
		t.Fatalf("count entities: %v", err)
	}
	if entityCount < 8 {
		t.Fatalf("expected at least 8 entities, got %d", entityCount)
	}

	var relCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM code_relationships WHERE repository_id = ?`, repoID).Scan(&relCount); err != nil {
		t.Fatalf("count relationships: %v", err)
	}
	if relCount < 8 {
		t.Fatalf("expected relationships, got %d", relCount)
	}

	var loginID string
	err = db.QueryRow(`
		SELECT id FROM code_entities
		WHERE repository_id = ? AND entity_type = ? AND qualified_name = ?
	`, repoID, codemodels.EntityTypeMethod, "internal/auth.Service.Login").Scan(&loginID)
	if err != nil {
		t.Fatalf("expected Login method entity: %v", err)
	}

	var serviceStruct string
	err = db.QueryRow(`
		SELECT id FROM code_entities
		WHERE repository_id = ? AND entity_type = ? AND qualified_name = ?
	`, repoID, codemodels.EntityTypeStruct, "internal/auth.Service").Scan(&serviceStruct)
	if err != nil {
		t.Fatalf("expected Service struct entity: %v", err)
	}
}
