package compression

import (
	stdcontext "context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/reponerve/reponerve/internal/agent/onboarding"
	ctxengine "github.com/reponerve/reponerve/internal/context"
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	memorystorage "github.com/reponerve/reponerve/internal/memory/storage"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
	models "github.com/reponerve/reponerve/pkg/models"
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
		obs := onboarding.NewService(g)
		s := NewService(g, obs, nil)

		opts := CompressionOptions{
			MaxDecisions: 5,
			MaxIntents:   5,
			MaxFacts:     5,
			MaxEvents:    5,
		}

		cCtx, err := s.Compress(ctx, repoID, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cCtx.RepositoryID != repoID {
			t.Errorf("expected RepositoryID %q, got %q", repoID, cCtx.RepositoryID)
		}
		if len(cCtx.Decisions) != 0 || len(cCtx.Intents) != 0 || len(cCtx.Facts) != 0 || len(cCtx.Events) != 0 {
			t.Errorf("expected all slices to be empty, got Decisions:%d, Intents:%d, Facts:%d, Events:%d",
				len(cCtx.Decisions), len(cCtx.Intents), len(cCtx.Facts), len(cCtx.Events))
		}
	})

	t.Run("Generator failures", func(t *testing.T) {
		expectedErr := errors.New("read error")
		r := &mockContextReader{
			err: expectedErr,
		}
		g := ctxengine.NewGenerator(r)
		obs := onboarding.NewService(g)
		s := NewService(g, obs, nil)

		_, err := s.Compress(ctx, repoID, CompressionOptions{})
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
		obs := onboarding.NewService(g)
		s := NewService(g, obs, nil)

		opts := CompressionOptions{
			MaxDecisions: 5,
			MaxIntents:   5,
			MaxFacts:     5,
			MaxEvents:    5,
		}

		cCtx, err := s.Compress(ctx, repoID, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cCtx.Facts) != 1 || cCtx.Facts[0].ID != "fact_1" {
			t.Errorf("expected 1 fact 'fact_1', got %d elements", len(cCtx.Facts))
		}
		if len(cCtx.Decisions) != 0 || len(cCtx.Intents) != 0 || len(cCtx.Events) != 0 {
			t.Errorf("expected other lists to be empty, got Decisions:%d, Intents:%d, Events:%d",
				len(cCtx.Decisions), len(cCtx.Intents), len(cCtx.Events))
		}
	})

	t.Run("Limit enforcement and sorting", func(t *testing.T) {
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
		obs := onboarding.NewService(g)
		s := NewService(g, obs, nil)

		// Max limits enforce truncation to 2 elements for all types
		opts := CompressionOptions{
			MaxDecisions: 2,
			MaxIntents:   2,
			MaxFacts:     2,
			MaxEvents:    2,
		}

		cCtx, err := s.Compress(ctx, repoID, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Expected sorted decisions: dec_c, dec_d, dec_a, dec_b -> truncated to 2: dec_c, dec_d
		expectedDecs := []string{"dec_c", "dec_d"}
		if len(cCtx.Decisions) != 2 {
			t.Fatalf("expected 2 decisions, got %d", len(cCtx.Decisions))
		}
		for i, id := range expectedDecs {
			if cCtx.Decisions[i].ID != id {
				t.Errorf("decisions truncated incorrectly at %d: expected %s, got %s", i, id, cCtx.Decisions[i].ID)
			}
		}

		// Expected sorted intents: int_c, int_d, int_a, int_b -> truncated to 2: int_c, int_d
		expectedInts := []string{"int_c", "int_d"}
		if len(cCtx.Intents) != 2 {
			t.Fatalf("expected 2 intents, got %d", len(cCtx.Intents))
		}
		for i, id := range expectedInts {
			if cCtx.Intents[i].ID != id {
				t.Errorf("intents truncated incorrectly at %d: expected %s, got %s", i, id, cCtx.Intents[i].ID)
			}
		}

		// Expected sorted events: evt_c, evt_d, evt_a, evt_b -> truncated to 2: evt_c, evt_d
		expectedEvts := []string{"evt_c", "evt_d"}
		if len(cCtx.Events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(cCtx.Events))
		}
		for i, id := range expectedEvts {
			if cCtx.Events[i].ID != id {
				t.Errorf("events truncated incorrectly at %d: expected %s, got %s", i, id, cCtx.Events[i].ID)
			}
		}

		// Expected sorted facts: fact_a, fact_d, fact_e, fact_b, fact_c -> truncated to 2: fact_a, fact_d
		expectedFacts := []string{"fact_a", "fact_d"}
		if len(cCtx.Facts) != 2 {
			t.Fatalf("expected 2 facts, got %d", len(cCtx.Facts))
		}
		for i, id := range expectedFacts {
			if cCtx.Facts[i].ID != id {
				t.Errorf("facts truncated incorrectly at %d: expected %s, got %s", i, id, cCtx.Facts[i].ID)
			}
		}
	})

	t.Run("Zero and negative limits", func(t *testing.T) {
		r := &mockContextReader{
			data: &ctxengine.ContextData{
				RepositoryID: repoID,
				Decisions: []*memorymodels.Decision{
					{ID: "dec_1"},
				},
				Intents: []*memorymodels.Intent{
					{ID: "int_1"},
				},
				Facts: []*memorymodels.Fact{
					{ID: "fact_1"},
				},
				Events: []*models.Event{
					{ID: "evt_1"},
				},
			},
		}

		g := ctxengine.NewGenerator(r)
		obs := onboarding.NewService(g)
		s := NewService(g, obs, nil)

		opts := CompressionOptions{
			MaxDecisions: 0,
			MaxIntents:   -1,
			MaxFacts:     0,
			MaxEvents:    -5,
		}

		cCtx, err := s.Compress(ctx, repoID, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cCtx.Decisions) != 0 || cCtx.Decisions == nil {
			t.Errorf("expected decisions slice to be empty and non-nil, got: %v", cCtx.Decisions)
		}
		if len(cCtx.Intents) != 0 || cCtx.Intents == nil {
			t.Errorf("expected intents slice to be empty and non-nil, got: %v", cCtx.Intents)
		}
		if len(cCtx.Facts) != 0 || cCtx.Facts == nil {
			t.Errorf("expected facts slice to be empty and non-nil, got: %v", cCtx.Facts)
		}
		if len(cCtx.Events) != 0 || cCtx.Events == nil {
			t.Errorf("expected events slice to be empty and non-nil, got: %v", cCtx.Events)
		}
	})

	t.Run("Topic relevance ranking", func(t *testing.T) {
		r := &mockContextReader{
			data: &ctxengine.ContextData{
				RepositoryID: repoID,
				Decisions: []*memorymodels.Decision{
					{ID: "dec_redis", Title: "Use Redis for caching"},
					{ID: "dec_sqlite", Title: "Local-first SQLite storage"},
				},
			},
		}
		g := ctxengine.NewGenerator(r)
		obs := onboarding.NewService(g)
		s := NewService(g, obs, nil)

		cCtx, err := s.Compress(ctx, repoID, CompressionOptions{
			Topic:        "sqlite",
			MaxDecisions: 1,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cCtx.Decisions) != 1 || cCtx.Decisions[0].ID != "dec_sqlite" {
			t.Fatalf("expected sqlite decision first, got %+v", cCtx.Decisions)
		}
	})
}

// --- Integration Tests ---

func TestService_Integration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-compression-integration-test-*")
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
	repoID := "repo_compression"

	// Insert Repository
	_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repoID, "Repo Compression", tempDir, "main")
	if err != nil {
		t.Fatalf("failed to insert repository: %v", err)
	}

	// Insert Source
	_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_compression", repoID, "adr", "docs/adr/0001.md", "Author", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("failed to insert source: %v", err)
	}

	// Setup Stores & insert multiple data entities to test truncation
	eventStore := sqlite.NewEventStore(db)
	for _, id := range []string{"evt_1", "evt_2", "evt_3"} {
		timestamp := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		if id == "evt_2" {
			timestamp = time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
		} else if id == "evt_3" {
			timestamp = time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
		}
		err = eventStore.UpsertEvent(ctx, &models.Event{
			ID:           id,
			RepositoryID: repoID,
			EventType:    "FEATURE",
			Title:        "Feature " + id,
			SourceID:     "src_compression",
			Timestamp:    timestamp,
		})
		if err != nil {
			t.Fatalf("failed to write event: %v", err)
		}
	}

	decisionStore := memorystorage.NewSQLiteDecisionStore(db)
	for _, id := range []string{"dec_1", "dec_2", "dec_3"} {
		createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		if id == "dec_2" {
			createdAt = time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
		} else if id == "dec_3" {
			createdAt = time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
		}
		err = decisionStore.UpsertDecision(ctx, &memorymodels.Decision{
			ID:           id,
			RepositoryID: repoID,
			Title:        "Decision " + id,
			Status:       "Accepted",
			SourceID:     "src_compression",
			CreatedAt:    createdAt,
		})
		if err != nil {
			t.Fatalf("failed to write decision: %v", err)
		}
	}

	// Instantiation of composed Readers, ContextReader and Generator
	eventReader := storage.NewSQLiteEventReader(db)
	decisionReader := storage.NewSQLiteDecisionReader(db)
	intentReader := storage.NewSQLiteIntentReader(db)
	factReader := storage.NewSQLiteFactReader(db)

	contextReader := ctxengine.NewMemoryContextReader(eventReader, decisionReader, intentReader, factReader)
	generator := ctxengine.NewGenerator(contextReader)

	// Instantiation of services
	obs := onboarding.NewService(generator)
	service := NewService(generator, obs, nil)

	// Compress with limit 2
	opts := CompressionOptions{
		MaxDecisions: 2,
		MaxIntents:   2,
		MaxFacts:     2,
		MaxEvents:    2,
	}

	cCtx, err := service.Compress(ctx, repoID, opts)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	if cCtx.RepositoryID != repoID {
		t.Errorf("expected RepositoryID %q, got %q", repoID, cCtx.RepositoryID)
	}

	// Verify decisions: dec_3 (most recent), dec_2 (next most recent)
	if len(cCtx.Decisions) != 2 {
		t.Fatalf("expected 2 decisions, got %d", len(cCtx.Decisions))
	}
	if cCtx.Decisions[0].ID != "dec_3" || cCtx.Decisions[1].ID != "dec_2" {
		t.Errorf("expected dec_3, dec_2, got %s, %s", cCtx.Decisions[0].ID, cCtx.Decisions[1].ID)
	}

	// Verify events: evt_3 (most recent), evt_2 (next most recent)
	if len(cCtx.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(cCtx.Events))
	}
	if cCtx.Events[0].ID != "evt_3" || cCtx.Events[1].ID != "evt_2" {
		t.Errorf("expected evt_3, evt_2, got %s, %s", cCtx.Events[0].ID, cCtx.Events[1].ID)
	}
}
