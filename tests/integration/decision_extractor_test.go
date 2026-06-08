package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/reponerve/reponerve/internal/extraction/decision"
	memorystorage "github.com/reponerve/reponerve/internal/memory/storage"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
	"github.com/reponerve/reponerve/pkg/models"
)

func TestDecisionExtractorIntegration(t *testing.T) {
	// Set up temp directory and SQLite database
	tempDir, err := os.MkdirTemp("", "reponerve-decision-extractor-*")
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
	repoID := "repo_decision_extractor_test"

	// Insert test repository
	_, err = db.Exec(
		"INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())",
		repoID, "test-repo", tempDir, "main",
	)
	if err != nil {
		t.Fatalf("failed to insert test repository: %v", err)
	}

	// Define test ADR sources
	testSources := []*models.Source{
		{
			ID:           "adr_001",
			RepositoryID: repoID,
			SourceType:   "adr",
			Reference:    "docs/adr/0001-use-sqlite.md",
			Title:        "Use SQLite Database",
			Timestamp:    time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
			MetadataJSON: `{"status": "Accepted", "path": "docs/adr/0001-use-sqlite.md"}`,
		},
		{
			ID:           "adr_002",
			RepositoryID: repoID,
			SourceType:   "adr",
			Reference:    "docs/adr/0002-use-mongodb.md",
			Title:        "Use MongoDB",
			Timestamp:    time.Date(2024, 3, 2, 10, 0, 0, 0, time.UTC),
			MetadataJSON: `{"status": "Rejected", "path": "docs/adr/0002-use-mongodb.md"}`,
		},
		{
			// Non-ADR source that should be skipped
			ID:           "commit_001",
			RepositoryID: repoID,
			SourceType:   "commit",
			Reference:    "commit_001",
			Title:        "feat: add database config",
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
	extractor := decision.NewExtractor()
	decisions, err := extractor.Extract(ctx, testSources)
	if err != nil {
		t.Fatalf("extraction failed: %v", err)
	}

	// Only ADR sources qualify
	if len(decisions) != 2 {
		t.Fatalf("expected 2 decisions, got %d", len(decisions))
	}

	// Persist decisions via DecisionStore
	decisionStore := memorystorage.NewSQLiteDecisionStore(db)
	for _, dec := range decisions {
		if err := decisionStore.UpsertDecision(ctx, dec); err != nil {
			t.Fatalf("failed to upsert decision %s: %v", dec.ID, err)
		}
	}

	// Verify decisions are persisted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM memory_decisions WHERE repository_id = ?", repoID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query decision count: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 decisions in DB, got %d", count)
	}

	// Verify fields for the accepted ADR
	var title, status, sourceID string
	err = db.QueryRow(
		"SELECT title, status, source_id FROM memory_decisions WHERE source_id = ?",
		"adr_001",
	).Scan(&title, &status, &sourceID)
	if err != nil {
		t.Fatalf("failed to query accepted decision: %v", err)
	}
	if title != "Use SQLite Database" {
		t.Errorf("expected title 'Use SQLite Database', got %q", title)
	}
	if status != "Accepted" {
		t.Errorf("expected status 'Accepted', got %q", status)
	}
	if sourceID != "adr_001" {
		t.Errorf("expected source_id 'adr_001', got %q", sourceID)
	}

	// Verify status for the rejected ADR
	err = db.QueryRow(
		"SELECT status FROM memory_decisions WHERE source_id = ?",
		"adr_002",
	).Scan(&status)
	if err != nil {
		t.Fatalf("failed to query rejected decision: %v", err)
	}
	if status != "Rejected" {
		t.Errorf("expected status 'Rejected', got %q", status)
	}

	// Verify idempotency: upserting the same decisions again must not create duplicates
	for _, dec := range decisions {
		if err := decisionStore.UpsertDecision(ctx, dec); err != nil {
			t.Fatalf("second upsert failed: %v", err)
		}
	}
	err = db.QueryRow("SELECT COUNT(*) FROM memory_decisions WHERE repository_id = ?", repoID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query decision count after re-upsert: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 decisions after idempotent upsert, got %d", count)
	}
}
