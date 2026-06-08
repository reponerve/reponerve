package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/reponerve/reponerve/internal/memory/linker"
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	memorystorage "github.com/reponerve/reponerve/internal/memory/storage"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
	models "github.com/reponerve/reponerve/pkg/models"
)

func TestLinkerIntegration(t *testing.T) {
	// Set up temp directory and SQLite database
	tempDir, err := os.MkdirTemp("", "reponerve-linker-integration-*")
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
	repoID := "repo_linker_test"

	// Insert test repository
	_, err = db.Exec(
		"INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())",
		repoID, "test-repo", tempDir, "main",
	)
	if err != nil {
		t.Fatalf("failed to insert test repository: %v", err)
	}

	// Define memories to link
	it := &memorymodels.Intent{
		ID:           "intent_001",
		RepositoryID: repoID,
		Description:  "Reduce latency",
		SourceID:     "adr_001",
		CreatedAt:    time.Now(),
	}
	dec := &memorymodels.Decision{
		ID:           "decision_001",
		RepositoryID: repoID,
		Title:        "Use Redis Cache",
		Status:       "Accepted",
		SourceID:     "adr_001",
		CreatedAt:    time.Now(),
	}
	evt := &models.Event{
		ID:           "event_001",
		RepositoryID: repoID,
		EventType:    "FEATURE_INTRODUCED",
		Title:        "Introduce Redis Cache",
		SourceID:     "commit_001",
		Timestamp:    time.Now(),
	}
	f := &memorymodels.Fact{
		ID:           "fact_001",
		RepositoryID: repoID,
		Subject:      "Auth Service",
		Predicate:    "USES",
		Object:       "Redis",
		SourceID:     "adr_001",
		CreatedAt:    time.Now(),
	}

	// Run Linker
	l := linker.NewLinker()
	rels, err := l.Link(ctx, linker.LinkInput{
		Intents:   []*memorymodels.Intent{it},
		Decisions: []*memorymodels.Decision{dec},
		Events:    []*models.Event{evt},
		Facts:     []*memorymodels.Fact{f},
	})
	if err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	// We expect 3 relationships:
	// - intent_001 -> decision_001 (INTENT_DRIVES_DECISION)
	// - decision_001 -> event_001 (DECISION_RESULTS_IN_EVENT)
	// - fact_001 -> decision_001 (FACT_SUPPORTS_DECISION)
	if len(rels) != 3 {
		t.Fatalf("expected 3 relationships, got %d", len(rels))
	}

	// Persist relationships via RelationshipStore
	relStore := memorystorage.NewSQLiteRelationshipStore(db)
	for _, rel := range rels {
		if err := relStore.UpsertRelationship(ctx, rel); err != nil {
			t.Fatalf("failed to upsert relationship %s: %v", rel.ID, err)
		}
	}

	// Verify relationships are persisted in DB
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM memory_relationships WHERE repository_id = ?", repoID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query relationship count: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 relationships in DB, got %d", count)
	}

	// Verify specific relationships
	var fromID, toID, relType string
	err = db.QueryRow(
		"SELECT from_id, to_id, relationship_type FROM memory_relationships WHERE relationship_type = ?",
		"INTENT_DRIVES_DECISION",
	).Scan(&fromID, &toID, &relType)
	if err != nil {
		t.Fatalf("failed to query INTENT_DRIVES_DECISION: %v", err)
	}
	if fromID != "intent_001" || toID != "decision_001" {
		t.Errorf("unexpected INTENT_DRIVES_DECISION relation: %s -> %s", fromID, toID)
	}

	err = db.QueryRow(
		"SELECT from_id, to_id, relationship_type FROM memory_relationships WHERE relationship_type = ?",
		"DECISION_RESULTS_IN_EVENT",
	).Scan(&fromID, &toID, &relType)
	if err != nil {
		t.Fatalf("failed to query DECISION_RESULTS_IN_EVENT: %v", err)
	}
	if fromID != "decision_001" || toID != "event_001" {
		t.Errorf("unexpected DECISION_RESULTS_IN_EVENT relation: %s -> %s", fromID, toID)
	}

	err = db.QueryRow(
		"SELECT from_id, to_id, relationship_type FROM memory_relationships WHERE relationship_type = ?",
		"FACT_SUPPORTS_DECISION",
	).Scan(&fromID, &toID, &relType)
	if err != nil {
		t.Fatalf("failed to query FACT_SUPPORTS_DECISION: %v", err)
	}
	if fromID != "fact_001" || toID != "decision_001" {
		t.Errorf("unexpected FACT_SUPPORTS_DECISION relation: %s -> %s", fromID, toID)
	}

	// Verify idempotency
	for _, rel := range rels {
		if err := relStore.UpsertRelationship(ctx, rel); err != nil {
			t.Fatalf("second upsert failed: %v", rel.ID)
		}
	}
	err = db.QueryRow("SELECT COUNT(*) FROM memory_relationships WHERE repository_id = ?", repoID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query count after second upsert: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 relationships after idempotent upsert, got %d", count)
	}
}
