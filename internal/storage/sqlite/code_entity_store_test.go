package sqlite_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

func TestSQLiteCodeEntityStoreUpsert(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-code-entity-store-*")
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

	_, err = db.Exec(`INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at)
		VALUES ('repo1', 'Repo', '/tmp', 'main', datetime('now'), datetime('now'))`)
	if err != nil {
		t.Fatalf("failed to seed repository: %v", err)
	}

	store := sqlite.NewSQLiteCodeEntityStore(db)
	now := time.Now().UTC().Truncate(time.Second)
	entity := &codemodels.CodeEntity{
		ID:            "entity1",
		RepositoryID:  "repo1",
		EntityType:    codemodels.EntityTypeFunction,
		Name:          "Login",
		QualifiedName: "internal/auth.Login",
		FilePath:      "internal/auth/auth.go",
		PackagePath:   "internal/auth",
		ModulePath:    "github.com/example/app",
		Language:      "go",
		StartLine:     10,
		EndLine:       20,
		EvidenceJSON:  `{"source":"go/ast"}`,
		IndexedAt:     now,
	}

	if err := store.UpsertCodeEntity(context.Background(), entity); err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM code_entities WHERE id = 'entity1'`).Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 entity, got %d", count)
	}
}
