package indexer_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/reponerve/reponerve/internal/code"
	"github.com/reponerve/reponerve/internal/code/indexer"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

func TestIndexer_MultiLanguageRepo(t *testing.T) {
	repoPath := filepath.Join("testdata", "multilang")
	absRepo, err := filepath.Abs(repoPath)
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}

	tempDir, err := os.MkdirTemp("", "reponerve-multilang-*")
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
		db,
		sqlite.NewSQLiteCodeEntityStore(db),
		sqlite.NewSQLiteCodeRelationshipStore(db),
		sqlite.NewSQLiteRepositoryCodeRelationshipStore(db),
		sqlite.NewSQLiteCodeIndexStateStore(db),
	)

	repoID := "repo_multilang"
	_, err = db.Exec(`INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())`,
		repoID, "multilang", absRepo, "main")
	if err != nil {
		t.Fatalf("insert repository: %v", err)
	}

	if err := idx.Index(context.Background(), repoID, absRepo); err != nil {
		t.Fatalf("index failed: %v", err)
	}

	assertEntity(t, db, repoID, codemodels.EntityTypeFile, "frontend/src/api.ts")
	assertEntity(t, db, repoID, codemodels.EntityTypeFunction, "frontend/src.getApiBase")
	assertEntity(t, db, repoID, codemodels.EntityTypeStruct, "frontend/src.ApiClient")
	assertEntity(t, db, repoID, codemodels.EntityTypeFunction, "services/app.health_check")
	assertEntity(t, db, repoID, codemodels.EntityTypeStruct, "services/app.Handler")
	assertEntity(t, db, repoID, codemodels.EntityTypeFunction, "crates/api/src.run")
	assertEntity(t, db, repoID, codemodels.EntityTypeStruct, "crates/api/src.Service")
	assertEntity(t, db, repoID, codemodels.EntityTypeFunction, "frontend/src.greet")
	assertEntity(t, db, repoID, codemodels.EntityTypeStruct, "java/src/main/java/com/example/api.Handler")
	assertEntity(t, db, repoID, codemodels.EntityTypeStruct, "dotnet/Api.Handler")
	assertEntity(t, db, repoID, codemodels.EntityTypeStruct, "ruby/lib.Handler")
	assertEntity(t, db, repoID, codemodels.EntityTypeStruct, "kotlin/src/main/kotlin/com/example.Handler")
	assertEntity(t, db, repoID, codemodels.EntityTypeStruct, "swift/Sources.Handler")
	assertEntity(t, db, repoID, codemodels.EntityTypeStruct, "php/src.Handler")
	assertEntity(t, db, repoID, codemodels.EntityTypeStruct, "cpp/src.Handler")
	assertEntity(t, db, repoID, codemodels.EntityTypeFunction, "c/src.run")
	assertEntity(t, db, repoID, codemodels.EntityTypeStruct, "scala/src/main/scala/com/example.Handler")
	assertEntity(t, db, repoID, codemodels.EntityTypeStruct, "lua/lib.Handler")
	assertEntity(t, db, repoID, codemodels.EntityTypeFunction, "bash/scripts.health_check")
	assertEntity(t, db, repoID, codemodels.EntityTypeStruct, "sql.handlers")
	assertEntity(t, db, repoID, codemodels.EntityTypeStruct, "dart/lib.Handler")
	assertEntity(t, db, repoID, codemodels.EntityTypeStruct, "elixir/lib.App.Handler")
	assertEntity(t, db, repoID, codemodels.EntityTypeStruct, "zig/src.Handler")

	codeSvc := code.NewService(
		storage.NewSQLiteCodeEntityReader(db),
		storage.NewSQLiteCodeRelationshipReader(db),
		storage.NewSQLiteRepositoryCodeRelationshipReader(db),
	)
	ctxOut, err := codeSvc.ResolveFile(context.Background(), repoID, "frontend/src/api.ts")
	if err != nil {
		t.Fatalf("resolve file: %v", err)
	}
	if len(ctxOut.Files) == 0 || len(ctxOut.Functions) == 0 {
		t.Fatalf("expected file and function context, got %+v", ctxOut)
	}
}

func assertEntity(t *testing.T, db *sqlite.Database, repoID, entityType, qualified string) {
	t.Helper()
	var id string
	err := db.QueryRow(`
		SELECT id FROM code_entities
		WHERE repository_id = ? AND entity_type = ? AND qualified_name = ?
	`, repoID, entityType, qualified).Scan(&id)
	if err != nil {
		t.Fatalf("expected entity %s %q: %v", entityType, qualified, err)
	}
}
