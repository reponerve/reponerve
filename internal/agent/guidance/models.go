package guidance

import (
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	models "github.com/reponerve/reponerve/pkg/models"
)

// Guidance represents deterministic architectural guidance for a decision or event.
type Guidance struct {
	EntityID        string                  `json:"entityId"`
	Reasons         []string                `json:"reasons"`
	SupportingFacts []*memorymodels.Fact    `json:"supportingFacts"`
	RelatedIntents  []*memorymodels.Intent  `json:"relatedIntents"`
	RelatedEvents   []*models.Event         `json:"relatedEvents"`
}
