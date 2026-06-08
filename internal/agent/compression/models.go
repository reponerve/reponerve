package compression

import (
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	models "github.com/reponerve/reponerve/pkg/models"
)

// CompressionOptions specifies limit thresholds for each entity type in the context.
type CompressionOptions struct {
	MaxDecisions int `json:"maxDecisions"`
	MaxIntents   int `json:"maxIntents"`
	MaxFacts     int `json:"maxFacts"`
	MaxEvents    int `json:"maxEvents"`
}

// CompressedContext holds the deterministically truncated repository context.
type CompressedContext struct {
	RepositoryID string                    `json:"repositoryId"`
	Decisions    []*memorymodels.Decision  `json:"decisions"`
	Intents      []*memorymodels.Intent    `json:"intents"`
	Facts        []*memorymodels.Fact      `json:"facts"`
	Events       []*models.Event           `json:"events"`
}
