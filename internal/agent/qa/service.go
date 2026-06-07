package qa

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"reponerve/internal/agent/guidance"
	"reponerve/internal/agent/impact"
	"reponerve/internal/agent/onboarding"
)

var (
	rxOverview1 = regexp.MustCompile(`(?i)^what is this repository\??$`)
	rxOverview2 = regexp.MustCompile(`(?i)^show repository overview\??$`)

	rxDecisionWhy     = regexp.MustCompile(`(?i)^why was decision\s+([a-zA-Z0-9_\-]+)\s+made\??$`)
	rxDecisionSupport = regexp.MustCompile(`(?i)^what supports decision\s+([a-zA-Z0-9_\-]+)\??$`)

	rxEventCause = regexp.MustCompile(`(?i)^what caused event\s+([a-zA-Z0-9_\-]+)\??$`)
	rxEventWhy   = regexp.MustCompile(`(?i)^why did event\s+([a-zA-Z0-9_\-]+)\s+happen\??$`)

	rxImpactDecision = regexp.MustCompile(`(?i)^what happens if decision\s+([a-zA-Z0-9_\-]+)\s+changes\??$`)
	rxImpactFact     = regexp.MustCompile(`(?i)^what depends on fact\s+([a-zA-Z0-9_\-]+)\??$`)
	rxImpactIntent   = regexp.MustCompile(`(?i)^what happens if intent\s+([a-zA-Z0-9_\-]+)\s+changes\??$`)
	rxImpactEvent    = regexp.MustCompile(`(?i)^what is the impact of event\s+([a-zA-Z0-9_\-]+)\??$`)
)

// Service routes supported repository questions to the corresponding domain intelligence service.
type Service struct {
	onboardingService *onboarding.Service
	guidanceService   *guidance.Service
	impactService     *impact.Service
}

// NewService constructs a new QA Service.
func NewService(
	obs *onboarding.Service,
	gs *guidance.Service,
	is *impact.Service,
) *Service {
	return &Service{
		onboardingService: obs,
		guidanceService:   gs,
		impactService:     is,
	}
}

// Answer processes a repository Question query and returns a structured Answer.
func (s *Service) Answer(ctx context.Context, repositoryID string, q Question) (*Answer, error) {
	trimmed := strings.TrimSpace(q.Text)

	// 1. Repository Overview
	if rxOverview1.MatchString(trimmed) || rxOverview2.MatchString(trimmed) {
		res, err := s.onboardingService.Generate(ctx, repositoryID)
		if err != nil {
			return nil, err
		}
		return &Answer{Question: q.Text, Result: res}, nil
	}

	// 2. Decision Guidance
	if matches := rxDecisionWhy.FindStringSubmatch(trimmed); len(matches) > 1 {
		res, err := s.guidanceService.GetDecisionGuidance(ctx, matches[1])
		if err != nil {
			return nil, err
		}
		return &Answer{Question: q.Text, Result: res}, nil
	}
	if matches := rxDecisionSupport.FindStringSubmatch(trimmed); len(matches) > 1 {
		res, err := s.guidanceService.GetDecisionGuidance(ctx, matches[1])
		if err != nil {
			return nil, err
		}
		return &Answer{Question: q.Text, Result: res}, nil
	}

	// 3. Event Guidance
	if matches := rxEventCause.FindStringSubmatch(trimmed); len(matches) > 1 {
		res, err := s.guidanceService.GetEventGuidance(ctx, matches[1])
		if err != nil {
			return nil, err
		}
		return &Answer{Question: q.Text, Result: res}, nil
	}
	if matches := rxEventWhy.FindStringSubmatch(trimmed); len(matches) > 1 {
		res, err := s.guidanceService.GetEventGuidance(ctx, matches[1])
		if err != nil {
			return nil, err
		}
		return &Answer{Question: q.Text, Result: res}, nil
	}

	// 4. Impact Analysis
	if matches := rxImpactDecision.FindStringSubmatch(trimmed); len(matches) > 1 {
		res, err := s.impactService.AnalyzeDecisionImpact(ctx, matches[1])
		if err != nil {
			return nil, err
		}
		return &Answer{Question: q.Text, Result: res}, nil
	}
	if matches := rxImpactFact.FindStringSubmatch(trimmed); len(matches) > 1 {
		res, err := s.impactService.AnalyzeFactImpact(ctx, matches[1])
		if err != nil {
			return nil, err
		}
		return &Answer{Question: q.Text, Result: res}, nil
	}
	if matches := rxImpactIntent.FindStringSubmatch(trimmed); len(matches) > 1 {
		res, err := s.impactService.AnalyzeIntentImpact(ctx, matches[1])
		if err != nil {
			return nil, err
		}
		return &Answer{Question: q.Text, Result: res}, nil
	}
	if matches := rxImpactEvent.FindStringSubmatch(trimmed); len(matches) > 1 {
		res, err := s.impactService.AnalyzeEventImpact(ctx, matches[1])
		if err != nil {
			return nil, err
		}
		return &Answer{Question: q.Text, Result: res}, nil
	}

	return nil, fmt.Errorf("unknown question: %s", q.Text)
}
