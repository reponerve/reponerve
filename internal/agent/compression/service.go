package compression

import (
	"context"
	"fmt"

	"reponerve/internal/agent/onboarding"
	ctxengine "reponerve/internal/context"
	memorymodels "reponerve/internal/memory/models"
	models "reponerve/pkg/models"
)

// Service provides deterministic context compression.
type Service struct {
	generator         *ctxengine.Generator
	onboardingService *onboarding.Service
}

// NewService constructs a new compression Service.
func NewService(generator *ctxengine.Generator, obs *onboarding.Service) *Service {
	return &Service{
		generator:         generator,
		onboardingService: obs,
	}
}

// Compress returns a CompressedContext with entity lists truncated according to CompressionOptions.
func (s *Service) Compress(ctx context.Context, repositoryID string, opts CompressionOptions) (*CompressedContext, error) {
	// Call generator.Generate to get the sorted repository context.
	repoCtx, err := s.generator.Generate(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate context: %w", err)
	}

	res := &CompressedContext{
		RepositoryID: repositoryID,
		Decisions:    []*memorymodels.Decision{},
		Intents:      []*memorymodels.Intent{},
		Facts:        []*memorymodels.Fact{},
		Events:       []*models.Event{},
	}

	if opts.MaxDecisions > 0 && len(repoCtx.Decisions) > 0 {
		limit := opts.MaxDecisions
		if limit > len(repoCtx.Decisions) {
			limit = len(repoCtx.Decisions)
		}
		res.Decisions = repoCtx.Decisions[:limit]
	}

	if opts.MaxIntents > 0 && len(repoCtx.Intents) > 0 {
		limit := opts.MaxIntents
		if limit > len(repoCtx.Intents) {
			limit = len(repoCtx.Intents)
		}
		res.Intents = repoCtx.Intents[:limit]
	}

	if opts.MaxFacts > 0 && len(repoCtx.Facts) > 0 {
		limit := opts.MaxFacts
		if limit > len(repoCtx.Facts) {
			limit = len(repoCtx.Facts)
		}
		res.Facts = repoCtx.Facts[:limit]
	}

	if opts.MaxEvents > 0 && len(repoCtx.Events) > 0 {
		limit := opts.MaxEvents
		if limit > len(repoCtx.Events) {
			limit = len(repoCtx.Events)
		}
		res.Events = repoCtx.Events[:limit]
	}

	return res, nil
}
