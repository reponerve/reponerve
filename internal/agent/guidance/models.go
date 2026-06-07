package guidance

import (
	memorymodels "reponerve/internal/memory/models"
	models "reponerve/pkg/models"
)

// Guidance represents deterministic architectural guidance for a decision or event.
type Guidance struct {
	EntityID        string                  `json:"entityId"`
	Reasons         []string                `json:"reasons"`
	SupportingFacts []*memorymodels.Fact    `json:"supportingFacts"`
	RelatedIntents  []*memorymodels.Intent  `json:"relatedIntents"`
	RelatedEvents   []*models.Event         `json:"relatedEvents"`
}
