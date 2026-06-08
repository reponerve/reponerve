package context

import (
	"time"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	models "github.com/reponerve/reponerve/pkg/models"
)

type RepositoryContext struct {
	RepositoryID string
	GeneratedAt  time.Time
	Decisions    []*memorymodels.Decision
	Intents      []*memorymodels.Intent
	Facts        []*memorymodels.Fact
	Events       []*models.Event
}
