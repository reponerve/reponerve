package sessionmemory

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	memorystorage "github.com/reponerve/reponerve/internal/memory/storage"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
	pkgmodels "github.com/reponerve/reponerve/pkg/models"
)

func TestRememberForgetRoundTrip(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "mem.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := migrations.RunUp(db); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	repoID := "repo_session"
	_, err = db.Exec(`INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())`,
		repoID, "test", dir, "main")
	if err != nil {
		t.Fatal(err)
	}

	factStore := memorystorage.NewSQLiteFactStore(db)
	sourceStore := sqlite.NewSourceStore(db)
	eventReader := storage.NewSQLiteEventReader(db)
	decisionReader := storage.NewSQLiteDecisionReader(db)
	factReader := storage.NewSQLiteFactReader(db)
	sourceReader := storage.NewSQLiteSourceReader(db)
	searchStore := sqlite.NewMemorySearchStore(db)

	svc := NewService(factStore, sourceStore, factReader, eventReader, decisionReader, searchStore, sourceReader, dir)

	fact, err := svc.Remember(ctx, RememberRequest{
		RepositoryID: repoID,
		Subject:      "authentication",
		Content:      "Use JWT middleware in internal/auth",
	})
	if err != nil {
		t.Fatal(err)
	}

	list, err := svc.ListSessionFacts(ctx, repoID)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 session fact, got %v err=%v", list, err)
	}

	bundle, err := svc.ExportHandoff(ctx, repoID)
	if err != nil {
		t.Fatal(err)
	}
	bundlePath := filepath.Join(dir, "handoff.json")
	if err := ExportHandoffFile(bundle, bundlePath); err != nil {
		t.Fatal(err)
	}

	if err := svc.Forget(ctx, repoID, fact.ID); err != nil {
		t.Fatal(err)
	}
	list, _ = svc.ListSessionFacts(ctx, repoID)
	if len(list) != 0 {
		t.Fatalf("expected 0 facts after forget, got %d", len(list))
	}

	imported, err := ImportHandoffFile(bundlePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.ImportHandoff(ctx, imported); err != nil {
		t.Fatal(err)
	}
	list, _ = svc.ListSessionFacts(ctx, repoID)
	if len(list) != 1 {
		t.Fatalf("expected handoff import to restore 1 fact, got %d", len(list))
	}
}

func TestWritebackQA(t *testing.T) {
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "mem.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_ = migrations.RunUp(db)
	ctx := context.Background()
	repoID := "repo_qa"
	_, _ = db.Exec(`INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())`,
		repoID, "test", dir, "main")

	svc := NewService(
		memorystorage.NewSQLiteFactStore(db),
		sqlite.NewSourceStore(db),
		storage.NewSQLiteFactReader(db),
		storage.NewSQLiteEventReader(db),
		storage.NewSQLiteDecisionReader(db),
		sqlite.NewMemorySearchStore(db),
		storage.NewSQLiteSourceReader(db),
		dir,
	)
	fact, err := svc.WritebackQA(ctx, WritebackRequest{
		RepositoryID: repoID,
		Question:     "Why SQLite?",
		Answer:       "Local-first ADR",
	})
	if err != nil {
		t.Fatal(err)
	}
	if fact.Subject != "Why SQLite?" {
		t.Fatalf("unexpected subject: %s", fact.Subject)
	}
	_ = memorymodels.Fact{}
	_ = pkgmodels.Source{}
	_ = time.Now()
}
