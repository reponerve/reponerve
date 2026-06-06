package context

import (
	stdcontext "context"

	memorymodels "reponerve/internal/memory/models"
	"reponerve/internal/query/storage"
	models "reponerve/pkg/models"
)

type MemoryContextReader struct {
	eventReader    storage.EventReader
	decisionReader storage.DecisionReader
	intentReader   storage.IntentReader
	factReader     storage.FactReader
}

func NewMemoryContextReader(
	er storage.EventReader,
	dr storage.DecisionReader,
	ir storage.IntentReader,
	fr storage.FactReader,
) *MemoryContextReader {
	return &MemoryContextReader{
		eventReader:    er,
		decisionReader: dr,
		intentReader:   ir,
		factReader:     fr,
	}
}

func (r *MemoryContextReader) ListEvents(ctx stdcontext.Context, repositoryID string) ([]*models.Event, error) {
	return r.eventReader.ListByRepository(ctx, repositoryID)
}

func (r *MemoryContextReader) ListDecisions(ctx stdcontext.Context, repositoryID string) ([]*memorymodels.Decision, error) {
	return r.decisionReader.ListByRepository(ctx, repositoryID)
}

func (r *MemoryContextReader) ListIntents(ctx stdcontext.Context, repositoryID string) ([]*memorymodels.Intent, error) {
	return r.intentReader.ListByRepository(ctx, repositoryID)
}

func (r *MemoryContextReader) ListFacts(ctx stdcontext.Context, repositoryID string) ([]*memorymodels.Fact, error) {
	return r.factReader.ListByRepository(ctx, repositoryID)
}

func (r *MemoryContextReader) ReadContext(ctx stdcontext.Context, repositoryID string) (*ContextData, error) {
	events, err := r.ListEvents(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	decisions, err := r.ListDecisions(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	intents, err := r.ListIntents(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	facts, err := r.ListFacts(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	return &ContextData{
		RepositoryID: repositoryID,
		Events:       events,
		Decisions:    decisions,
		Intents:      intents,
		Facts:        facts,
	}, nil
}
