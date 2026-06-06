package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"reponerve/internal/extraction/fact"
	memorystorage "reponerve/internal/memory/storage"
	"reponerve/internal/storage/migrations"
	"reponerve/internal/storage/sqlite"
	"reponerve/pkg/models"
)

func TestFactExtractorIntegration(t *testing.T) {
	// Set up temp directory and SQLite database
	tempDir, err := os.MkdirTemp("", "reponerve-fact-extractor-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if err := migrations.RunUp(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	ctx := context.Background()
	repoID := "repo_fact_extractor_test"

	// Insert test repository
	_, err = db.Exec(
		"INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())",
		repoID, "test-repo", tempDir, "main",
	)
	if err != nil {
		t.Fatalf("failed to insert test repository: %v", err)
	}

	// Define test sources (both adr and commit)
	testSources := []*models.Source{
		{
			ID:           "adr_001",
			RepositoryID: repoID,
			SourceType:   "adr",
			Reference:    "docs/adr/0001-use-sqlite.md",
			Title:        "Use SQLite Database",
			Timestamp:    time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
			MetadataJSON: `{"content": "Authentication Service uses Redis. API Gateway depends on Auth Service.", "path": "docs/adr/0001-use-sqlite.md"}`,
		},
		{
			ID:           "adr_002",
			RepositoryID: repoID,
			SourceType:   "adr",
			Reference:    "docs/adr/0002-postgres.md",
			Title:        "Use Postgres",
			Timestamp:    time.Date(2024, 3, 2, 10, 0, 0, 0, time.UTC),
			MetadataJSON: `{"content": "Billing stores data in PostgreSQL.", "path": "docs/adr/0002-postgres.md"}`,
		},
		{
			ID:           "commit_001",
			RepositoryID: repoID,
			SourceType:   "commit",
			Reference:    "commit_001",
			Title:        "feat: User Service calls Notification Service",
			Timestamp:    time.Date(2024, 3, 3, 9, 0, 0, 0, time.UTC),
		},
	}

	// Insert sources into DB so FK constraints pass
	for _, src := range testSources {
		_, err = db.Exec(
			"INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, datetime())",
			src.ID, src.RepositoryID, src.SourceType, src.Reference, src.Title, src.Author, src.Timestamp,
		)
		if err != nil {
			t.Fatalf("failed to insert source %s: %v", src.ID, err)
		}
	}

	// Run extractor
	extractor := fact.NewExtractor()
	facts, err := extractor.Extract(ctx, testSources)
	if err != nil {
		t.Fatalf("extraction failed: %v", err)
	}

	// We expect 3 facts (2 from adr_001, 1 from adr_002, commit_001 should be ignored)
	if len(facts) != 3 {
		t.Fatalf("expected 3 facts, got %d", len(facts))
	}

	// Persist facts via FactStore
	factStore := memorystorage.NewSQLiteFactStore(db)
	for _, f := range facts {
		if err := factStore.UpsertFact(ctx, f); err != nil {
			t.Fatalf("failed to upsert fact %s: %v", f.ID, err)
		}
	}

	// Verify facts are persisted in DB
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM memory_facts WHERE repository_id = ?", repoID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query fact count: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 facts in DB, got %d", count)
	}

	// Verify first fact
	var subject, predicate, object, sourceID string
	err = db.QueryRow(
		"SELECT subject, predicate, object, source_id FROM memory_facts WHERE source_id = ? AND predicate = ?",
		"adr_001", "USES",
	).Scan(&subject, &predicate, &object, &sourceID)
	if err != nil {
		t.Fatalf("failed to query USES fact: %v", err)
	}
	if subject != "Authentication Service" {
		t.Errorf("expected subject 'Authentication Service', got %q", subject)
	}
	if object != "Redis" {
		t.Errorf("expected object 'Redis', got %q", object)
	}

	// Verify second fact
	err = db.QueryRow(
		"SELECT subject, predicate, object, source_id FROM memory_facts WHERE source_id = ? AND predicate = ?",
		"adr_001", "DEPENDS_ON",
	).Scan(&subject, &predicate, &object, &sourceID)
	if err != nil {
		t.Fatalf("failed to query DEPENDS_ON fact: %v", err)
	}
	if subject != "API Gateway" {
		t.Errorf("expected subject 'API Gateway', got %q", subject)
	}
	if object != "Auth Service" {
		t.Errorf("expected object 'Auth Service', got %q", object)
	}

	// Verify third fact
	err = db.QueryRow(
		"SELECT subject, predicate, object, source_id FROM memory_facts WHERE source_id = ? AND predicate = ?",
		"adr_002", "STORES_IN",
	).Scan(&subject, &predicate, &object, &sourceID)
	if err != nil {
		t.Fatalf("failed to query STORES_IN fact: %v", err)
	}
	if subject != "Billing" {
		t.Errorf("expected subject 'Billing', got %q", subject)
	}
	if object != "PostgreSQL" {
		t.Errorf("expected object 'PostgreSQL', got %q", object)
	}

	// Verify idempotency
	for _, f := range facts {
		if err := factStore.UpsertFact(ctx, f); err != nil {
			t.Fatalf("second upsert failed: %v", err)
		}
	}
	err = db.QueryRow("SELECT COUNT(*) FROM memory_facts WHERE repository_id = ?", repoID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query fact count after re-upsert: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 facts after idempotent upsert, got %d", count)
	}
}
