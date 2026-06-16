package development

import (
	"context"
	"fmt"
	"sort"
	"strings"

	agentsearch "github.com/reponerve/reponerve/internal/agent/search"
)

const maxOnboardingDecisions = 10

// DevelopmentOnboardingGuide is a first-day context package for humans and agents.
type DevelopmentOnboardingGuide struct {
	RepositoryID    string             `json:"repository_id"`
	SuggestedSteps  []string           `json:"suggested_steps"`
	KeyDecisions    []EntityRef        `json:"key_decisions"`
	Orientation     *DevelopmentAnswer `json:"orientation,omitempty"`
	AssignmentPlan  *DevelopmentPlan   `json:"assignment_plan,omitempty"`
	EntityBriefings []EntityBriefing   `json:"entity_briefings,omitempty"`
	SourceServices  []string           `json:"source_services"`
}

// Onboard assembles day-one repository context and an optional assignment plan.
func (s *Service) Onboard(ctx context.Context, req DevelopmentRequest) (*DevelopmentOnboardingGuide, error) {
	if strings.TrimSpace(req.RepositoryID) == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}

	out := &DevelopmentOnboardingGuide{
		RepositoryID: req.RepositoryID,
		SuggestedSteps: []string{
			"1. Read orientation summary and key decisions.",
			"2. Use explain on domains you will touch.",
			"3. For assignments, follow assignment_plan suggested_steps.",
			"4. Run analyze_topic_impact before broad refactors.",
			"5. Run review before merge.",
		},
		SourceServices: []string{sourceRepositorySearch, sourceCodeIntelligence},
	}

	if s.decisionReader != nil {
		decisions, err := s.decisionReader.ListByRepository(ctx, req.RepositoryID)
		if err != nil {
			return nil, fmt.Errorf("list decisions: %w", err)
		}
		sort.Slice(decisions, func(i, j int) bool {
			return decisions[i].Title < decisions[j].Title
		})
		for i, d := range decisions {
			if i >= maxOnboardingDecisions {
				break
			}
			out.KeyDecisions = append(out.KeyDecisions, EntityRef{
				EntityType: agentsearch.EntityTypeDecision,
				EntityID:   d.ID,
				Label:      d.Title,
			})
		}
	}

	orientation, err := s.Ask(ctx, DevelopmentRequest{
		RepositoryID: req.RepositoryID,
		Topic:        "What does this repository do?",
	})
	if err == nil && orientation != nil {
		out.Orientation = orientation
		for _, svc := range orientation.SourceServices {
			out.SourceServices = appendUniqueSource(out.SourceServices, svc)
		}
	}

	assignment := strings.TrimSpace(req.Topic)
	if assignment != "" {
		plan, err := s.Plan(ctx, DevelopmentRequest{
			RepositoryID: req.RepositoryID,
			Topic:        assignment,
		})
		if err != nil {
			return nil, err
		}
		out.AssignmentPlan = plan
		out.EntityBriefings = plan.EntityBriefings
		for _, svc := range plan.SourceServices {
			out.SourceServices = appendUniqueSource(out.SourceServices, svc)
		}
		if len(plan.SuggestedSteps) > 0 {
			out.SuggestedSteps = append(out.SuggestedSteps, "Assignment steps:")
			out.SuggestedSteps = append(out.SuggestedSteps, plan.SuggestedSteps...)
		}
	}

	sort.Strings(out.SourceServices)
	sortEntityRefs(out.KeyDecisions)
	return out, nil
}

func appendUniqueSource(services []string, svc string) []string {
	for _, existing := range services {
		if existing == svc {
			return services
		}
	}
	return append(services, svc)
}
