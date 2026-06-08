package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	ownerextraction "github.com/reponerve/reponerve/internal/ownership/extraction"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
	"github.com/reponerve/reponerve/pkg/models"
)

func TestOwnershipExtractorIntegration(t *testing.T) {
	// Set up temp directory and SQLite database
	tempDir, err := os.MkdirTemp("", "reponerve-ownership-extractor-*")
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
	repoID := "repo_ownership_extractor_test"

	// Insert test repository
	_, err = db.Exec(
		"INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())",
		repoID, "test-repo", tempDir, "main",
	)
	if err != nil {
		t.Fatalf("failed to insert test repository: %v", err)
	}

	t1 := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 3, 3, 15, 0, 0, 0, time.UTC)

	// Define test commit sources (some from same author for deduplication testing)
	testSources := []*models.Source{
		{
			ID:           "commit_1",
			RepositoryID: repoID,
			SourceType:   "commit",
			Reference:    "commit_1",
			Title:        "feat(auth): introduce jwt",
			Author:       "Alice <alice@example.com>",
			Timestamp:    t2,
		},
		{
			ID:           "commit_2",
			RepositoryID: repoID,
			SourceType:   "commit",
			Reference:    "commit_2",
			Title:        "fix: token issue",
			Author:       "Alice Smith <alice@example.com>",
			Timestamp:    t1,
		},
		{
			ID:           "commit_3",
			RepositoryID: repoID,
			SourceType:   "commit",
			Reference:    "commit_3",
			Title:        "initial repo checkin",
			Author:       "Bob <bob@example.com>",
			Timestamp:    t3,
		},
		{
			ID:           "adr_001",
			RepositoryID: repoID,
			SourceType:   "adr",
			Reference:    "docs/adr/0001.md",
			Title:        "use jwt",
			Author:       "Alice",
			Timestamp:    t1,
		},
	}

	// Insert sources into DB
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
	extractor := ownerextraction.NewExtractor()
	contribs, err := extractor.Extract(ctx, testSources)
	if err != nil {
		t.Fatalf("extraction failed: %v", err)
	}

	// Should extract 2 contributors (Alice and Bob), ignoring the ADR source
	if len(contribs) != 2 {
		t.Fatalf("expected 2 contributors, got %d", len(contribs))
	}

	// Persist contributors via SQLiteContributorStore
	store := sqlite.NewSQLiteContributorStore(db)
	for _, c := range contribs {
		if err := store.UpsertContributor(ctx, c); err != nil {
			t.Fatalf("failed to upsert contributor %s: %v", c.ID, err)
		}
	}

	// Verify count in DB
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM contributors WHERE repository_id = ?", repoID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query contributor count: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 contributors in DB, got %d", count)
	}

	// Find Alice in DB
	var gotName, gotEmail string
	var gotFirstSeen, gotLastSeen time.Time
	var gotCommitCount int
	err = db.QueryRow("SELECT name, email, first_seen, last_seen, commit_count FROM contributors WHERE email = ?", "alice@example.com").
		Scan(&gotName, &gotEmail, &gotFirstSeen, &gotLastSeen, &gotCommitCount)
	if err != nil {
		t.Fatalf("failed to query Alice: %v", err)
	}

	if gotName != "Alice Smith" {
		t.Errorf("expected name 'Alice Smith', got %q", gotName)
	}
	if gotCommitCount != 2 {
		t.Errorf("expected commit count 2, got %d", gotCommitCount)
	}
	if !gotFirstSeen.Equal(t1) || !gotLastSeen.Equal(t2) {
		t.Errorf("incorrect timestamp boundaries for Alice: %v to %v", gotFirstSeen, gotLastSeen)
	}

	// Verify idempotency
	for _, c := range contribs {
		if err := store.UpsertContributor(ctx, c); err != nil {
			t.Fatalf("second upsert failed: %v", err)
		}
	}
	err = db.QueryRow("SELECT COUNT(*) FROM contributors WHERE repository_id = ?", repoID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query contributor count after re-upsert: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 contributors after idempotent upserts, got %d", count)
	}
}
