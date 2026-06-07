package onboarding

import (
	stdcontext "context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	ctxengine "reponerve/internal/context"
	memorymodels "reponerve/internal/memory/models"
	memorystorage "reponerve/internal/memory/storage"
	"reponerve/internal/query/storage"
	"reponerve/internal/storage/migrations"
	"reponerve/internal/storage/sqlite"
	models "reponerve/pkg/models"
)

// --- Mock Context Reader ---

type mockContextReader struct {
	data *ctxengine.ContextData
	err  error
}

func (m *mockContextReader) ReadContext(ctx stdcontext.Context, repositoryID string) (*ctxengine.ContextData, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

// --- Service Unit Tests ---

func TestService_Unit(t *testing.T) {
	ctx := stdcontext.Background()
	repoID := "test_repo"

	t.Run("Empty repository", func(t *testing.T) {
		r := &mockContextReader{
			data: &ctxengine.ContextData{
				RepositoryID: repoID,
			},
		}
		g := ctxengine.NewGenerator(r)
		s := NewService(g)

		pkg, err := s.Generate(ctx, repoID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if pkg.RepositoryID != repoID {
			t.Errorf("expected RepositoryID %q, got %q", repoID, pkg.RepositoryID)
		}
		if len(pkg.Decisions) != 0 || len(pkg.Intents) != 0 || len(pkg.Facts) != 0 || len(pkg.Events) != 0 {
			t.Errorf("expected all slices to be empty, got Decisions:%d, Intents:%d, Facts:%d, Events:%d",
				len(pkg.Decisions), len(pkg.Intents), len(pkg.Facts), len(pkg.Events))
		}
		expectedSummary := "Repository Onboarding:\n- 0 decisions\n- 0 intents\n- 0 facts\n- 0 events"
		if pkg.Summary != expectedSummary {
			t.Errorf("expected summary %q, got %q", expectedSummary, pkg.Summary)
		}
	})

	t.Run("Reader failures", func(t *testing.T) {
		expectedErr := errors.New("read error")
		r := &mockContextReader{
			err: expectedErr,
		}
		g := ctxengine.NewGenerator(r)
		s := NewService(g)

		_, err := s.Generate(ctx, repoID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "read error") {
			t.Errorf("expected error message to contain 'read error', got: %v", err)
		}
	})

	t.Run("Partial data scenario", func(t *testing.T) {
		r := &mockContextReader{
			data: &ctxengine.ContextData{
				RepositoryID: repoID,
				Facts: []*memorymodels.Fact{
					{ID: "fact_1", Subject: "AuthService", Predicate: "USES", Object: "PostgreSQL"},
				},
			},
		}
		g := ctxengine.NewGenerator(r)
		s := NewService(g)

		pkg, err := s.Generate(ctx, repoID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(pkg.Facts) != 1 || pkg.Facts[0].ID != "fact_1" {
			t.Errorf("expected 1 fact 'fact_1', got %d elements", len(pkg.Facts))
		}
		if len(pkg.Decisions) != 0 || len(pkg.Intents) != 0 || len(pkg.Events) != 0 {
			t.Errorf("expected other lists to be empty, got Decisions:%d, Intents:%d, Events:%d",
				len(pkg.Decisions), len(pkg.Intents), len(pkg.Events))
		}
		expectedSummary := "Repository Onboarding:\n- 0 decisions\n- 0 intents\n- 1 facts\n- 0 events"
		if pkg.Summary != expectedSummary {
			t.Errorf("expected summary %q, got %q", expectedSummary, pkg.Summary)
		}
	})

	t.Run("Deterministic ordering", func(t *testing.T) {
		t1 := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
		t2 := time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)
		t3 := time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC)

		// Decisions: CreatedAt descending, ID descending fallback
		decisions := []*memorymodels.Decision{
			{ID: "dec_a", CreatedAt: t2},
			{ID: "dec_b", CreatedAt: t1},
			{ID: "dec_c", CreatedAt: t3},
			{ID: "dec_d", CreatedAt: t2},
		}

		// Intents: CreatedAt descending, ID descending fallback
		intents := []*memorymodels.Intent{
			{ID: "int_a", CreatedAt: t2},
			{ID: "int_b", CreatedAt: t1},
			{ID: "int_c", CreatedAt: t3},
			{ID: "int_d", CreatedAt: t2},
		}

		// Events: Timestamp descending, ID descending fallback
		events := []*models.Event{
			{ID: "evt_a", Timestamp: t2},
			{ID: "evt_b", Timestamp: t1},
			{ID: "evt_c", Timestamp: t3},
			{ID: "evt_d", Timestamp: t2},
		}

		// Facts: Alphabetical by Subject ascending, ID ascending fallback
		facts := []*memorymodels.Fact{
			{ID: "fact_c", Subject: "Zoo"},
			{ID: "fact_a", Subject: "Apple"},
			{ID: "fact_b", Subject: "Banana"},
			{ID: "fact_d", Subject: "Apple"},
			{ID: "fact_e", Subject: "Apple"},
		}

		r := &mockContextReader{
			data: &ctxengine.ContextData{
				RepositoryID: repoID,
				Decisions:    decisions,
				Intents:      intents,
				Events:       events,
				Facts:        facts,
			},
		}

		g := ctxengine.NewGenerator(r)
		s := NewService(g)

		pkg, err := s.Generate(ctx, repoID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify Decisions (dec_c, dec_d, dec_a, dec_b)
		expectedDecs := []string{"dec_c", "dec_d", "dec_a", "dec_b"}
		for i, id := range expectedDecs {
			if pkg.Decisions[i].ID != id {
				t.Errorf("decisions sorted incorrectly at %d: expected %s, got %s", i, id, pkg.Decisions[i].ID)
			}
		}

		// Verify Intents (int_c, int_d, int_a, int_b)
		expectedInts := []string{"int_c", "int_d", "int_a", "int_b"}
		for i, id := range expectedInts {
			if pkg.Intents[i].ID != id {
				t.Errorf("intents sorted incorrectly at %d: expected %s, got %s", i, id, pkg.Intents[i].ID)
			}
		}

		// Verify Events (evt_c, evt_d, evt_a, evt_b)
		expectedEvts := []string{"evt_c", "evt_d", "evt_a", "evt_b"}
		for i, id := range expectedEvts {
			if pkg.Events[i].ID != id {
				t.Errorf("events sorted incorrectly at %d: expected %s, got %s", i, id, pkg.Events[i].ID)
			}
		}

		// Verify Facts (fact_a, fact_d, fact_e, fact_b, fact_c)
		expectedFacts := []string{"fact_a", "fact_d", "fact_e", "fact_b", "fact_c"}
		for i, id := range expectedFacts {
			if pkg.Facts[i].ID != id {
				t.Errorf("facts sorted incorrectly at %d: expected %s, got %s", i, id, pkg.Facts[i].ID)
			}
		}

		expectedSummary := "Repository Onboarding:\n- 4 decisions\n- 4 intents\n- 5 facts\n- 4 events"
		if pkg.Summary != expectedSummary {
			t.Errorf("expected summary %q, got %q", expectedSummary, pkg.Summary)
		}
	})
}

// --- Integration Tests ---

func TestService_Integration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-onboarding-integration-test-*")
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
	repoID := "repo_onboarding"

	// Insert Repository
	_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repoID, "Repo Onboarding", tempDir, "main")
	if err != nil {
		t.Fatalf("failed to insert repository: %v", err)
	}

	// Insert Source
	_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_onboarding", repoID, "adr", "docs/adr/0001.md", "Author", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("failed to insert source: %v", err)
	}

	// Setup Stores & insert data
	eventStore := sqlite.NewEventStore(db)
	err = eventStore.UpsertEvent(ctx, &models.Event{
		ID:           "evt_1",
		RepositoryID: repoID,
		EventType:    "FEATURE",
		Title:        "Feature A",
		SourceID:     "src_onboarding",
		Timestamp:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
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
		SourceID:     "src_onboarding",
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
		SourceID:     "src_onboarding",
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
		SourceID:     "src_onboarding",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to write fact: %v", err)
	}

	// Instantiation of composed Readers, ContextReader and Generator
	eventReader := storage.NewSQLiteEventReader(db)
	decisionReader := storage.NewSQLiteDecisionReader(db)
	intentReader := storage.NewSQLiteIntentReader(db)
	factReader := storage.NewSQLiteFactReader(db)

	contextReader := ctxengine.NewMemoryContextReader(eventReader, decisionReader, intentReader, factReader)
	generator := ctxengine.NewGenerator(contextReader)

	// Instantiation of Onboarding Service
	service := NewService(generator)

	// Generate and verify onboarding package
	pkg, err := service.Generate(ctx, repoID)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if pkg.RepositoryID != repoID {
		t.Errorf("expected RepositoryID %q, got %q", repoID, pkg.RepositoryID)
	}
	if len(pkg.Decisions) != 1 || pkg.Decisions[0].ID != "dec_1" {
		t.Errorf("expected decision dec_1, got %v", pkg.Decisions)
	}
	if len(pkg.Intents) != 1 || pkg.Intents[0].ID != "int_1" {
		t.Errorf("expected intent int_1, got %v", pkg.Intents)
	}
	if len(pkg.Facts) != 1 || pkg.Facts[0].ID != "fact_1" {
		t.Errorf("expected fact fact_1, got %v", pkg.Facts)
	}
	if len(pkg.Events) != 1 || pkg.Events[0].ID != "evt_1" {
		t.Errorf("expected event evt_1, got %v", pkg.Events)
	}

	expectedSummary := "Repository Onboarding:\n- 1 decisions\n- 1 intents\n- 1 facts\n- 1 events"
	if pkg.Summary != expectedSummary {
		t.Errorf("expected summary %q, got %q", expectedSummary, pkg.Summary)
	}
}
