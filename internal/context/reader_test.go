package context

import (
	stdcontext "context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	memorymodels "reponerve/internal/memory/models"
	memorystorage "reponerve/internal/memory/storage"
	"reponerve/internal/query/storage"
	"reponerve/internal/storage/migrations"
	"reponerve/internal/storage/sqlite"
	models "reponerve/pkg/models"
)

// --- Mock Readers for Unit Testing ---

type mockEventReader struct {
	events []*models.Event
	err    error
}

func (m *mockEventReader) GetByID(ctx stdcontext.Context, id string) (*models.Event, error) {
	return nil, nil
}
func (m *mockEventReader) ListByRepository(ctx stdcontext.Context, repositoryID string) ([]*models.Event, error) {
	return m.events, m.err
}
func (m *mockEventReader) ListAll(ctx stdcontext.Context) ([]*models.Event, error) {
	return m.events, m.err
}

type mockDecisionReader struct {
	decisions []*memorymodels.Decision
	err       error
}

func (m *mockDecisionReader) GetByID(ctx stdcontext.Context, id string) (*memorymodels.Decision, error) {
	return nil, nil
}
func (m *mockDecisionReader) ListByRepository(ctx stdcontext.Context, repositoryID string) ([]*memorymodels.Decision, error) {
	return m.decisions, m.err
}
func (m *mockDecisionReader) ListAll(ctx stdcontext.Context) ([]*memorymodels.Decision, error) {
	return m.decisions, m.err
}

type mockIntentReader struct {
	intents []*memorymodels.Intent
	err     error
}

func (m *mockIntentReader) GetByID(ctx stdcontext.Context, id string) (*memorymodels.Intent, error) {
	return nil, nil
}
func (m *mockIntentReader) ListByRepository(ctx stdcontext.Context, repositoryID string) ([]*memorymodels.Intent, error) {
	return m.intents, m.err
}
func (m *mockIntentReader) ListAll(ctx stdcontext.Context) ([]*memorymodels.Intent, error) {
	return m.intents, m.err
}

type mockFactReader struct {
	facts []*memorymodels.Fact
	err   error
}

func (m *mockFactReader) GetByID(ctx stdcontext.Context, id string) (*memorymodels.Fact, error) {
	return nil, nil
}
func (m *mockFactReader) ListByRepository(ctx stdcontext.Context, repositoryID string) ([]*memorymodels.Fact, error) {
	return m.facts, m.err
}
func (m *mockFactReader) ListAll(ctx stdcontext.Context) ([]*memorymodels.Fact, error) {
	return m.facts, m.err
}

// --- Unit Tests ---

func TestMemoryContextReader_Unit(t *testing.T) {
	ctx := stdcontext.Background()
	repoID := "test_repo"

	t.Run("Empty repository", func(t *testing.T) {
		r := NewMemoryContextReader(
			&mockEventReader{},
			&mockDecisionReader{},
			&mockIntentReader{},
			&mockFactReader{},
		)

		data, err := r.ReadContext(ctx, repoID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if data.RepositoryID != repoID {
			t.Errorf("expected RepositoryID %q, got %q", repoID, data.RepositoryID)
		}
		if len(data.Events) != 0 || len(data.Decisions) != 0 || len(data.Intents) != 0 || len(data.Facts) != 0 {
			t.Errorf("expected all slices to be empty, got: %+v", data)
		}
	})

	t.Run("Populated repository", func(t *testing.T) {
		events := []*models.Event{{ID: "evt_1", Title: "Event 1"}}
		decisions := []*memorymodels.Decision{{ID: "dec_1", Title: "Decision 1"}}
		intents := []*memorymodels.Intent{{ID: "int_1", Description: "Intent 1"}}
		facts := []*memorymodels.Fact{{ID: "fact_1", Subject: "Fact 1"}}

		r := NewMemoryContextReader(
			&mockEventReader{events: events},
			&mockDecisionReader{decisions: decisions},
			&mockIntentReader{intents: intents},
			&mockFactReader{facts: facts},
		)

		data, err := r.ReadContext(ctx, repoID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(data.Events) != 1 || data.Events[0].ID != "evt_1" {
			t.Errorf("unexpected events: %+v", data.Events)
		}
		if len(data.Decisions) != 1 || data.Decisions[0].ID != "dec_1" {
			t.Errorf("unexpected decisions: %+v", data.Decisions)
		}
		if len(data.Intents) != 1 || data.Intents[0].ID != "int_1" {
			t.Errorf("unexpected intents: %+v", data.Intents)
		}
		if len(data.Facts) != 1 || data.Facts[0].ID != "fact_1" {
			t.Errorf("unexpected facts: %+v", data.Facts)
		}
	})

	t.Run("Reader failures", func(t *testing.T) {
		expectedErr := errors.New("database query error")
		r := NewMemoryContextReader(
			&mockEventReader{},
			&mockDecisionReader{err: expectedErr},
			&mockIntentReader{},
			&mockFactReader{},
		)

		_, err := r.ReadContext(ctx, repoID)
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("Partial results", func(t *testing.T) {
		events := []*models.Event{{ID: "evt_1", Title: "Event 1"}}
		decisions := []*memorymodels.Decision{{ID: "dec_1", Title: "Decision 1"}}

		r := NewMemoryContextReader(
			&mockEventReader{events: events},
			&mockDecisionReader{decisions: decisions},
			&mockIntentReader{},
			&mockFactReader{},
		)

		data, err := r.ReadContext(ctx, repoID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(data.Events) != 1 || len(data.Decisions) != 1 {
			t.Errorf("expected events and decisions to be populated, got: %+v", data)
		}
		if len(data.Intents) != 0 || len(data.Facts) != 0 {
			t.Errorf("expected intents and facts to be empty, got: %+v", data)
		}
	})
}

// --- Integration Tests ---

func TestMemoryContextReader_Integration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-context-reader-test-*")
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

	ctx := stdcontext.Background()
	repoID := "repo_context"

	// 1. Insert Repository
	_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repoID, "Repo Context", tempDir, "main")
	if err != nil {
		t.Fatalf("failed to insert repository: %v", err)
	}

	// 2. Insert Source
	_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_context", repoID, "adr", "docs/adr/0001.md", "Author", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("failed to insert source: %v", err)
	}

	// 3. Setup Writers & insert data
	eventStore := sqlite.NewEventStore(db)
	err = eventStore.UpsertEvent(ctx, &models.Event{
		ID:           "evt_context",
		RepositoryID: repoID,
		EventType:    "FEATURE",
		Title:        "Context Feature",
		SourceID:     "src_context",
		Timestamp:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to write event: %v", err)
	}

	decisionStore := memorystorage.NewSQLiteDecisionStore(db)
	err = decisionStore.UpsertDecision(ctx, &memorymodels.Decision{
		ID:           "dec_context",
		RepositoryID: repoID,
		Title:        "Use context pattern",
		Status:       "Accepted",
		SourceID:     "src_context",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to write decision: %v", err)
	}

	intentStore := memorystorage.NewSQLiteIntentStore(db)
	err = intentStore.UpsertIntent(ctx, &memorymodels.Intent{
		ID:           "int_context",
		RepositoryID: repoID,
		Description:  "Provide high-signal details",
		SourceID:     "src_context",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to write intent: %v", err)
	}

	factStore := memorystorage.NewSQLiteFactStore(db)
	err = factStore.UpsertFact(ctx, &memorymodels.Fact{
		ID:           "fact_context",
		RepositoryID: repoID,
		Subject:      "Authentication Service",
		Predicate:    "USES",
		Object:       "Redis",
		SourceID:     "src_context",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to write fact: %v", err)
	}

	// 4. Instantiation of composed Readers
	eventReader := storage.NewSQLiteEventReader(db)
	decisionReader := storage.NewSQLiteDecisionReader(db)
	intentReader := storage.NewSQLiteIntentReader(db)
	factReader := storage.NewSQLiteFactReader(db)

	contextReader := NewMemoryContextReader(eventReader, decisionReader, intentReader, factReader)

	// 5. Query context and assert
	data, err := contextReader.ReadContext(ctx, repoID)
	if err != nil {
		t.Fatalf("ReadContext failed: %v", err)
	}

	if data.RepositoryID != repoID {
		t.Errorf("expected RepositoryID %q, got %q", repoID, data.RepositoryID)
	}
	if len(data.Events) != 1 || data.Events[0].ID != "evt_context" {
		t.Errorf("unexpected event: %+v", data.Events)
	}
	if len(data.Decisions) != 1 || data.Decisions[0].ID != "dec_context" {
		t.Errorf("unexpected decision: %+v", data.Decisions)
	}
	if len(data.Intents) != 1 || data.Intents[0].ID != "int_context" {
		t.Errorf("unexpected intent: %+v", data.Intents)
	}
	if len(data.Facts) != 1 || data.Facts[0].ID != "fact_context" {
		t.Errorf("unexpected fact: %+v", data.Facts)
	}
}
