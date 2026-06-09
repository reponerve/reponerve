package onboarding

import (
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	models "github.com/reponerve/reponerve/pkg/models"
)

// OnboardingPackage represents a structured repository onboarding snapshot.
type OnboardingPackage struct {
	RepositoryID string                   `json:"repositoryId"`
	Summary      string                   `json:"summary"`
	Decisions    []*memorymodels.Decision `json:"decisions"`
	Intents      []*memorymodels.Intent   `json:"intents"`
	Facts        []*memorymodels.Fact     `json:"facts"`
	Events       []*models.Event          `json:"events"`
}
