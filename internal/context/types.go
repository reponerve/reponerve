package context

import (
	stdcontext "context"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	models "github.com/reponerve/reponerve/pkg/models"
)

type ContextData struct {
	RepositoryID string
	Decisions    []*memorymodels.Decision
	Intents      []*memorymodels.Intent
	Facts        []*memorymodels.Fact
	Events       []*models.Event
}

type EventContextReader interface {
	ListEvents(ctx stdcontext.Context, repositoryID string) ([]*models.Event, error)
}

type DecisionContextReader interface {
	ListDecisions(ctx stdcontext.Context, repositoryID string) ([]*memorymodels.Decision, error)
}

type IntentContextReader interface {
	ListIntents(ctx stdcontext.Context, repositoryID string) ([]*memorymodels.Intent, error)
}

type FactContextReader interface {
	ListFacts(ctx stdcontext.Context, repositoryID string) ([]*memorymodels.Fact, error)
}

type ContextReader interface {
	ReadContext(ctx stdcontext.Context, repositoryID string) (*ContextData, error)
}
