package onboarding

import (
	"context"
	"fmt"

	ctxengine "reponerve/internal/context"
)

// Service provides deterministic repository onboarding packages.
type Service struct {
	generator *ctxengine.Generator
}

// NewService constructs a new onboarding Service.
func NewService(generator *ctxengine.Generator) *Service {
	return &Service{generator: generator}
}

// Generate builds an OnboardingPackage for a given repository.
func (s *Service) Generate(ctx context.Context, repositoryID string) (*OnboardingPackage, error) {
	repoCtx, err := s.generator.Generate(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate repository context: %w", err)
	}

	summary := fmt.Sprintf("Repository Onboarding:\n- %d decisions\n- %d intents\n- %d facts\n- %d events",
		len(repoCtx.Decisions), len(repoCtx.Intents), len(repoCtx.Facts), len(repoCtx.Events))

	return &OnboardingPackage{
		RepositoryID: repositoryID,
		Summary:      summary,
		Decisions:    repoCtx.Decisions,
		Intents:      repoCtx.Intents,
		Facts:        repoCtx.Facts,
		Events:       repoCtx.Events,
	}, nil
}
