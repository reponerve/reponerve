package guidance

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
	intents map[string]*memorymodels.Intent
	err     error
}

func (m *mockIntentReader) GetByID(ctx context.Context, id string) (*memorymodels.Intent, error) {
	if m.err != nil {
		return nil, m.err
	}
	it, ok := m.intents[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return it, nil
}
func (m *mockIntentReader) ListByRepository(ctx context.Context, repositoryID string) ([]*memorymodels.Intent, error) {
	return nil, nil
}
func (m *mockIntentReader) ListAll(ctx context.Context) ([]*memorymodels.Intent, error) {
	return nil, nil
}

type mockFactReader struct {
	facts map[string]*memorymodels.Fact
	err   error
}

func (m *mockFactReader) GetByID(ctx context.Context, id string) (*memorymodels.Fact, error) {
	if m.err != nil {
		return nil, m.err
	}
	f, ok := m.facts[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return f, nil
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

	t.Run("Decision guidance success and sorting", func(t *testing.T) {
		dr := &mockDecisionReader{dec: &memorymodels.Decision{ID: "dec_1", RepositoryID: "repo_1", Title: "Decision 1"}}
		ir := &mockIntentReader{intents: map[string]*memorymodels.Intent{
			"int_2": {ID: "int_2", Description: "Intent B"},
			"int_1": {ID: "int_1", Description: "Intent A"},
		}}
		fr := &mockFactReader{facts: map[string]*memorymodels.Fact{
			"fact_2": {ID: "fact_2", Subject: "Fact B"},
			"fact_1": {ID: "fact_1", Subject: "Fact A"},
		}}
		er := &mockEventReader{evt: &models.Event{ID: "evt_1", Title: "Event A"}}
		rr := &mockRelationshipReader{rels: []*memorymodels.Relationship{
			{ID: "r1", Type: "INTENT_DRIVES_DECISION", FromID: "int_2", ToID: "dec_1"},
			{ID: "r2", Type: "INTENT_DRIVES_DECISION", FromID: "int_1", ToID: "dec_1"},
			{ID: "r3", Type: "FACT_SUPPORTS_DECISION", FromID: "fact_2", ToID: "dec_1"},
			{ID: "r4", Type: "FACT_SUPPORTS_DECISION", FromID: "fact_1", ToID: "dec_1"},
			{ID: "r5", Type: "DECISION_RESULTS_IN_EVENT", FromID: "dec_1", ToID: "evt_1"},
		}}

		s := NewService(dr, ir, fr, er, rr)
		guidance, err := s.GetDecisionGuidance(ctx, "dec_1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if guidance.EntityID != "dec_1" {
			t.Errorf("expected EntityID dec_1, got %q", guidance.EntityID)
		}

		// Reasons should match sorted intents (int_1 then int_2 descriptions)
		if len(guidance.Reasons) != 2 || guidance.Reasons[0] != "Intent A" || guidance.Reasons[1] != "Intent B" {
			t.Errorf("reasons sorted incorrectly: %v", guidance.Reasons)
		}

		// SupportingFacts sorted by ID (fact_1 then fact_2)
		if len(guidance.SupportingFacts) != 2 || guidance.SupportingFacts[0].ID != "fact_1" || guidance.SupportingFacts[1].ID != "fact_2" {
			t.Errorf("supporting facts sorted incorrectly: %v", guidance.SupportingFacts)
		}

		// RelatedIntents sorted by ID (int_1 then int_2)
		if len(guidance.RelatedIntents) != 2 || guidance.RelatedIntents[0].ID != "int_1" || guidance.RelatedIntents[1].ID != "int_2" {
			t.Errorf("related intents sorted incorrectly: %v", guidance.RelatedIntents)
		}

		// RelatedEvents should contain evt_1
		if len(guidance.RelatedEvents) != 1 || guidance.RelatedEvents[0].ID != "evt_1" {
			t.Errorf("related events incorrect: %v", guidance.RelatedEvents)
		}
	})

	t.Run("Event guidance success and sorting", func(t *testing.T) {
		dr := &mockDecisionReader{dec: &memorymodels.Decision{ID: "dec_1", Title: "Decision A"}}
		ir := &mockIntentReader{intents: map[string]*memorymodels.Intent{
			"int_1": {ID: "int_1", Description: "Intent A"},
		}}
		fr := &mockFactReader{}
		er := &mockEventReader{evt: &models.Event{ID: "evt_1", RepositoryID: "repo_1", Title: "Event 1"}}
		rr := &mockRelationshipReader{rels: []*memorymodels.Relationship{
			{ID: "r1", Type: "DECISION_RESULTS_IN_EVENT", FromID: "dec_1", ToID: "evt_1"},
			{ID: "r2", Type: "INTENT_DRIVES_DECISION", FromID: "int_1", ToID: "dec_1"},
		}}

		s := NewService(dr, ir, fr, er, rr)
		guidance, err := s.GetEventGuidance(ctx, "evt_1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if guidance.EntityID != "evt_1" {
			t.Errorf("expected EntityID evt_1, got %q", guidance.EntityID)
		}

		expectedReasons := []string{
			"Caused by decision: Decision A",
			"Driven by intent: Intent A",
		}
		if len(guidance.Reasons) != 2 || guidance.Reasons[0] != expectedReasons[0] || guidance.Reasons[1] != expectedReasons[1] {
			t.Errorf("reasons incorrect: expected %v, got %v", expectedReasons, guidance.Reasons)
		}

		if len(guidance.RelatedIntents) != 1 || guidance.RelatedIntents[0].ID != "int_1" {
			t.Errorf("related intents incorrect: %v", guidance.RelatedIntents)
		}

		if len(guidance.SupportingFacts) != 0 || len(guidance.RelatedEvents) != 0 {
			t.Errorf("expected empty SupportingFacts and RelatedEvents, got %d and %d", len(guidance.SupportingFacts), len(guidance.RelatedEvents))
		}
	})

	t.Run("Missing relationships", func(t *testing.T) {
		dr := &mockDecisionReader{dec: &memorymodels.Decision{ID: "dec_1", RepositoryID: "repo_1", Title: "Decision A"}}
		ir := &mockIntentReader{}
		fr := &mockFactReader{}
		er := &mockEventReader{}
		rr := &mockRelationshipReader{rels: nil}

		s := NewService(dr, ir, fr, er, rr)
		guidance, err := s.GetDecisionGuidance(ctx, "dec_1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(guidance.Reasons) != 0 || len(guidance.SupportingFacts) != 0 || len(guidance.RelatedIntents) != 0 || len(guidance.RelatedEvents) != 0 {
			t.Errorf("expected all slices to be empty, got: %+v", guidance)
		}
	})

	t.Run("Entity not found (empty repository)", func(t *testing.T) {
		dr := &mockDecisionReader{dec: nil} // triggers ErrNoRows
		ir := &mockIntentReader{}
		fr := &mockFactReader{}
		er := &mockEventReader{evt: nil} // triggers ErrNoRows
		rr := &mockRelationshipReader{}

		s := NewService(dr, ir, fr, er, rr)

		_, err := s.GetDecisionGuidance(ctx, "dec_missing")
		if err == nil {
			t.Error("expected error for missing decision, got nil")
		}

		_, err = s.GetEventGuidance(ctx, "evt_missing")
		if err == nil {
			t.Error("expected error for missing event, got nil")
		}
	})

	t.Run("Reader failures propagation", func(t *testing.T) {
		expectedErr := errors.New("database failure")
		dr := &mockDecisionReader{err: expectedErr}
		ir := &mockIntentReader{}
		fr := &mockFactReader{}
		er := &mockEventReader{err: expectedErr}
		rr := &mockRelationshipReader{}

		s := NewService(dr, ir, fr, er, rr)

		_, err := s.GetDecisionGuidance(ctx, "dec_1")
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}

		_, err = s.GetEventGuidance(ctx, "evt_1")
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}

// --- Integration Tests ---

func TestService_Integration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-guidance-integration-test-*")
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
	repoID := "repo_guidance"

	// 1. Insert Repository
	_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repoID, "Repo Guidance", tempDir, "main")
	if err != nil {
		t.Fatalf("failed to insert repository: %v", err)
	}

	// 2. Insert Source
	_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_guidance", repoID, "adr", "docs/adr/0001.md", "Author", "2024-01-01T00:00:00Z")
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
		SourceID:     "src_guidance",
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
		SourceID:     "src_guidance",
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
		SourceID:     "src_guidance",
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
		SourceID:     "src_guidance",
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

	// 6. Verify Decision Guidance
	decGuidance, err := service.GetDecisionGuidance(ctx, "dec_1")
	if err != nil {
		t.Fatalf("failed to get decision guidance: %v", err)
	}
	if len(decGuidance.Reasons) != 1 || decGuidance.Reasons[0] != "Intent A" {
		t.Errorf("unexpected reasons: %v", decGuidance.Reasons)
	}
	if len(decGuidance.SupportingFacts) != 1 || decGuidance.SupportingFacts[0].ID != "fact_1" {
		t.Errorf("unexpected supporting facts: %v", decGuidance.SupportingFacts)
	}
	if len(decGuidance.RelatedEvents) != 1 || decGuidance.RelatedEvents[0].ID != "evt_1" {
		t.Errorf("unexpected related events: %v", decGuidance.RelatedEvents)
	}

	// 7. Verify Event Guidance
	evtGuidance, err := service.GetEventGuidance(ctx, "evt_1")
	if err != nil {
		t.Fatalf("failed to get event guidance: %v", err)
	}
	expectedReasons := []string{
		"Caused by decision: Decision A",
		"Driven by intent: Intent A",
	}
	if len(evtGuidance.Reasons) != 2 || evtGuidance.Reasons[0] != expectedReasons[0] || evtGuidance.Reasons[1] != expectedReasons[1] {
		t.Errorf("unexpected event reasons: %v", evtGuidance.Reasons)
	}
}
