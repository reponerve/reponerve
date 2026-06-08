package guidance

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	"github.com/reponerve/reponerve/internal/query/storage"
	models "github.com/reponerve/reponerve/pkg/models"
)

// Service generates deterministic architectural guidance from repository memory.
type Service struct {
	decisionReader     storage.DecisionReader
	intentReader       storage.IntentReader
	factReader         storage.FactReader
	eventReader        storage.EventReader
	relationshipReader storage.RelationshipReader
}

// NewService constructs a new guidance Service.
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

// GetDecisionGuidance resolves driving intents, supporting facts, and resulting events for a decision.
func (s *Service) GetDecisionGuidance(ctx context.Context, decisionID string) (*Guidance, error) {
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

	var relatedIntents []*memorymodels.Intent
	var supportingFacts []*memorymodels.Fact
	var relatedEvents []*models.Event

	for _, r := range allRels {
		if r.ToID == decisionID {
			if r.Type == "INTENT_DRIVES_DECISION" {
				it, err := s.intentReader.GetByID(ctx, r.FromID)
				if err == nil {
					relatedIntents = append(relatedIntents, it)
				}
			} else if r.Type == "FACT_SUPPORTS_DECISION" {
				f, err := s.factReader.GetByID(ctx, r.FromID)
				if err == nil {
					supportingFacts = append(supportingFacts, f)
				}
			}
		} else if r.FromID == decisionID && r.Type == "DECISION_RESULTS_IN_EVENT" {
			e, err := s.eventReader.GetByID(ctx, r.ToID)
			if err == nil {
				relatedEvents = append(relatedEvents, e)
			}
		}
	}

	// Deterministic sorting by ID ascending
	sort.Slice(relatedIntents, func(i, j int) bool {
		return relatedIntents[i].ID < relatedIntents[j].ID
	})
	sort.Slice(supportingFacts, func(i, j int) bool {
		return supportingFacts[i].ID < supportingFacts[j].ID
	})
	sort.Slice(relatedEvents, func(i, j int) bool {
		return relatedEvents[i].ID < relatedEvents[j].ID
	})

	var reasons []string
	for _, it := range relatedIntents {
		reasons = append(reasons, it.Description)
	}

	return &Guidance{
		EntityID:        decisionID,
		Reasons:         reasons,
		SupportingFacts: supportingFacts,
		RelatedIntents:  relatedIntents,
		RelatedEvents:   relatedEvents,
	}, nil
}

// GetEventGuidance resolves causing decisions and indirect intents for an event.
func (s *Service) GetEventGuidance(ctx context.Context, eventID string) (*Guidance, error) {
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

	var causingDecisions []*memorymodels.Decision
	for _, r := range allRels {
		if r.ToID == eventID && r.Type == "DECISION_RESULTS_IN_EVENT" {
			dec, err := s.decisionReader.GetByID(ctx, r.FromID)
			if err == nil {
				causingDecisions = append(causingDecisions, dec)
			}
		}
	}

	// Sort causing decisions by ID ascending
	sort.Slice(causingDecisions, func(i, j int) bool {
		return causingDecisions[i].ID < causingDecisions[j].ID
	})

	var reasons []string
	for _, dec := range causingDecisions {
		reasons = append(reasons, "Caused by decision: "+dec.Title)
	}

	intentMap := make(map[string]*memorymodels.Intent)
	for _, dec := range causingDecisions {
		for _, r := range allRels {
			if r.ToID == dec.ID && r.Type == "INTENT_DRIVES_DECISION" {
				intent, err := s.intentReader.GetByID(ctx, r.FromID)
				if err == nil {
					intentMap[intent.ID] = intent
				}
			}
		}
	}

	var relatedIntents []*memorymodels.Intent
	for _, it := range intentMap {
		relatedIntents = append(relatedIntents, it)
	}

	// Sort related intents by ID ascending
	sort.Slice(relatedIntents, func(i, j int) bool {
		return relatedIntents[i].ID < relatedIntents[j].ID
	})

	for _, it := range relatedIntents {
		reasons = append(reasons, "Driven by intent: "+it.Description)
	}

	return &Guidance{
		EntityID:        eventID,
		Reasons:         reasons,
		SupportingFacts: []*memorymodels.Fact{},
		RelatedIntents:  relatedIntents,
		RelatedEvents:   []*models.Event{},
	}, nil
}
