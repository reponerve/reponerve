package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/reponerve/reponerve/internal/extraction/intent"
	memorystorage "github.com/reponerve/reponerve/internal/memory/storage"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
	"github.com/reponerve/reponerve/pkg/models"
)

func TestIntentExtractorIntegration(t *testing.T) {
	// Set up temp directory and SQLite database
	tempDir, err := os.MkdirTemp("", "reponerve-intent-extractor-*")
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
	repoID := "repo_intent_extractor_test"

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
			MetadataJSON: `{"content": "We need to reduce database latency and improve response times.", "path": "docs/adr/0001-use-sqlite.md"}`,
		},
		{
			ID:           "commit_001",
			RepositoryID: repoID,
			SourceType:   "commit",
			Reference:    "commit_001",
			Title:        "feat(cache): optimize cache invalidation logic",
			Timestamp:    time.Date(2024, 3, 2, 10, 0, 0, 0, time.UTC),
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
	extractor := intent.NewExtractor()
	intents, err := extractor.Extract(ctx, testSources)
	if err != nil {
		t.Fatalf("extraction failed: %v", err)
	}

	// We expect 3 intents:
	// - "Reduce Database Latency" from adr_001
	// - "Improve Response Times" from adr_001
	// - "Optimize Cache Invalidation Logic" from commit_001
	if len(intents) != 3 {
		t.Fatalf("expected 3 intents, got %d", len(intents))
	}

	// Persist intents via IntentStore
	intentStore := memorystorage.NewSQLiteIntentStore(db)
	for _, it := range intents {
		if err := intentStore.UpsertIntent(ctx, it); err != nil {
			t.Fatalf("failed to upsert intent %s: %v", it.ID, err)
		}
	}

	// Verify intents are persisted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM memory_intents WHERE repository_id = ?", repoID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query intent count: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 intents in DB, got %d", count)
	}

	// Verify description for the first extracted intent
	var desc, sourceID string
	err = db.QueryRow(
		"SELECT description, source_id FROM memory_intents WHERE source_id = ? AND description LIKE ?",
		"adr_001", "%Latency",
	).Scan(&desc, &sourceID)
	if err != nil {
		t.Fatalf("failed to query adr intent 1: %v", err)
	}
	if desc != "Reduce Database Latency" {
		t.Errorf("expected description 'Reduce Database Latency', got %q", desc)
	}

	// Verify description for the second extracted intent
	err = db.QueryRow(
		"SELECT description, source_id FROM memory_intents WHERE source_id = ? AND description LIKE ?",
		"adr_001", "%Response%",
	).Scan(&desc, &sourceID)
	if err != nil {
		t.Fatalf("failed to query adr intent 2: %v", err)
	}
	if desc != "Improve Response Times" {
		t.Errorf("expected description 'Improve Response Times', got %q", desc)
	}

	// Verify description for the commit intent
	err = db.QueryRow(
		"SELECT description, source_id FROM memory_intents WHERE source_id = ?",
		"commit_001",
	).Scan(&desc, &sourceID)
	if err != nil {
		t.Fatalf("failed to query commit intent: %v", err)
	}
	if desc != "Optimize Cache Invalidation Logic" {
		t.Errorf("expected description 'Optimize Cache Invalidation Logic', got %q", desc)
	}

	// Verify idempotency
	for _, it := range intents {
		if err := intentStore.UpsertIntent(ctx, it); err != nil {
			t.Fatalf("second upsert failed: %v", err)
		}
	}
	err = db.QueryRow("SELECT COUNT(*) FROM memory_intents WHERE repository_id = ?", repoID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query intent count after re-upsert: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 intents after idempotent upsert, got %d", count)
	}
}
