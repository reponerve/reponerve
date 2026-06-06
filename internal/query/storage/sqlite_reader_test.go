package storage

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	memorymodels "reponerve/internal/memory/models"
	memorystorage "reponerve/internal/memory/storage"
	"reponerve/internal/storage/migrations"
	"reponerve/internal/storage/sqlite"
	models "reponerve/pkg/models"
)

func TestSQLiteReaders(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-query-readers-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := migrations.RunUp(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	ctx := context.Background()
	repoID1 := "repo_1"
	repoID2 := "repo_2"

	// 1. Insert Repositories
	_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repoID1, "Repo 1", tempDir, "main")
	if err != nil {
		t.Fatalf("failed to insert repository 1: %v", err)
	}
	_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repoID2, "Repo 2", tempDir, "main")
	if err != nil {
		t.Fatalf("failed to insert repository 2: %v", err)
	}

	// 2. Insert Sources (required for FK constraints)
	_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_1", repoID1, "adr", "docs/adr/0001.md", "Author", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("failed to insert source: %v", err)
	}
	_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_2", repoID2, "commit", "commit_1", "Author", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("failed to insert source: %v", err)
	}

	// Instantiate stores (writers) and readers
	eventStore := sqlite.NewEventStore(db)
	eventReader := NewSQLiteEventReader(db)

	decisionStore := memorystorage.NewSQLiteDecisionStore(db)
	decisionReader := NewSQLiteDecisionReader(db)

	intentStore := memorystorage.NewSQLiteIntentStore(db)
	intentReader := NewSQLiteIntentReader(db)

	factStore := memorystorage.NewSQLiteFactStore(db)
	factReader := NewSQLiteFactReader(db)

	relStore := memorystorage.NewSQLiteRelationshipStore(db)
	relReader := NewSQLiteRelationshipReader(db)

	// --- TEST EMPTY RESULT SETS ---
	t.Run("Empty Database Checks", func(t *testing.T) {
		// Event
		_, err := eventReader.GetByID(ctx, "non_existent")
		if err != sql.ErrNoRows {
			t.Errorf("expected sql.ErrNoRows, got %v", err)
		}
		list, _ := eventReader.ListAll(ctx)
		if len(list) != 0 {
			t.Errorf("expected empty list, got %d", len(list))
		}
		listRepo, _ := eventReader.ListByRepository(ctx, repoID1)
		if len(listRepo) != 0 {
			t.Errorf("expected empty list by repo, got %d", len(listRepo))
		}

		// Decision
		_, err = decisionReader.GetByID(ctx, "non_existent")
		if err != sql.ErrNoRows {
			t.Errorf("expected sql.ErrNoRows, got %v", err)
		}
		decList, _ := decisionReader.ListAll(ctx)
		if len(decList) != 0 {
			t.Errorf("expected empty list, got %d", len(decList))
		}

		// Intent
		_, err = intentReader.GetByID(ctx, "non_existent")
		if err != sql.ErrNoRows {
			t.Errorf("expected sql.ErrNoRows, got %v", err)
		}
		intList, _ := intentReader.ListAll(ctx)
		if len(intList) != 0 {
			t.Errorf("expected empty list, got %d", len(intList))
		}

		// Fact
		_, err = factReader.GetByID(ctx, "non_existent")
		if err != sql.ErrNoRows {
			t.Errorf("expected sql.ErrNoRows, got %v", err)
		}
		factList, _ := factReader.ListAll(ctx)
		if len(factList) != 0 {
			t.Errorf("expected empty list, got %d", len(factList))
		}

		// Relationship
		_, err = relReader.GetByID(ctx, "non_existent")
		if err != sql.ErrNoRows {
			t.Errorf("expected sql.ErrNoRows, got %v", err)
		}
		relList, _ := relReader.ListAll(ctx)
		if len(relList) != 0 {
			t.Errorf("expected empty list, got %d", len(relList))
		}
	})

	// --- WRITE RECORDS TO TEST POPULATED CHECKS ---
	// 1. Events (one with empty description, one with nil/null description)
	evt1 := &models.Event{
		ID:           "evt_1",
		RepositoryID: repoID1,
		EventType:    "FEATURE_INTRODUCED",
		Title:        "Feature One",
		Description:  "", // Will write as NULL
		SourceID:     "src_1",
		Timestamp:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	evt2 := &models.Event{
		ID:           "evt_2",
		RepositoryID: repoID2,
		EventType:    "DEFECT_RESOLVED",
		Title:        "Fix One",
		Description:  "Fixed bug",
		SourceID:     "src_2",
		Timestamp:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	if err := eventStore.UpsertEvent(ctx, evt1); err != nil {
		t.Fatalf("failed to insert event 1: %v", err)
	}
	if err := eventStore.UpsertEvent(ctx, evt2); err != nil {
		t.Fatalf("failed to insert event 2: %v", err)
	}

	// 2. Decisions
	dec1 := &memorymodels.Decision{
		ID:           "dec_1",
		RepositoryID: repoID1,
		Title:        "Use SQLite",
		Status:       "Accepted",
		SourceID:     "src_1",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	dec2 := &memorymodels.Decision{
		ID:           "dec_2",
		RepositoryID: repoID2,
		Title:        "Use Go",
		Status:       "Accepted",
		SourceID:     "src_2",
		CreatedAt:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	if err := decisionStore.UpsertDecision(ctx, dec1); err != nil {
		t.Fatalf("failed to insert decision 1: %v", err)
	}
	if err := decisionStore.UpsertDecision(ctx, dec2); err != nil {
		t.Fatalf("failed to insert decision 2: %v", err)
	}

	// 3. Intents
	int1 := &memorymodels.Intent{
		ID:           "int_1",
		RepositoryID: repoID1,
		Description:  "Simplify Configuration",
		SourceID:     "src_1",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := intentStore.UpsertIntent(ctx, int1); err != nil {
		t.Fatalf("failed to insert intent 1: %v", err)
	}

	// 4. Facts
	fact1 := &memorymodels.Fact{
		ID:           "fact_1",
		RepositoryID: repoID1,
		Subject:      "Auth Service",
		Predicate:    "USES",
		Object:       "Redis",
		SourceID:     "src_1",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := factStore.UpsertFact(ctx, fact1); err != nil {
		t.Fatalf("failed to insert fact 1: %v", err)
	}

	// 5. Relationships
	rel1 := &memorymodels.Relationship{
		ID:           "rel_1",
		RepositoryID: repoID1,
		FromID:       "int_1",
		ToID:         "dec_1",
		Type:         "INTENT_DRIVES_DECISION",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := relStore.UpsertRelationship(ctx, rel1); err != nil {
		t.Fatalf("failed to insert relationship 1: %v", err)
	}

	// --- RUN TESTS ON POPULATED DATABASE ---
	t.Run("Event Reader Tests", func(t *testing.T) {
		// GetByID (checks safe null handling for description on evt1)
		evt, err := eventReader.GetByID(ctx, "evt_1")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if evt.Title != "Feature One" {
			t.Errorf("expected Title 'Feature One', got %q", evt.Title)
		}
		if evt.Description != "" {
			t.Errorf("expected Description to be empty string, got %q", evt.Description)
		}

		evtWithDesc, err := eventReader.GetByID(ctx, "evt_2")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if evtWithDesc.Description != "Fixed bug" {
			t.Errorf("expected Description 'Fixed bug', got %q", evtWithDesc.Description)
		}

		// ListAll
		all, err := eventReader.ListAll(ctx)
		if err != nil {
			t.Fatalf("ListAll failed: %v", err)
		}
		if len(all) != 2 {
			t.Errorf("expected 2 events, got %d", len(all))
		}

		// ListByRepository
		repo1Events, err := eventReader.ListByRepository(ctx, repoID1)
		if err != nil {
			t.Fatalf("ListByRepository failed: %v", err)
		}
		if len(repo1Events) != 1 {
			t.Errorf("expected 1 event for repo 1, got %d", len(repo1Events))
		}
		if repo1Events[0].ID != "evt_1" {
			t.Errorf("expected event evt_1, got %s", repo1Events[0].ID)
		}
	})

	t.Run("Decision Reader Tests", func(t *testing.T) {
		dec, err := decisionReader.GetByID(ctx, "dec_1")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if dec.Title != "Use SQLite" {
			t.Errorf("expected Title 'Use SQLite', got %q", dec.Title)
		}

		all, err := decisionReader.ListAll(ctx)
		if err != nil {
			t.Fatalf("ListAll failed: %v", err)
		}
		if len(all) != 2 {
			t.Errorf("expected 2 decisions, got %d", len(all))
		}

		repo1Decisions, err := decisionReader.ListByRepository(ctx, repoID1)
		if err != nil {
			t.Fatalf("ListByRepository failed: %v", err)
		}
		if len(repo1Decisions) != 1 {
			t.Errorf("expected 1 decision for repo 1, got %d", len(repo1Decisions))
		}
	})

	t.Run("Intent Reader Tests", func(t *testing.T) {
		intent, err := intentReader.GetByID(ctx, "int_1")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if intent.Description != "Simplify Configuration" {
			t.Errorf("expected Description 'Simplify Configuration', got %q", intent.Description)
		}

		all, err := intentReader.ListAll(ctx)
		if err != nil {
			t.Fatalf("ListAll failed: %v", err)
		}
		if len(all) != 1 {
			t.Errorf("expected 1 intent, got %d", len(all))
		}

		repo2Intents, err := intentReader.ListByRepository(ctx, repoID2)
		if err != nil {
			t.Fatalf("ListByRepository failed: %v", err)
		}
		if len(repo2Intents) != 0 {
			t.Errorf("expected 0 intents for repo 2, got %d", len(repo2Intents))
		}
	})

	t.Run("Fact Reader Tests", func(t *testing.T) {
		fact, err := factReader.GetByID(ctx, "fact_1")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if fact.Subject != "Auth Service" || fact.Predicate != "USES" || fact.Object != "Redis" {
			t.Errorf("unexpected fact contents: %+v", fact)
		}

		all, err := factReader.ListAll(ctx)
		if err != nil {
			t.Fatalf("ListAll failed: %v", err)
		}
		if len(all) != 1 {
			t.Errorf("expected 1 fact, got %d", len(all))
		}
	})

	t.Run("Relationship Reader Tests", func(t *testing.T) {
		rel, err := relReader.GetByID(ctx, "rel_1")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if rel.FromID != "int_1" || rel.ToID != "dec_1" || rel.Type != "INTENT_DRIVES_DECISION" {
			t.Errorf("unexpected relationship contents: %+v", rel)
		}

		all, err := relReader.ListAll(ctx)
		if err != nil {
			t.Fatalf("ListAll failed: %v", err)
		}
		if len(all) != 1 {
			t.Errorf("expected 1 relationship, got %d", len(all))
		}
	})
}
