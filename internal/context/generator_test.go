package context

import (
	stdcontext "context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	memorystorage "github.com/reponerve/reponerve/internal/memory/storage"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
	models "github.com/reponerve/reponerve/pkg/models"
)

// --- Mock Context Reader ---

type mockContextReader struct {
	data *ContextData
	err  error
}

func (m *mockContextReader) ReadContext(ctx stdcontext.Context, repositoryID string) (*ContextData, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

// --- Generator Unit Tests ---

func TestGenerator_Unit(t *testing.T) {
	ctx := stdcontext.Background()
	repoID := "test_repo"

	t.Run("Empty repository", func(t *testing.T) {
		r := &mockContextReader{
			data: &ContextData{
				RepositoryID: repoID,
			},
		}
		g := NewGenerator(r)

		rc, err := g.Generate(ctx, repoID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rc.RepositoryID != repoID {
			t.Errorf("expected RepositoryID %q, got %q", repoID, rc.RepositoryID)
		}
		if len(rc.Decisions) != 0 || len(rc.Intents) != 0 || len(rc.Facts) != 0 || len(rc.Events) != 0 {
			t.Errorf("expected all slices to be empty, got Decisions:%d, Intents:%d, Facts:%d, Events:%d",
				len(rc.Decisions), len(rc.Intents), len(rc.Facts), len(rc.Events))
		}
		if rc.GeneratedAt.IsZero() {
			t.Error("expected GeneratedAt to be non-zero")
		}
	})

	t.Run("Reader failures", func(t *testing.T) {
		expectedErr := errors.New("read error")
		r := &mockContextReader{
			err: expectedErr,
		}
		g := NewGenerator(r)

		_, err := g.Generate(ctx, repoID)
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("Ordering guarantees", func(t *testing.T) {
		// Prepare unsorted data
		t1 := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
		t2 := time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)
		t3 := time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC)

		// Decisions: Most recent first (CreatedAt descending, ID descending fallback)
		decisions := []*memorymodels.Decision{
			{ID: "dec_a", CreatedAt: t2},
			{ID: "dec_b", CreatedAt: t1},
			{ID: "dec_c", CreatedAt: t3},
			{ID: "dec_d", CreatedAt: t2}, // same time as dec_a, dec_d should come first if sorting descending by ID
		}

		// Intents: Most recent first (CreatedAt descending, ID descending fallback)
		intents := []*memorymodels.Intent{
			{ID: "int_a", CreatedAt: t2},
			{ID: "int_b", CreatedAt: t1},
			{ID: "int_c", CreatedAt: t3},
			{ID: "int_d", CreatedAt: t2}, // same time as int_a, int_d should come first
		}

		// Events: Most recent first (Timestamp descending, ID descending fallback)
		events := []*models.Event{
			{ID: "evt_a", Timestamp: t2},
			{ID: "evt_b", Timestamp: t1},
			{ID: "evt_c", Timestamp: t3},
			{ID: "evt_d", Timestamp: t2}, // same time as evt_a, evt_d should come first
		}

		// Facts: Alphabetical by Subject (ascending, ID ascending fallback)
		facts := []*memorymodels.Fact{
			{ID: "fact_c", Subject: "Zoo"},
			{ID: "fact_a", Subject: "Apple"},
			{ID: "fact_b", Subject: "Banana"},
			{ID: "fact_d", Subject: "Apple"}, // same subject as Apple, fact_a should come first (ID alphabetical ascending: fact_a < fact_d)
			{ID: "fact_e", Subject: "Apple"}, // same subject as Apple, fact_d should come next (fact_d < fact_e)
		}

		r := &mockContextReader{
			data: &ContextData{
				RepositoryID: repoID,
				Decisions:    decisions,
				Intents:      intents,
				Events:       events,
				Facts:        facts,
			},
		}

		g := NewGenerator(r)
		rc, err := g.Generate(ctx, repoID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify decisions sorting (dec_c [t3], dec_d [t2], dec_a [t2], dec_b [t1])
		expectedDecs := []string{"dec_c", "dec_d", "dec_a", "dec_b"}
		for i, id := range expectedDecs {
			if rc.Decisions[i].ID != id {
				t.Errorf("decisions sorted incorrectly at index %d: expected %s, got %s", i, id, rc.Decisions[i].ID)
			}
		}

		// Verify intents sorting (int_c [t3], int_d [t2], int_a [t2], int_b [t1])
		expectedInts := []string{"int_c", "int_d", "int_a", "int_b"}
		for i, id := range expectedInts {
			if rc.Intents[i].ID != id {
				t.Errorf("intents sorted incorrectly at index %d: expected %s, got %s", i, id, rc.Intents[i].ID)
			}
		}

		// Verify events sorting (evt_c [t3], evt_d [t2], evt_a [t2], evt_b [t1])
		expectedEvts := []string{"evt_c", "evt_d", "evt_a", "evt_b"}
		for i, id := range expectedEvts {
			if rc.Events[i].ID != id {
				t.Errorf("events sorted incorrectly at index %d: expected %s, got %s", i, id, rc.Events[i].ID)
			}
		}

		// Verify facts sorting (fact_a [Apple], fact_d [Apple], fact_e [Apple], fact_b [Banana], fact_c [Zoo])
		expectedFacts := []string{"fact_a", "fact_d", "fact_e", "fact_b", "fact_c"}
		for i, id := range expectedFacts {
			if rc.Facts[i].ID != id {
				t.Errorf("facts sorted incorrectly at index %d: expected %s, got %s", i, id, rc.Facts[i].ID)
			}
		}
	})
}

// --- Integration Tests ---

func TestGenerator_Integration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-generator-integration-test-*")
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

	// 3. Setup Stores & insert data
	eventStore := sqlite.NewEventStore(db)
	err = eventStore.UpsertEvent(ctx, &models.Event{
		ID:           "evt_1",
		RepositoryID: repoID,
		EventType:    "FEATURE",
		Title:        "Feature A",
		SourceID:     "src_context",
		Timestamp:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to write event: %v", err)
	}
	err = eventStore.UpsertEvent(ctx, &models.Event{
		ID:           "evt_2",
		RepositoryID: repoID,
		EventType:    "FEATURE",
		Title:        "Feature B",
		SourceID:     "src_context",
		Timestamp:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to write event: %v", err)
	}

	decisionStore := memorystorage.NewSQLiteDecisionStore(db)
	err = decisionStore.UpsertDecision(ctx, &memorymodels.Decision{
		ID:           "dec_1",
		RepositoryID: repoID,
		Title:        "Decision A",
		Status:       "Accepted",
		SourceID:     "src_context",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to write decision: %v", err)
	}

	intentStore := memorystorage.NewSQLiteIntentStore(db)
	err = intentStore.UpsertIntent(ctx, &memorymodels.Intent{
		ID:           "int_1",
		RepositoryID: repoID,
		Description:  "Intent A",
		SourceID:     "src_context",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to write intent: %v", err)
	}

	factStore := memorystorage.NewSQLiteFactStore(db)
	err = factStore.UpsertFact(ctx, &memorymodels.Fact{
		ID:           "fact_1",
		RepositoryID: repoID,
		Subject:      "Service B",
		Predicate:    "DEPENDS_ON",
		Object:       "Service A",
		SourceID:     "src_context",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to write fact: %v", err)
	}
	err = factStore.UpsertFact(ctx, &memorymodels.Fact{
		ID:           "fact_2",
		RepositoryID: repoID,
		Subject:      "Service A",
		Predicate:    "USES",
		Object:       "Redis",
		SourceID:     "src_context",
		CreatedAt:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to write fact: %v", err)
	}

	// 4. Instantiation of composed Readers, ContextReader and Generator
	eventReader := storage.NewSQLiteEventReader(db)
	decisionReader := storage.NewSQLiteDecisionReader(db)
	intentReader := storage.NewSQLiteIntentReader(db)
	factReader := storage.NewSQLiteFactReader(db)

	contextReader := NewMemoryContextReader(eventReader, decisionReader, intentReader, factReader)
	generator := NewGenerator(contextReader)

	// 5. Query context and assert sorting
	rc, err := generator.Generate(ctx, repoID)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if rc.RepositoryID != repoID {
		t.Errorf("expected RepositoryID %q, got %q", repoID, rc.RepositoryID)
	}

	// Verify Events (evt_2 [Jan 2] then evt_1 [Jan 1])
	if len(rc.Events) != 2 || rc.Events[0].ID != "evt_2" || rc.Events[1].ID != "evt_1" {
		t.Errorf("events sorted incorrectly: %+v", rc.Events)
	}

	// Verify Facts (fact_2 [Service A] then fact_1 [Service B])
	if len(rc.Facts) != 2 || rc.Facts[0].ID != "fact_2" || rc.Facts[1].ID != "fact_1" {
		t.Errorf("facts sorted incorrectly: %+v", rc.Facts)
	}
}
