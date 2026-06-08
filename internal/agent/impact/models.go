package impact

import (
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	models "github.com/reponerve/reponerve/pkg/models"
)

// ImpactReport holds the result of a deterministic upstream and downstream impact analysis.
type ImpactReport struct {
	EntityID  string                   `json:"entityId"`
	Decisions []*memorymodels.Decision `json:"decisions"`
	Intents   []*memorymodels.Intent   `json:"intents"`
	Facts     []*memorymodels.Fact     `json:"facts"`
	Events    []*models.Event          `json:"events"`
}
