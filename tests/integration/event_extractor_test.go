package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/reponerve/reponerve/internal/extraction/event"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
	"github.com/reponerve/reponerve/pkg/models"
)

func TestEventExtractorIntegration(t *testing.T) {
	// Set up temp directory and SQLite database
	tempDir, err := os.MkdirTemp("", "reponerve-event-extractor-*")
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
	repoID := "repo_event_extractor_test"

	// Insert test repository
	_, err = db.Exec(
		"INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())",
		repoID, "test-repo", tempDir, "main",
	)
	if err != nil {
		t.Fatalf("failed to insert test repository: %v", err)
	}

	// Define test commit sources
	testSources := []*models.Source{
		{
			ID:           "commit_feat_001",
			RepositoryID: repoID,
			SourceType:   "commit",
			Reference:    "commit_feat_001",
			Title:        "feat(auth): introduce jwt authentication",
			Author:       "Alice",
			Timestamp:    time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			ID:           "commit_fix_001",
			RepositoryID: repoID,
			SourceType:   "commit",
			Reference:    "commit_fix_001",
			Title:        "fix: resolve token expiry bug",
			Author:       "Bob",
			Timestamp:    time.Date(2024, 3, 2, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:           "commit_plain_001",
			RepositoryID: repoID,
			SourceType:   "commit",
			Reference:    "commit_plain_001",
			Title:        "initial project setup",
			Author:       "Charlie",
			Timestamp:    time.Date(2024, 2, 28, 9, 0, 0, 0, time.UTC),
		},
		{
			ID:           "adr_001",
			RepositoryID: repoID,
			SourceType:   "adr",
			Reference:    "docs/adr/001-use-jwt.md",
			Title:        "feat: this adr should be skipped",
			Timestamp:    time.Date(2024, 2, 27, 0, 0, 0, 0, time.UTC),
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
	extractor := event.NewExtractor()
	events, err := extractor.Extract(ctx, testSources)
	if err != nil {
		t.Fatalf("extraction failed: %v", err)
	}

	// Only feat and fix commits qualify; plain commit and adr are skipped
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	// Persist events via EventStore
	eventStore := sqlite.NewEventStore(db)
	for _, evt := range events {
		if err := eventStore.UpsertEvent(ctx, evt); err != nil {
			t.Fatalf("failed to upsert event %s: %v", evt.ID, err)
		}
	}

	// Verify events are persisted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM memory_events WHERE repository_id = ?", repoID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query event count: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 events in DB, got %d", count)
	}

	// Verify EventType for the feat commit
	var eventType, sourceID string
	err = db.QueryRow(
		"SELECT event_type, source_id FROM memory_events WHERE source_id = ?",
		"commit_feat_001",
	).Scan(&eventType, &sourceID)
	if err != nil {
		t.Fatalf("failed to query feat event: %v", err)
	}
	if eventType != event.EventTypeFeatureIntroduced {
		t.Errorf("expected event_type %q, got %q", event.EventTypeFeatureIntroduced, eventType)
	}
	if sourceID != "commit_feat_001" {
		t.Errorf("expected source_id %q, got %q", "commit_feat_001", sourceID)
	}

	// Verify EventType for the fix commit
	err = db.QueryRow(
		"SELECT event_type, source_id FROM memory_events WHERE source_id = ?",
		"commit_fix_001",
	).Scan(&eventType, &sourceID)
	if err != nil {
		t.Fatalf("failed to query fix event: %v", err)
	}
	if eventType != event.EventTypeDefectResolved {
		t.Errorf("expected event_type %q, got %q", event.EventTypeDefectResolved, eventType)
	}

	// Verify idempotency: upserting the same events again must not create duplicates
	for _, evt := range events {
		if err := eventStore.UpsertEvent(ctx, evt); err != nil {
			t.Fatalf("second upsert failed: %v", err)
		}
	}
	err = db.QueryRow("SELECT COUNT(*) FROM memory_events WHERE repository_id = ?", repoID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query event count after re-upsert: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 events after idempotent upsert, got %d", count)
	}
}
