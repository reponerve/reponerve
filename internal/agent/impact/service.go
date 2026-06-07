package impact

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	memorymodels "reponerve/internal/memory/models"
	"reponerve/internal/query/storage"
	models "reponerve/pkg/models"
)

// Service provides deterministic impact analysis using repository memory.
type Service struct {
	decisionReader     storage.DecisionReader
	intentReader       storage.IntentReader
	factReader         storage.FactReader
	eventReader        storage.EventReader
	relationshipReader storage.RelationshipReader
}

// NewService constructs a new impact Service.
func NewService(
	dr storage.DecisionReader,
	ir storage.IntentReader,
	fr storage.FactReader,
	er storage.EventReader,
	rr storage.RelationshipReader,
) *Service {
	return &Service{
		decisionReader:     dr,
		intentReader:       ir,
		factReader:         fr,
		eventReader:        er,
		relationshipReader: rr,
	}
}

// AnalyzeDecisionImpact resolves related intents, supporting facts, and resulting events for a decision.
func (s *Service) AnalyzeDecisionImpact(ctx context.Context, decisionID string) (*ImpactReport, error) {
	dec, err := s.decisionReader.GetByID(ctx, decisionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("decision with ID %q not found", decisionID)
		}
		return nil, fmt.Errorf("failed to fetch decision: %w", err)
	}

	allRels, err := s.relationshipReader.ListByRepository(ctx, dec.RepositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch relationships: %w", err)
	}

	var intents []*memorymodels.Intent
	var facts []*memorymodels.Fact
	var events []*models.Event

	for _, r := range allRels {
		if r.ToID == decisionID {
			if r.Type == "INTENT_DRIVES_DECISION" {
				it, err := s.intentReader.GetByID(ctx, r.FromID)
				if err == nil {
					intents = append(intents, it)
				}
			} else if r.Type == "FACT_SUPPORTS_DECISION" {
				f, err := s.factReader.GetByID(ctx, r.FromID)
				if err == nil {
					facts = append(facts, f)
				}
			}
		} else if r.FromID == decisionID && r.Type == "DECISION_RESULTS_IN_EVENT" {
			e, err := s.eventReader.GetByID(ctx, r.ToID)
			if err == nil {
				events = append(events, e)
			}
		}
	}

	// Deterministic sorting by ID ascending
	sort.Slice(intents, func(i, j int) bool { return intents[i].ID < intents[j].ID })
	sort.Slice(facts, func(i, j int) bool { return facts[i].ID < facts[j].ID })
	sort.Slice(events, func(i, j int) bool { return events[i].ID < events[j].ID })

	return &ImpactReport{
		EntityID:  decisionID,
		Decisions: []*memorymodels.Decision{},
		Intents:   intents,
		Facts:     facts,
		Events:    events,
	}, nil
}

// AnalyzeEventImpact resolves causing decisions and related intents for an event.
func (s *Service) AnalyzeEventImpact(ctx context.Context, eventID string) (*ImpactReport, error) {
	evt, err := s.eventReader.GetByID(ctx, eventID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event with ID %q not found", eventID)
		}
		return nil, fmt.Errorf("failed to fetch event: %w", err)
	}

	allRels, err := s.relationshipReader.ListByRepository(ctx, evt.RepositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch relationships: %w", err)
	}

	var decisions []*memorymodels.Decision
	intentMap := make(map[string]*memorymodels.Intent)

	for _, r := range allRels {
		if r.ToID == eventID && r.Type == "DECISION_RESULTS_IN_EVENT" {
			dec, err := s.decisionReader.GetByID(ctx, r.FromID)
			if err == nil {
				decisions = append(decisions, dec)
				// Find intents driving this decision
				for _, r2 := range allRels {
					if r2.ToID == dec.ID && r2.Type == "INTENT_DRIVES_DECISION" {
						it, err := s.intentReader.GetByID(ctx, r2.FromID)
						if err == nil {
							intentMap[it.ID] = it
						}
					}
				}
			}
		}
	}

	var intents []*memorymodels.Intent
	for _, it := range intentMap {
		intents = append(intents, it)
	}

	// Deterministic sorting by ID ascending
	sort.Slice(decisions, func(i, j int) bool { return decisions[i].ID < decisions[j].ID })
	sort.Slice(intents, func(i, j int) bool { return intents[i].ID < intents[j].ID })

	return &ImpactReport{
		EntityID:  eventID,
		Decisions: decisions,
		Intents:   intents,
		Facts:     []*memorymodels.Fact{},
		Events:    []*models.Event{},
	}, nil
}

// AnalyzeIntentImpact resolves driven decisions and resulting events for an intent.
func (s *Service) AnalyzeIntentImpact(ctx context.Context, intentID string) (*ImpactReport, error) {
	intent, err := s.intentReader.GetByID(ctx, intentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("intent with ID %q not found", intentID)
		}
		return nil, fmt.Errorf("failed to fetch intent: %w", err)
	}

	allRels, err := s.relationshipReader.ListByRepository(ctx, intent.RepositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch relationships: %w", err)
	}

	var decisions []*memorymodels.Decision
	eventMap := make(map[string]*models.Event)

	for _, r := range allRels {
		if r.FromID == intentID && r.Type == "INTENT_DRIVES_DECISION" {
			dec, err := s.decisionReader.GetByID(ctx, r.ToID)
			if err == nil {
				decisions = append(decisions, dec)
				// Find events resulting from this decision
				for _, r2 := range allRels {
					if r2.FromID == dec.ID && r2.Type == "DECISION_RESULTS_IN_EVENT" {
						e, err := s.eventReader.GetByID(ctx, r2.ToID)
						if err == nil {
							eventMap[e.ID] = e
						}
					}
				}
			}
		}
	}

	var events []*models.Event
	for _, e := range eventMap {
		events = append(events, e)
	}

	// Deterministic sorting by ID ascending
	sort.Slice(decisions, func(i, j int) bool { return decisions[i].ID < decisions[j].ID })
	sort.Slice(events, func(i, j int) bool { return events[i].ID < events[j].ID })

	return &ImpactReport{
		EntityID:  intentID,
		Decisions: decisions,
		Intents:   []*memorymodels.Intent{},
		Facts:     []*memorymodels.Fact{},
		Events:    events,
	}, nil
}

// AnalyzeFactImpact resolves supported decisions and resulting events for a fact.
func (s *Service) AnalyzeFactImpact(ctx context.Context, factID string) (*ImpactReport, error) {
	fact, err := s.factReader.GetByID(ctx, factID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("fact with ID %q not found", factID)
		}
		return nil, fmt.Errorf("failed to fetch fact: %w", err)
	}

	allRels, err := s.relationshipReader.ListByRepository(ctx, fact.RepositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch relationships: %w", err)
	}

	var decisions []*memorymodels.Decision
	eventMap := make(map[string]*models.Event)

	for _, r := range allRels {
		if r.FromID == factID && r.Type == "FACT_SUPPORTS_DECISION" {
			dec, err := s.decisionReader.GetByID(ctx, r.ToID)
			if err == nil {
				decisions = append(decisions, dec)
				// Find events resulting from this decision
				for _, r2 := range allRels {
					if r2.FromID == dec.ID && r2.Type == "DECISION_RESULTS_IN_EVENT" {
						e, err := s.eventReader.GetByID(ctx, r2.ToID)
						if err == nil {
							eventMap[e.ID] = e
						}
					}
				}
			}
		}
	}

	var events []*models.Event
	for _, e := range eventMap {
		events = append(events, e)
	}

	// Deterministic sorting by ID ascending
	sort.Slice(decisions, func(i, j int) bool { return decisions[i].ID < decisions[j].ID })
	sort.Slice(events, func(i, j int) bool { return events[i].ID < events[j].ID })

	return &ImpactReport{
		EntityID:  factID,
		Decisions: decisions,
		Intents:   []*memorymodels.Intent{},
		Facts:     []*memorymodels.Fact{},
		Events:    events,
	}, nil
}
