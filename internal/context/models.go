package context

import (
	"time"

	memorymodels "reponerve/internal/memory/models"
	models "reponerve/pkg/models"
)

type RepositoryContext struct {
	RepositoryID string
	GeneratedAt  time.Time
	Decisions    []*memorymodels.Decision
	Intents      []*memorymodels.Intent
	Facts        []*memorymodels.Fact
	Events       []*models.Event
}
