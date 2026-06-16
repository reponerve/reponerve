package linker_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/reponerve/reponerve/internal/code"
	codelinker "github.com/reponerve/reponerve/internal/code/linker"
	"github.com/reponerve/reponerve/internal/code/indexer"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

func TestLinker_LinksEventToCodeFile(t *testing.T) {
	repoPath := filepath.Join("..", "indexer", "testdata", "samplemodule")
	absRepo, err := filepath.Abs(repoPath)
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}

	tempDir, err := os.MkdirTemp("", "reponerve-code-linker-*")
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

	repoID := "repo_linker"
	_, err = db.Exec(`INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())`,
		repoID, "sample", absRepo, "main")
	if err != nil {
		t.Fatalf("insert repository: %v", err)
	}

	codeEntityStore := sqlite.NewSQLiteCodeEntityStore(db)
	codeRelStore := sqlite.NewSQLiteCodeRelationshipStore(db)
	stateStore := sqlite.NewSQLiteCodeIndexStateStore(db)
	idx := indexer.New(db, codeEntityStore, codeRelStore, sqlite.NewSQLiteRepositoryCodeRelationshipStore(db), stateStore)
	if err := idx.Index(context.Background(), repoID, absRepo); err != nil {
		t.Fatalf("index failed: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at)
		VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())
	`, "src_1", repoID, "commit", "abc123", "sample commit", "author")
	if err != nil {
		t.Fatalf("insert source: %v", err)
	}

	evtID := "evt_link_test"
	_, err = db.Exec(`
		INSERT INTO memory_events (id, repository_id, event_type, title, description, source_id, timestamp, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, evtID, repoID, "FEATURE_INTRODUCED",
		"Update internal/auth/service.go for login flow", "", "src_1", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("insert event: %v", err)
	}

	linker := codelinker.New(
		storage.NewSQLiteEventReader(db),
		storage.NewSQLiteDecisionReader(db),
		storage.NewSQLiteFactReader(db),
		storage.NewSQLiteSourceReader(db),
		storage.NewSQLiteCodeEntityReader(db),
		sqlite.NewSQLiteRepositoryCodeRelationshipStore(db),
		stateStore,
	)
	if err := linker.Link(context.Background(), repoID); err != nil {
		t.Fatalf("link failed: %v", err)
	}

	fileQualified := "internal/auth/service.go"
	fileEntityID := code.EntityID(repoID, codemodels.EntityTypeFile, fileQualified)

	var count int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM repository_code_relationships
		WHERE repository_id = ? AND repository_entity_id = ? AND code_entity_id = ?
	`, repoID, evtID, fileEntityID).Scan(&count)
	if err != nil {
		t.Fatalf("query links: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 repository-code link, got %d", count)
	}

	var linkCount int
	if err := db.QueryRow(`SELECT link_count FROM code_index_state WHERE repository_id = ?`, repoID).Scan(&linkCount); err != nil {
		t.Fatalf("read link count: %v", err)
	}
	if linkCount < 1 {
		t.Fatalf("expected link_count >= 1, got %d", linkCount)
	}
}
