package impact

import (
	"context"
	"database/sql"
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

// --- Mock Readers for Unit Tests ---

type mockDecisionReader struct {
	dec *memorymodels.Decision
	err error
}

func (m *mockDecisionReader) GetByID(ctx context.Context, id string) (*memorymodels.Decision, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.dec == nil {
		return nil, sql.ErrNoRows
	}
	return m.dec, nil
}
func (m *mockDecisionReader) ListByRepository(ctx context.Context, repositoryID string) ([]*memorymodels.Decision, error) {
	return nil, nil
}
func (m *mockDecisionReader) ListAll(ctx context.Context) ([]*memorymodels.Decision, error) {
	return nil, nil
}

type mockIntentReader struct {
	it  *memorymodels.Intent
	err error
}

func (m *mockIntentReader) GetByID(ctx context.Context, id string) (*memorymodels.Intent, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.it == nil {
		return nil, sql.ErrNoRows
	}
	return m.it, nil
}
func (m *mockIntentReader) ListByRepository(ctx context.Context, repositoryID string) ([]*memorymodels.Intent, error) {
	return nil, nil
}
func (m *mockIntentReader) ListAll(ctx context.Context) ([]*memorymodels.Intent, error) {
	return nil, nil
}

type mockFactReader struct {
	f   *memorymodels.Fact
	err error
}

func (m *mockFactReader) GetByID(ctx context.Context, id string) (*memorymodels.Fact, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.f == nil {
		return nil, sql.ErrNoRows
	}
	return m.f, nil
}
func (m *mockFactReader) ListByRepository(ctx context.Context, repositoryID string) ([]*memorymodels.Fact, error) {
	return nil, nil
}
func (m *mockFactReader) ListAll(ctx context.Context) ([]*memorymodels.Fact, error) {
	return nil, nil
}

type mockEventReader struct {
	evt *models.Event
	err error
}

func (m *mockEventReader) GetByID(ctx context.Context, id string) (*models.Event, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.evt == nil {
		return nil, sql.ErrNoRows
	}
	return m.evt, nil
}
func (m *mockEventReader) ListByRepository(ctx context.Context, repositoryID string) ([]*models.Event, error) {
	return nil, nil
}
func (m *mockEventReader) ListAll(ctx context.Context) ([]*models.Event, error) {
	return nil, nil
}

type mockRelationshipReader struct {
	rels []*memorymodels.Relationship
	err  error
}

func (m *mockRelationshipReader) GetByID(ctx context.Context, id string) (*memorymodels.Relationship, error) {
	return nil, nil
}
func (m *mockRelationshipReader) ListByRepository(ctx context.Context, repositoryID string) ([]*memorymodels.Relationship, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.rels, nil
}
func (m *mockRelationshipReader) ListAll(ctx context.Context) ([]*memorymodels.Relationship, error) {
	return nil, nil
}

// --- Unit Tests ---

func TestService_Unit(t *testing.T) {
	ctx := context.Background()

	t.Run("Decision Impact", func(t *testing.T) {
		dr := &mockDecisionReader{dec: &memorymodels.Decision{ID: "dec_1", RepositoryID: "repo_1"}}
		ir := &mockIntentReader{it: &memorymodels.Intent{ID: "int_1"}}
		fr := &mockFactReader{f: &memorymodels.Fact{ID: "fact_1"}}
		er := &mockEventReader{evt: &models.Event{ID: "evt_1"}}
		rr := &mockRelationshipReader{rels: []*memorymodels.Relationship{
			{ID: "r1", Type: "INTENT_DRIVES_DECISION", FromID: "int_1", ToID: "dec_1"},
			{ID: "r2", Type: "FACT_SUPPORTS_DECISION", FromID: "fact_1", ToID: "dec_1"},
			{ID: "r3", Type: "DECISION_RESULTS_IN_EVENT", FromID: "dec_1", ToID: "evt_1"},
		}}

		s := NewService(dr, ir, fr, er, rr)
		report, err := s.AnalyzeDecisionImpact(ctx, "dec_1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if report.EntityID != "dec_1" {
			t.Errorf("expected EntityID dec_1, got %q", report.EntityID)
		}
		if len(report.Intents) != 1 || report.Intents[0].ID != "int_1" {
			t.Errorf("expected 1 intent, got %v", report.Intents)
		}
		if len(report.Facts) != 1 || report.Facts[0].ID != "fact_1" {
			t.Errorf("expected 1 fact, got %v", report.Facts)
		}
		if len(report.Events) != 1 || report.Events[0].ID != "evt_1" {
			t.Errorf("expected 1 event, got %v", report.Events)
		}
	})

	t.Run("Event Impact", func(t *testing.T) {
		dr := &mockDecisionReader{dec: &memorymodels.Decision{ID: "dec_1"}}
		ir := &mockIntentReader{it: &memorymodels.Intent{ID: "int_1"}}
		fr := &mockFactReader{}
		er := &mockEventReader{evt: &models.Event{ID: "evt_1", RepositoryID: "repo_1"}}
		rr := &mockRelationshipReader{rels: []*memorymodels.Relationship{
			{ID: "r1", Type: "DECISION_RESULTS_IN_EVENT", FromID: "dec_1", ToID: "evt_1"},
			{ID: "r2", Type: "INTENT_DRIVES_DECISION", FromID: "int_1", ToID: "dec_1"},
		}}

		s := NewService(dr, ir, fr, er, rr)
		report, err := s.AnalyzeEventImpact(ctx, "evt_1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if report.EntityID != "evt_1" {
			t.Errorf("expected EntityID evt_1, got %q", report.EntityID)
		}
		if len(report.Decisions) != 1 || report.Decisions[0].ID != "dec_1" {
			t.Errorf("expected 1 causing decision, got %v", report.Decisions)
		}
		if len(report.Intents) != 1 || report.Intents[0].ID != "int_1" {
			t.Errorf("expected 1 related intent, got %v", report.Intents)
		}
	})

	t.Run("Intent Impact", func(t *testing.T) {
		dr := &mockDecisionReader{dec: &memorymodels.Decision{ID: "dec_1"}}
		ir := &mockIntentReader{it: &memorymodels.Intent{ID: "int_1", RepositoryID: "repo_1"}}
		fr := &mockFactReader{}
		er := &mockEventReader{evt: &models.Event{ID: "evt_1"}}
		rr := &mockRelationshipReader{rels: []*memorymodels.Relationship{
			{ID: "r1", Type: "INTENT_DRIVES_DECISION", FromID: "int_1", ToID: "dec_1"},
			{ID: "r2", Type: "DECISION_RESULTS_IN_EVENT", FromID: "dec_1", ToID: "evt_1"},
		}}

		s := NewService(dr, ir, fr, er, rr)
		report, err := s.AnalyzeIntentImpact(ctx, "int_1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if report.EntityID != "int_1" {
			t.Errorf("expected EntityID int_1, got %q", report.EntityID)
		}
		if len(report.Decisions) != 1 || report.Decisions[0].ID != "dec_1" {
			t.Errorf("expected 1 driven decision, got %v", report.Decisions)
		}
		if len(report.Events) != 1 || report.Events[0].ID != "evt_1" {
			t.Errorf("expected 1 resulting event, got %v", report.Events)
		}
	})

	t.Run("Fact Impact", func(t *testing.T) {
		dr := &mockDecisionReader{dec: &memorymodels.Decision{ID: "dec_1"}}
		ir := &mockIntentReader{}
		fr := &mockFactReader{f: &memorymodels.Fact{ID: "fact_1", RepositoryID: "repo_1"}}
		er := &mockEventReader{evt: &models.Event{ID: "evt_1"}}
		rr := &mockRelationshipReader{rels: []*memorymodels.Relationship{
			{ID: "r1", Type: "FACT_SUPPORTS_DECISION", FromID: "fact_1", ToID: "dec_1"},
			{ID: "r2", Type: "DECISION_RESULTS_IN_EVENT", FromID: "dec_1", ToID: "evt_1"},
		}}

		s := NewService(dr, ir, fr, er, rr)
		report, err := s.AnalyzeFactImpact(ctx, "fact_1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if report.EntityID != "fact_1" {
			t.Errorf("expected EntityID fact_1, got %q", report.EntityID)
		}
		if len(report.Decisions) != 1 || report.Decisions[0].ID != "dec_1" {
			t.Errorf("expected 1 supported decision, got %v", report.Decisions)
		}
		if len(report.Events) != 1 || report.Events[0].ID != "evt_1" {
			t.Errorf("expected 1 resulting event, got %v", report.Events)
		}
	})

	t.Run("Entity not found / missing entities", func(t *testing.T) {
		dr := &mockDecisionReader{dec: nil}
		ir := &mockIntentReader{it: nil}
		fr := &mockFactReader{f: nil}
		er := &mockEventReader{evt: nil}
		rr := &mockRelationshipReader{}

		s := NewService(dr, ir, fr, er, rr)

		_, err := s.AnalyzeDecisionImpact(ctx, "dec_missing")
		if err == nil {
			t.Error("expected error for missing decision, got nil")
		}

		_, err = s.AnalyzeEventImpact(ctx, "evt_missing")
		if err == nil {
			t.Error("expected error for missing event, got nil")
		}

		_, err = s.AnalyzeIntentImpact(ctx, "intent_missing")
		if err == nil {
			t.Error("expected error for missing intent, got nil")
		}

		_, err = s.AnalyzeFactImpact(ctx, "fact_missing")
		if err == nil {
			t.Error("expected error for missing fact, got nil")
		}
	})

	t.Run("Reader failures propagation", func(t *testing.T) {
		expectedErr := errors.New("failed reader")
		dr := &mockDecisionReader{err: expectedErr}
		ir := &mockIntentReader{err: expectedErr}
		fr := &mockFactReader{err: expectedErr}
		er := &mockEventReader{err: expectedErr}
		rr := &mockRelationshipReader{}

		s := NewService(dr, ir, fr, er, rr)

		_, err := s.AnalyzeDecisionImpact(ctx, "dec_1")
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}

		_, err = s.AnalyzeEventImpact(ctx, "evt_1")
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}

		_, err = s.AnalyzeIntentImpact(ctx, "intent_1")
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}

		_, err = s.AnalyzeFactImpact(ctx, "fact_1")
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}

// --- Integration Tests ---

func TestService_Integration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-impact-integration-test-*")
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
	repoID := "repo_impact"

	// 1. Insert Repository
	_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repoID, "Repo Impact", tempDir, "main")
	if err != nil {
		t.Fatalf("failed to insert repository: %v", err)
	}

	// 2. Insert Source
	_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_impact", repoID, "adr", "docs/adr/0001.md", "Author", "2024-01-01T00:00:00Z")
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
		SourceID:     "src_impact",
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
		SourceID:     "src_impact",
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
		SourceID:     "src_impact",
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
		SourceID:     "src_impact",
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

	// 5. Instantiation of Readers and Guidance Service
	eventReader := storage.NewSQLiteEventReader(db)
	decisionReader := storage.NewSQLiteDecisionReader(db)
	intentReader := storage.NewSQLiteIntentReader(db)
	factReader := storage.NewSQLiteFactReader(db)
	relationshipReader := storage.NewSQLiteRelationshipReader(db)

	service := NewService(decisionReader, intentReader, factReader, eventReader, relationshipReader)

	// 6. Verify Intent Impact
	report, err := service.AnalyzeIntentImpact(ctx, "int_1")
	if err != nil {
		t.Fatalf("failed to resolve intent impact: %v", err)
	}
	if report.EntityID != "int_1" {
		t.Errorf("expected EntityID int_1, got %q", report.EntityID)
	}
	if len(report.Decisions) != 1 || report.Decisions[0].ID != "dec_1" {
		t.Errorf("unexpected decisions list: %v", report.Decisions)
	}
	if len(report.Events) != 1 || report.Events[0].ID != "evt_1" {
		t.Errorf("unexpected events list: %v", report.Events)
	}
}
