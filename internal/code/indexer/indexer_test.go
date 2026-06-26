package indexer_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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
		db,
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

	var serviceSignature string
	err = db.QueryRow(`
		SELECT signature FROM code_entities
		WHERE repository_id = ? AND entity_type = ? AND qualified_name = ?
	`, repoID, codemodels.EntityTypeStruct, "internal/auth.Service").Scan(&serviceSignature)
	if err != nil {
		t.Fatalf("expected Service struct entity: %v", err)
	}
	if !strings.Contains(serviceSignature, "store *store.Store") {
		t.Fatalf("expected struct fields in signature, got %q", serviceSignature)
	}
}

func TestIndexer_MultiModuleScanPersistsEachModulePath(t *testing.T) {
	repoPath := t.TempDir()
	writeFile(t, filepath.Join(repoPath, "go.work"), "go 1.22\n\nuse (\n\t./moda\n\t./modb\n)\n")
	writeFile(t, filepath.Join(repoPath, "moda", "go.mod"), "module example.com/moda\n\ngo 1.22\n")
	writeFile(t, filepath.Join(repoPath, "moda", "a", "a.go"), "package a\n\nfunc A() {}\n")
	writeFile(t, filepath.Join(repoPath, "modb", "go.mod"), "module example.com/modb\n\ngo 1.22\n")
	writeFile(t, filepath.Join(repoPath, "modb", "b", "b.go"), "package b\n\nfunc B() {}\n")

	tempDir := t.TempDir()
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

	repoID := "repo_multimodule"
	if _, err := db.Exec(`INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())`,
		repoID, "multimodule", repoPath, "main"); err != nil {
		t.Fatalf("insert repository: %v", err)
	}

	if err := idx.IndexModules(context.Background(), repoID, repoPath, []string{"example.com/moda", "example.com/modb"}); err != nil {
		t.Fatalf("index modules failed: %v", err)
	}

	rows, err := db.Query(`
		SELECT file_path, module_path FROM code_entities
		WHERE repository_id = ? AND entity_type = ? AND file_path IN (?, ?)
	`, repoID, codemodels.EntityTypeFile, "moda/a/a.go", "modb/b/b.go")
	if err != nil {
		t.Fatalf("query module paths: %v", err)
	}
	defer rows.Close()

	got := map[string]string{}
	for rows.Next() {
		var filePath, modulePath string
		if err := rows.Scan(&filePath, &modulePath); err != nil {
			t.Fatalf("scan module path: %v", err)
		}
		got[filePath] = modulePath
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}

	want := map[string]string{
		"moda/a/a.go": "example.com/moda",
		"modb/b/b.go": "example.com/modb",
	}
	for filePath, wantModule := range want {
		if got[filePath] != wantModule {
			t.Fatalf("%s module_path = %q, want %q (all: %#v)", filePath, got[filePath], wantModule, got)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
