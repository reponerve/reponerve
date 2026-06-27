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

func TestMergeCodeIndexModulesDeletesScopedRepositoryCodeLinks(t *testing.T) {
	ctx := context.Background()
	tempDir, err := os.MkdirTemp("", "reponerve-code-index-tx-*")
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
	if _, err := db.ExecContext(ctx, `INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at)
		VALUES ('repo1', 'Repo', '/tmp/repo', 'main', datetime('now'), datetime('now'))`); err != nil {
		t.Fatalf("failed to seed repository: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	oldScoped := testCodeEntity("scoped-old", "repo1", "module/a", "a/old.go", now)
	untouched := testCodeEntity("untouched", "repo1", "module/b", "b/file.go", now)
	if err := db.ReplaceCodeIndex(ctx, "repo1", []*codemodels.CodeEntity{oldScoped, untouched}, nil); err != nil {
		t.Fatalf("replace code index: %v", err)
	}

	linkStore := sqlite.NewSQLiteRepositoryCodeRelationshipStore(db)
	if err := linkStore.UpsertRepositoryCodeRelationship(ctx, testRepositoryCodeLink("link-scoped", "repo1", oldScoped.ID, now)); err != nil {
		t.Fatalf("upsert scoped repository-code link: %v", err)
	}
	if err := linkStore.UpsertRepositoryCodeRelationship(ctx, testRepositoryCodeLink("link-untouched", "repo1", untouched.ID, now)); err != nil {
		t.Fatalf("upsert untouched repository-code link: %v", err)
	}

	newScoped := testCodeEntity("scoped-new", "repo1", "module/a", "a/new.go", now)
	if err := db.MergeCodeIndexModules(ctx, "repo1", []string{"module/a"}, []*codemodels.CodeEntity{newScoped}, nil); err != nil {
		t.Fatalf("merge scoped module: %v", err)
	}

	var scopedLinks int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM repository_code_relationships WHERE code_entity_id = ?`, oldScoped.ID).Scan(&scopedLinks); err != nil {
		t.Fatalf("query scoped links: %v", err)
	}
	if scopedLinks != 0 {
		t.Fatalf("expected scoped repository-code links to be deleted, got %d", scopedLinks)
	}

	var untouchedLinks int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM repository_code_relationships WHERE code_entity_id = ?`, untouched.ID).Scan(&untouchedLinks); err != nil {
		t.Fatalf("query untouched links: %v", err)
	}
	if untouchedLinks != 1 {
		t.Fatalf("expected untouched module link to remain, got %d", untouchedLinks)
	}

	var newScopedEntities int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM code_entities WHERE id = ?`, newScoped.ID).Scan(&newScopedEntities); err != nil {
		t.Fatalf("query new scoped entity: %v", err)
	}
	if newScopedEntities != 1 {
		t.Fatalf("expected new scoped entity to be inserted, got %d", newScopedEntities)
	}
}

func testCodeEntity(id, repositoryID, modulePath, filePath string, indexedAt time.Time) *codemodels.CodeEntity {
	return &codemodels.CodeEntity{
		ID:            id,
		RepositoryID:  repositoryID,
		EntityType:    codemodels.EntityTypeFile,
		Name:          filepath.Base(filePath),
		QualifiedName: filePath,
		FilePath:      filePath,
		PackagePath:   filepath.Dir(filePath),
		ModulePath:    modulePath,
		Language:      "go",
		StartLine:     1,
		EndLine:       1,
		EvidenceJSON:  `{"source":"test"}`,
		IndexedAt:     indexedAt,
	}
}

func testRepositoryCodeLink(id, repositoryID, codeEntityID string, indexedAt time.Time) *codemodels.RepositoryCodeRelationship {
	return &codemodels.RepositoryCodeRelationship{
		ID:                   id,
		RepositoryID:         repositoryID,
		RepositoryEntityID:   "repo-entity-" + id,
		RepositoryEntityType: "EVENT",
		CodeEntityID:         codeEntityID,
		CodeEntityType:       codemodels.EntityTypeFile,
		RelationshipType:     "EVENT_REFERENCES_CODE",
		EvidenceJSON:         `{"source":"test"}`,
		IndexedAt:            indexedAt,
	}
}
