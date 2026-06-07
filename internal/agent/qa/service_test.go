package qa

import (
	stdcontext "context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"reponerve/internal/agent/guidance"
	"reponerve/internal/agent/impact"
	"reponerve/internal/agent/onboarding"
	ctxengine "reponerve/internal/context"
	memorymodels "reponerve/internal/memory/models"
	memorystorage "reponerve/internal/memory/storage"
	"reponerve/internal/query/storage"
	"reponerve/internal/storage/migrations"
	"reponerve/internal/storage/sqlite"
	models "reponerve/pkg/models"
)

// --- Mock Readers for Q&A Unit Testing ---

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

type mockDecisionReader struct {
	dec *memorymodels.Decision
	err error
}

func (m *mockDecisionReader) GetByID(ctx stdcontext.Context, id string) (*memorymodels.Decision, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.dec == nil || m.dec.ID != id {
		return nil, sql.ErrNoRows
	}
	return m.dec, nil
}
func (m *mockDecisionReader) ListByRepository(ctx stdcontext.Context, repositoryID string) ([]*memorymodels.Decision, error) {
	return nil, nil
}
func (m *mockDecisionReader) ListAll(ctx stdcontext.Context) ([]*memorymodels.Decision, error) {
	return nil, nil
}

type mockIntentReader struct {
	it  *memorymodels.Intent
	err error
}

func (m *mockIntentReader) GetByID(ctx stdcontext.Context, id string) (*memorymodels.Intent, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.it == nil || m.it.ID != id {
		return nil, sql.ErrNoRows
	}
	return m.it, nil
}
func (m *mockIntentReader) ListByRepository(ctx stdcontext.Context, repositoryID string) ([]*memorymodels.Intent, error) {
	return nil, nil
}
func (m *mockIntentReader) ListAll(ctx stdcontext.Context) ([]*memorymodels.Intent, error) {
	return nil, nil
}

type mockFactReader struct {
	f   *memorymodels.Fact
	err error
}

func (m *mockFactReader) GetByID(ctx stdcontext.Context, id string) (*memorymodels.Fact, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.f == nil || m.f.ID != id {
		return nil, sql.ErrNoRows
	}
	return m.f, nil
}
func (m *mockFactReader) ListByRepository(ctx stdcontext.Context, repositoryID string) ([]*memorymodels.Fact, error) {
	return nil, nil
}
func (m *mockFactReader) ListAll(ctx stdcontext.Context) ([]*memorymodels.Fact, error) {
	return nil, nil
}

type mockEventReader struct {
	evt *models.Event
	err error
}

func (m *mockEventReader) GetByID(ctx stdcontext.Context, id string) (*models.Event, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.evt == nil || m.evt.ID != id {
		return nil, sql.ErrNoRows
	}
	return m.evt, nil
}
func (m *mockEventReader) ListByRepository(ctx stdcontext.Context, repositoryID string) ([]*models.Event, error) {
	return nil, nil
}
func (m *mockEventReader) ListAll(ctx stdcontext.Context) ([]*models.Event, error) {
	return nil, nil
}

type mockRelationshipReader struct {
	rels []*memorymodels.Relationship
	err  error
}

func (m *mockRelationshipReader) GetByID(ctx stdcontext.Context, id string) (*memorymodels.Relationship, error) {
	return nil, nil
}
func (m *mockRelationshipReader) ListByRepository(ctx stdcontext.Context, repositoryID string) ([]*memorymodels.Relationship, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.rels, nil
}
func (m *mockRelationshipReader) ListAll(ctx stdcontext.Context) ([]*memorymodels.Relationship, error) {
	return nil, nil
}

// --- Unit Tests ---

func TestService_Unit(t *testing.T) {
	ctx := stdcontext.Background()
	repoID := "test_repo"

	// Setup mock services
	ctxReader := &mockContextReader{data: &ctxengine.ContextData{RepositoryID: repoID}}
	generator := ctxengine.NewGenerator(ctxReader)
	obs := onboarding.NewService(generator)

	dr := &mockDecisionReader{dec: &memorymodels.Decision{ID: "dec_1", RepositoryID: repoID, Title: "Decision A"}}
	ir := &mockIntentReader{it: &memorymodels.Intent{ID: "int_1", RepositoryID: repoID, Description: "Intent A"}}
	fr := &mockFactReader{f: &memorymodels.Fact{ID: "fact_1", RepositoryID: repoID, Subject: "Fact A"}}
	er := &mockEventReader{evt: &models.Event{ID: "evt_1", RepositoryID: repoID, Title: "Event A"}}
	rr := &mockRelationshipReader{rels: []*memorymodels.Relationship{
		{ID: "r1", Type: "INTENT_DRIVES_DECISION", FromID: "int_1", ToID: "dec_1"},
		{ID: "r2", Type: "FACT_SUPPORTS_DECISION", FromID: "fact_1", ToID: "dec_1"},
		{ID: "r3", Type: "DECISION_RESULTS_IN_EVENT", FromID: "dec_1", ToID: "evt_1"},
	}}

	gs := guidance.NewService(dr, ir, fr, er, rr)
	is := impact.NewService(dr, ir, fr, er, rr)

	qaService := NewService(obs, gs, is)

	t.Run("Repository overview questions", func(t *testing.T) {
		questions := []string{
			"What is this repository?",
			"Show repository overview",
			"what is this repository",
		}
		for _, qText := range questions {
			ans, err := qaService.Answer(ctx, repoID, Question{Text: qText})
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", qText, err)
			}
			if ans.Question != qText {
				t.Errorf("expected Question %q, got %q", qText, ans.Question)
			}
			pkg, ok := ans.Result.(*onboarding.OnboardingPackage)
			if !ok {
				t.Errorf("expected *onboarding.OnboardingPackage result, got %T", ans.Result)
			}
			if pkg.RepositoryID != repoID {
				t.Errorf("expected package RepositoryID %q, got %q", repoID, pkg.RepositoryID)
			}
		}
	})

	t.Run("Decision guidance questions", func(t *testing.T) {
		questions := []string{
			"Why was decision dec_1 made?",
			"What supports decision dec_1",
		}
		for _, qText := range questions {
			ans, err := qaService.Answer(ctx, repoID, Question{Text: qText})
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", qText, err)
			}
			g, ok := ans.Result.(*guidance.Guidance)
			if !ok {
				t.Errorf("expected *guidance.Guidance result, got %T", ans.Result)
			}
			if g.EntityID != "dec_1" {
				t.Errorf("expected entity ID dec_1, got %q", g.EntityID)
			}
			if len(g.Reasons) != 1 || g.Reasons[0] != "Intent A" {
				t.Errorf("expected intent A under reasons, got %v", g.Reasons)
			}
		}
	})

	t.Run("Event guidance questions", func(t *testing.T) {
		ans, err := qaService.Answer(ctx, repoID, Question{Text: "what caused event evt_1?"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		g, ok := ans.Result.(*guidance.Guidance)
		if !ok {
			t.Errorf("expected *guidance.Guidance result, got %T", ans.Result)
		}
		if g.EntityID != "evt_1" {
			t.Errorf("expected event ID evt_1, got %q", g.EntityID)
		}
		if len(g.Reasons) < 2 || g.Reasons[0] != "Caused by decision: Decision A" {
			t.Errorf("unexpected event reasons: %v", g.Reasons)
		}
	})

	t.Run("Impact questions", func(t *testing.T) {
		t.Run("Decision impact", func(t *testing.T) {
			ans, err := qaService.Answer(ctx, repoID, Question{Text: "what happens if decision dec_1 changes?"})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			rep, ok := ans.Result.(*impact.ImpactReport)
			if !ok {
				t.Errorf("expected *impact.ImpactReport result, got %T", ans.Result)
			}
			if rep.EntityID != "dec_1" {
				t.Errorf("expected EntityID dec_1, got %q", rep.EntityID)
			}
		})

		t.Run("Fact impact", func(t *testing.T) {
			ans, err := qaService.Answer(ctx, repoID, Question{Text: "What depends on fact fact_1?"})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			rep, ok := ans.Result.(*impact.ImpactReport)
			if !ok {
				t.Errorf("expected *impact.ImpactReport result, got %T", ans.Result)
			}
			if rep.EntityID != "fact_1" {
				t.Errorf("expected EntityID fact_1, got %q", rep.EntityID)
			}
		})
	})

	t.Run("Unsupported questions", func(t *testing.T) {
		_, err := qaService.Answer(ctx, repoID, Question{Text: "Who is the lead developer of this repository?"})
		if err == nil {
			t.Error("expected error for unsupported question, got nil")
		}
		if !strings.Contains(err.Error(), "unknown question") {
			t.Errorf("expected error message to contain 'unknown question', got: %v", err)
		}
	})

	t.Run("Missing entities", func(t *testing.T) {
		_, err := qaService.Answer(ctx, repoID, Question{Text: "why was decision dec_missing made?"})
		if err == nil {
			t.Error("expected error for missing decision, got nil")
		}
		if !strings.Contains(err.Error(), "decision with ID \"dec_missing\" not found") {
			t.Errorf("expected error to contain 'not found', got: %v", err)
		}
	})
}

// --- Integration Tests ---

func TestService_Integration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-qa-integration-test-*")
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
	repoID := "repo_qa"

	// 1. Insert Repository
	_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repoID, "Repo QA", tempDir, "main")
	if err != nil {
		t.Fatalf("failed to insert repository: %v", err)
	}

	// 2. Insert Source
	_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_qa", repoID, "adr", "docs/adr/0001.md", "Author", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("failed to insert source: %v", err)
	}

	// 3. Write data to stores
	eventStore := sqlite.NewEventStore(db)
	err = eventStore.UpsertEvent(ctx, &models.Event{
		ID:           "evt_1",
		RepositoryID: repoID,
		EventType:    "FEATURE",
		Title:        "Feature A",
		SourceID:     "src_qa",
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
		SourceID:     "src_qa",
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
		SourceID:     "src_qa",
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
		SourceID:     "src_qa",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to write fact: %v", err)
	}

	// 4. Write relationships to store
	relationshipStore := memorystorage.NewSQLiteRelationshipStore(db)
	err = relationshipStore.UpsertRelationship(ctx, &memorymodels.Relationship{
		ID:           "rel_1",
		RepositoryID: repoID,
		FromID:       "int_1",
		ToID:         "dec_1",
		Type:         "INTENT_DRIVES_DECISION",
	})
	if err != nil {
		t.Fatalf("failed to write relationship: %v", err)
	}
	err = relationshipStore.UpsertRelationship(ctx, &memorymodels.Relationship{
		ID:           "rel_2",
		RepositoryID: repoID,
		FromID:       "fact_1",
		ToID:         "dec_1",
		Type:         "FACT_SUPPORTS_DECISION",
	})
	if err != nil {
		t.Fatalf("failed to write relationship: %v", err)
	}
	err = relationshipStore.UpsertRelationship(ctx, &memorymodels.Relationship{
		ID:           "rel_3",
		RepositoryID: repoID,
		FromID:       "dec_1",
		ToID:         "evt_1",
		Type:         "DECISION_RESULTS_IN_EVENT",
	})
	if err != nil {
		t.Fatalf("failed to write relationship: %v", err)
	}

	// 5. Instantiation of composed Readers, ContextReader and Generator
	eventReader := storage.NewSQLiteEventReader(db)
	decisionReader := storage.NewSQLiteDecisionReader(db)
	intentReader := storage.NewSQLiteIntentReader(db)
	factReader := storage.NewSQLiteFactReader(db)
	relationshipReader := storage.NewSQLiteRelationshipReader(db)

	contextReader := ctxengine.NewMemoryContextReader(eventReader, decisionReader, intentReader, factReader)
	generator := ctxengine.NewGenerator(contextReader)

	// Instantiation of Domain Services
	obs := onboarding.NewService(generator)
	gs := guidance.NewService(decisionReader, intentReader, factReader, eventReader, relationshipReader)
	is := impact.NewService(decisionReader, intentReader, factReader, eventReader, relationshipReader)

	qaService := NewService(obs, gs, is)

	// 6. Test Q&A routing end-to-end
	ans, err := qaService.Answer(ctx, repoID, Question{Text: "Why was decision dec_1 made?"})
	if err != nil {
		t.Fatalf("Q&A Answer failed: %v", err)
	}
	if ans.Question != "Why was decision dec_1 made?" {
		t.Errorf("unexpected question text: %q", ans.Question)
	}

	g, ok := ans.Result.(*guidance.Guidance)
	if !ok {
		t.Fatalf("expected *guidance.Guidance, got %T", ans.Result)
	}
	if len(g.Reasons) != 1 || g.Reasons[0] != "Intent A" {
		t.Errorf("unexpected reasons: %v", g.Reasons)
	}
}
