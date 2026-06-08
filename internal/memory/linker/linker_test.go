package linker

import (
	"context"
	"testing"
	"time"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	models "github.com/reponerve/reponerve/pkg/models"
)

func TestLink_IntentDrivesDecision(t *testing.T) {
	linker := NewLinker()
	repoID := "test-repo"

	it := &memorymodels.Intent{
		ID:           "intent_1",
		RepositoryID: repoID,
		Description:  "Reduce latency",
		SourceID:     "adr_1",
		CreatedAt:    time.Now(),
	}
	dec := &memorymodels.Decision{
		ID:           "decision_1",
		RepositoryID: repoID,
		Title:        "Use Redis",
		Status:       "Accepted",
		SourceID:     "adr_1",
		CreatedAt:    time.Now(),
	}

	input := LinkInput{
		Intents:   []*memorymodels.Intent{it},
		Decisions: []*memorymodels.Decision{dec},
	}

	rels, err := linker.Link(context.Background(), input)
	if err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	if len(rels) != 1 {
		t.Fatalf("expected 1 relationship, got %d", len(rels))
	}

	r := rels[0]
	if r.FromID != it.ID {
		t.Errorf("expected FromID %q, got %q", it.ID, r.FromID)
	}
	if r.ToID != dec.ID {
		t.Errorf("expected ToID %q, got %q", dec.ID, r.ToID)
	}
	if r.Type != "INTENT_DRIVES_DECISION" {
		t.Errorf("expected Type 'INTENT_DRIVES_DECISION', got %q", r.Type)
	}
	if r.RepositoryID != repoID {
		t.Errorf("expected RepositoryID %q, got %q", repoID, r.RepositoryID)
	}
	if r.ID == "" {
		t.Error("expected non-empty relationship ID")
	}
}

func TestLink_DecisionResultsInEvent(t *testing.T) {
	linker := NewLinker()
	repoID := "test-repo"

	dec := &memorymodels.Decision{
		ID:           "decision_1",
		RepositoryID: repoID,
		Title:        "Use Redis Cache",
		Status:       "Accepted",
		SourceID:     "adr_1",
		CreatedAt:    time.Now(),
	}
	evt := &models.Event{
		ID:           "event_1",
		RepositoryID: repoID,
		EventType:    "FEATURE_INTRODUCED",
		Title:        "Introduce Redis Cache",
		SourceID:     "commit_1",
		Timestamp:    time.Now(),
	}

	input := LinkInput{
		Decisions: []*memorymodels.Decision{dec},
		Events:    []*models.Event{evt},
	}

	rels, err := linker.Link(context.Background(), input)
	if err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	if len(rels) != 1 {
		t.Fatalf("expected 1 relationship, got %d", len(rels))
	}

	r := rels[0]
	if r.FromID != dec.ID {
		t.Errorf("expected FromID %q, got %q", dec.ID, r.FromID)
	}
	if r.ToID != evt.ID {
		t.Errorf("expected ToID %q, got %q", evt.ID, r.ToID)
	}
	if r.Type != "DECISION_RESULTS_IN_EVENT" {
		t.Errorf("expected Type 'DECISION_RESULTS_IN_EVENT', got %q", r.Type)
	}
}

func TestLink_FactSupportsDecision(t *testing.T) {
	linker := NewLinker()
	repoID := "test-repo"

	f := &memorymodels.Fact{
		ID:           "fact_1",
		RepositoryID: repoID,
		Subject:      "Auth Service",
		Predicate:    "USES",
		Object:       "Redis",
		SourceID:     "adr_2",
		CreatedAt:    time.Now(),
	}
	dec := &memorymodels.Decision{
		ID:           "decision_1",
		RepositoryID: repoID,
		Title:        "Use Redis Cache",
		Status:       "Accepted",
		SourceID:     "adr_1",
		CreatedAt:    time.Now(),
	}

	input := LinkInput{
		Facts:     []*memorymodels.Fact{f},
		Decisions: []*memorymodels.Decision{dec},
	}

	rels, err := linker.Link(context.Background(), input)
	if err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	if len(rels) != 1 {
		t.Fatalf("expected 1 relationship, got %d", len(rels))
	}

	r := rels[0]
	if r.FromID != f.ID {
		t.Errorf("expected FromID %q, got %q", f.ID, r.FromID)
	}
	if r.ToID != dec.ID {
		t.Errorf("expected ToID %q, got %q", dec.ID, r.ToID)
	}
	if r.Type != "FACT_SUPPORTS_DECISION" {
		t.Errorf("expected Type 'FACT_SUPPORTS_DECISION', got %q", r.Type)
	}
}

func TestLink_NegativeCases(t *testing.T) {
	linker := NewLinker()
	repoID := "test-repo"

	// Unrelated memories
	it := &memorymodels.Intent{
		ID:           "intent_1",
		RepositoryID: repoID,
		Description:  "Optimize Storage",
		SourceID:     "adr_1",
		CreatedAt:    time.Now(),
	}
	dec := &memorymodels.Decision{
		ID:           "decision_1",
		RepositoryID: repoID,
		Title:        "Use Go",
		Status:       "Accepted",
		SourceID:     "adr_2",
		CreatedAt:    time.Now(),
	}
	evt := &models.Event{
		ID:           "event_1",
		RepositoryID: repoID,
		Title:        "Initial Commit",
		SourceID:     "commit_1",
		Timestamp:    time.Now(),
	}
	f := &memorymodels.Fact{
		ID:           "fact_1",
		RepositoryID: repoID,
		Subject:      "Auth Service",
		Predicate:    "USES",
		Object:       "Redis",
		SourceID:     "adr_3",
		CreatedAt:    time.Now(),
	}

	input := LinkInput{
		Intents:   []*memorymodels.Intent{it},
		Decisions: []*memorymodels.Decision{dec},
		Events:    []*models.Event{evt},
		Facts:     []*memorymodels.Fact{f},
	}

	rels, err := linker.Link(context.Background(), input)
	if err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	if len(rels) != 0 {
		t.Errorf("expected 0 relationships, got %d: %+v", len(rels), rels[0])
	}
}
