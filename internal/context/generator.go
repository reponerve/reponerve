package context

import (
	stdcontext "context"
	"sort"
	"time"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	models "github.com/reponerve/reponerve/pkg/models"
)

type Generator struct {
	reader ContextReader
}

func NewGenerator(reader ContextReader) *Generator {
	return &Generator{reader: reader}
}

// Generate aggregates memory entities and returns a RepositoryContext with deterministic sorting.
func (g *Generator) Generate(ctx stdcontext.Context, repositoryID string) (*RepositoryContext, error) {
	data, err := g.reader.ReadContext(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	decisions := make([]*memorymodels.Decision, len(data.Decisions))
	copy(decisions, data.Decisions)

	intents := make([]*memorymodels.Intent, len(data.Intents))
	copy(intents, data.Intents)

	events := make([]*models.Event, len(data.Events))
	copy(events, data.Events)

	facts := make([]*memorymodels.Fact, len(data.Facts))
	copy(facts, data.Facts)

	// Sort Decisions: Most recent first (CreatedAt descending), fallback to ID descending
	sort.Slice(decisions, func(i, j int) bool {
		if decisions[i].CreatedAt.Equal(decisions[j].CreatedAt) {
			return decisions[i].ID > decisions[j].ID
		}
		return decisions[i].CreatedAt.After(decisions[j].CreatedAt)
	})

	// Sort Intents: Most recent first (CreatedAt descending), fallback to ID descending
	sort.Slice(intents, func(i, j int) bool {
		if intents[i].CreatedAt.Equal(intents[j].CreatedAt) {
			return intents[i].ID > intents[j].ID
		}
		return intents[i].CreatedAt.After(intents[j].CreatedAt)
	})

	// Sort Events: Most recent first (Timestamp descending), fallback to ID descending
	sort.Slice(events, func(i, j int) bool {
		if events[i].Timestamp.Equal(events[j].Timestamp) {
			return events[i].ID > events[j].ID
		}
		return events[i].Timestamp.After(events[j].Timestamp)
	})

	// Sort Facts: Alphabetical by Subject (ascending), fallback to ID ascending
	sort.Slice(facts, func(i, j int) bool {
		if facts[i].Subject == facts[j].Subject {
			return facts[i].ID < facts[j].ID
		}
		return facts[i].Subject < facts[j].Subject
	})

	return &RepositoryContext{
		RepositoryID: repositoryID,
		GeneratedAt:  time.Now(),
		Decisions:    decisions,
		Intents:      intents,
		Facts:        facts,
		Events:       events,
	}, nil
}
