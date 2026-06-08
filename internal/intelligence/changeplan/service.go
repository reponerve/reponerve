package changeplan

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"reponerve/internal/graph/impact"
)

// Service generates change plans based on graph-aware impact analysis.
type Service struct {
	impactService *impact.Service
}

// NewService constructs a new change planning Service.
func NewService(impactService *impact.Service) *Service {
	return &Service{
		impactService: impactService,
	}
}

// GenerateDecisionPlan generates a change plan prior to modifying a decision.
func (s *Service) GenerateDecisionPlan(ctx context.Context, repositoryID string, decisionID string) (*ChangePlan, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	if decisionID == "" {
		return nil, fmt.Errorf("decision ID cannot be empty")
	}

	report, err := s.impactService.AnalyzeDecisionImpact(ctx, repositoryID, decisionID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze decision impact: %w", err)
	}

	return s.buildChangePlan(ctx, repositoryID, report)
}

// GenerateFactPlan generates a change plan prior to modifying a fact.
func (s *Service) GenerateFactPlan(ctx context.Context, repositoryID string, factID string) (*ChangePlan, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	if factID == "" {
		return nil, fmt.Errorf("fact ID cannot be empty")
	}

	report, err := s.impactService.AnalyzeFactImpact(ctx, repositoryID, factID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze fact impact: %w", err)
	}

	return s.buildChangePlan(ctx, repositoryID, report)
}

// GenerateEventPlan generates a change plan prior to modifying an event.
func (s *Service) GenerateEventPlan(ctx context.Context, repositoryID string, eventID string) (*ChangePlan, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	if eventID == "" {
		return nil, fmt.Errorf("event ID cannot be empty")
	}

	report, err := s.impactService.AnalyzeEventImpact(ctx, repositoryID, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze event impact: %w", err)
	}

	return s.buildChangePlan(ctx, repositoryID, report)
}

// GenerateContributorPlan generates a change plan prior to modifying contributor areas.
func (s *Service) GenerateContributorPlan(ctx context.Context, repositoryID string, contributorID string) (*ChangePlan, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	if contributorID == "" {
		return nil, fmt.Errorf("contributor ID cannot be empty")
	}

	report, err := s.impactService.AnalyzeContributorImpact(ctx, repositoryID, contributorID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze contributor impact: %w", err)
	}

	return s.buildChangePlan(ctx, repositoryID, report)
}

func (s *Service) buildChangePlan(ctx context.Context, repositoryID string, report *impact.ImpactReport) (*ChangePlan, error) {
	type key struct {
		entityType string
		entityID   string
	}
	itemsMap := make(map[key]*ChangePlanItem)

	for _, path := range report.ImpactPaths {
		if path.Path == nil {
			continue
		}
		// Start from idx = 1 to skip the starting node (the entity being changed)
		for idx := 1; idx < len(path.Path.Nodes); idx++ {
			node := path.Path.Nodes[idx]
			if node == nil || node.EntityID == "" {
				continue
			}

			var entityType string
			switch string(node.NodeType) {
			case "DECISION":
				entityType = "DECISION"
			case "FACT":
				entityType = "FACT"
			case "EVENT":
				entityType = "EVENT"
			case "CONTRIBUTOR":
				entityType = "CONTRIBUTOR"
			default:
				// ignore node types not valid in ChangePlanItem
				continue
			}

			// Map hop distance to Priority:
			// Priority 1: Direct impact (idx == 1)
			// Priority 2: One-hop impact (idx == 2)
			// Priority 3: Multi-hop impact (idx >= 3)
			priority := 3
			if idx == 1 {
				priority = 1
			} else if idx == 2 {
				priority = 2
			}

			k := key{entityType: entityType, entityID: node.EntityID}

			evidence := map[string]interface{}{
				"impact_path_length": idx,
				"impact_reason":      path.Reason,
			}
			evidenceBytes, _ := json.Marshal(evidence)

			var explanation string
			switch entityType {
			case "DECISION":
				if priority == 1 {
					explanation = "This decision should be reviewed because it directly depends on the changed decision."
				} else {
					explanation = "This decision should be reviewed because it participates in the impacted repository knowledge chain."
				}
			case "FACT":
				explanation = "This fact participates in the impacted repository knowledge chain."
			case "EVENT":
				explanation = "This event participates in the impacted repository knowledge chain."
			case "CONTRIBUTOR":
				explanation = "This contributor owns repository areas connected to the planned change."
			}

			item := &ChangePlanItem{
				EntityType:   entityType,
				EntityID:     node.EntityID,
				Priority:     priority,
				EvidenceJSON: string(evidenceBytes),
				Explanation:  explanation,
			}

			if err := ValidateItem(item); err != nil {
				return nil, fmt.Errorf("failed to validate change plan item: %w", err)
			}

			if existing, ok := itemsMap[k]; ok {
				if priority < existing.Priority {
					itemsMap[k] = item
				}
			} else {
				itemsMap[k] = item
			}
		}
	}

	items := make([]*ChangePlanItem, 0, len(itemsMap))
	for _, item := range itemsMap {
		items = append(items, item)
	}

	// Sort order: Priority ASC, EntityType ASC, EntityID ASC
	sort.Slice(items, func(i, j int) bool {
		if items[i].Priority != items[j].Priority {
			return items[i].Priority < items[j].Priority
		}
		if items[i].EntityType != items[j].EntityType {
			return items[i].EntityType < items[j].EntityType
		}
		return items[i].EntityID < items[j].EntityID
	})

	return &ChangePlan{Items: items}, nil
}
