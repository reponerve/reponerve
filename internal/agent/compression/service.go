package compression

import (
	"context"
	"fmt"
	"sort"

	"github.com/reponerve/reponerve/internal/agent/onboarding"
	ctxengine "github.com/reponerve/reponerve/internal/context"
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	"github.com/reponerve/reponerve/internal/query/storage"
	models "github.com/reponerve/reponerve/pkg/models"
)

// Service provides deterministic context compression.
type Service struct {
	generator          *ctxengine.Generator
	onboardingService  *onboarding.Service
	relationshipReader storage.RelationshipReader
}

// NewService constructs a new compression Service.
func NewService(
	generator *ctxengine.Generator,
	obs *onboarding.Service,
	relationshipReader storage.RelationshipReader,
) *Service {
	return &Service{
		generator:          generator,
		onboardingService:  obs,
		relationshipReader: relationshipReader,
	}
}

// Compress returns a CompressedContext with entity lists truncated according to CompressionOptions.
func (s *Service) Compress(ctx context.Context, repositoryID string, opts CompressionOptions) (*CompressedContext, error) {
	repoCtx, err := s.generator.Generate(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate context: %w", err)
	}

	tokens := topicTokens(opts.Topic)
	scores := make(map[string]int)

	decisions := rankDecisions(repoCtx.Decisions, tokens, scores)
	intents := rankIntents(repoCtx.Intents, tokens, scores)
	facts := rankFacts(repoCtx.Facts, tokens, scores)
	events := rankEvents(repoCtx.Events, tokens, scores)

	if len(tokens) > 0 && s.relationshipReader != nil {
		rels, relErr := s.relationshipReader.ListByRepository(ctx, repositoryID)
		if relErr == nil {
			edges := make([][2]string, 0, len(rels))
			for _, rel := range rels {
				edges = append(edges, [2]string{rel.FromID, rel.ToID})
			}
			applyRelationshipBoost(scores, edges)
			decisions = rerankDecisions(decisions, scores)
			intents = rerankIntents(intents, scores)
			facts = rerankFacts(facts, scores)
			events = rerankEvents(events, scores)
		}
	}

	res := &CompressedContext{
		RepositoryID: repositoryID,
		Decisions:    []*memorymodels.Decision{},
		Intents:      []*memorymodels.Intent{},
		Facts:        []*memorymodels.Fact{},
		Events:       []*models.Event{},
	}

	if opts.TokenBudget > 0 {
		res.Decisions, res.Intents, res.Facts, res.Events = packByTokenBudget(
			decisions, intents, facts, events, scores, opts.TokenBudget,
		)
	} else {
		res.Decisions = truncateDecisions(decisions, opts.MaxDecisions)
		res.Intents = truncateIntents(intents, opts.MaxIntents)
		res.Facts = truncateFacts(facts, opts.MaxFacts)
		res.Events = truncateEvents(events, opts.MaxEvents)
	}

	return res, nil
}

func rankDecisions(items []*memorymodels.Decision, tokens []string, scores map[string]int) []*memorymodels.Decision {
	if len(items) == 0 {
		return items
	}
	out := make([]*memorymodels.Decision, len(items))
	copy(out, items)
	if len(tokens) == 0 {
		return out
	}
	for _, d := range out {
		scores[d.ID] += scoreText(tokens, d.Title, d.Status)
	}
	sort.SliceStable(out, func(i, j int) bool {
		si, sj := scores[out[i].ID], scores[out[j].ID]
		if si != sj {
			return si > sj
		}
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func rerankDecisions(items []*memorymodels.Decision, scores map[string]int) []*memorymodels.Decision {
	if len(items) == 0 {
		return items
	}
	out := make([]*memorymodels.Decision, len(items))
	copy(out, items)
	sort.SliceStable(out, func(i, j int) bool {
		si, sj := scores[out[i].ID], scores[out[j].ID]
		if si != sj {
			return si > sj
		}
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func rankIntents(items []*memorymodels.Intent, tokens []string, scores map[string]int) []*memorymodels.Intent {
	if len(items) == 0 {
		return items
	}
	out := make([]*memorymodels.Intent, len(items))
	copy(out, items)
	if len(tokens) == 0 {
		return out
	}
	for _, it := range out {
		scores[it.ID] += scoreText(tokens, it.Description)
	}
	sort.SliceStable(out, func(i, j int) bool {
		si, sj := scores[out[i].ID], scores[out[j].ID]
		if si != sj {
			return si > sj
		}
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func rerankIntents(items []*memorymodels.Intent, scores map[string]int) []*memorymodels.Intent {
	if len(items) == 0 {
		return items
	}
	out := make([]*memorymodels.Intent, len(items))
	copy(out, items)
	sort.SliceStable(out, func(i, j int) bool {
		si, sj := scores[out[i].ID], scores[out[j].ID]
		if si != sj {
			return si > sj
		}
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func rankFacts(items []*memorymodels.Fact, tokens []string, scores map[string]int) []*memorymodels.Fact {
	if len(items) == 0 {
		return items
	}
	out := make([]*memorymodels.Fact, len(items))
	copy(out, items)
	if len(tokens) == 0 {
		return out
	}
	for _, f := range out {
		scores[f.ID] += scoreText(tokens, f.Subject, f.Predicate, f.Object)
	}
	sort.SliceStable(out, func(i, j int) bool {
		si, sj := scores[out[i].ID], scores[out[j].ID]
		if si != sj {
			return si > sj
		}
		if out[i].Subject == out[j].Subject {
			return out[i].ID < out[j].ID
		}
		return out[i].Subject < out[j].Subject
	})
	return out
}

func rerankFacts(items []*memorymodels.Fact, scores map[string]int) []*memorymodels.Fact {
	if len(items) == 0 {
		return items
	}
	out := make([]*memorymodels.Fact, len(items))
	copy(out, items)
	sort.SliceStable(out, func(i, j int) bool {
		si, sj := scores[out[i].ID], scores[out[j].ID]
		if si != sj {
			return si > sj
		}
		if out[i].Subject == out[j].Subject {
			return out[i].ID < out[j].ID
		}
		return out[i].Subject < out[j].Subject
	})
	return out
}

func rankEvents(items []*models.Event, tokens []string, scores map[string]int) []*models.Event {
	if len(items) == 0 {
		return nil
	}
	out := make([]*models.Event, len(items))
	copy(out, items)
	if len(tokens) == 0 {
		return out
	}
	for _, e := range out {
		scores[e.ID] += scoreText(tokens, e.Title, string(e.EventType))
	}
	sort.SliceStable(out, func(i, j int) bool {
		si, sj := scores[out[i].ID], scores[out[j].ID]
		if si != sj {
			return si > sj
		}
		if out[i].Timestamp.Equal(out[j].Timestamp) {
			return out[i].ID > out[j].ID
		}
		return out[i].Timestamp.After(out[j].Timestamp)
	})
	return out
}

func rerankEvents(items []*models.Event, scores map[string]int) []*models.Event {
	if len(items) == 0 {
		return items
	}
	out := make([]*models.Event, len(items))
	copy(out, items)
	sort.SliceStable(out, func(i, j int) bool {
		si, sj := scores[out[i].ID], scores[out[j].ID]
		if si != sj {
			return si > sj
		}
		if out[i].Timestamp.Equal(out[j].Timestamp) {
			return out[i].ID > out[j].ID
		}
		return out[i].Timestamp.After(out[j].Timestamp)
	})
	return out
}

type scoredItem struct {
	kind  string
	score int
	text  string
	dec   *memorymodels.Decision
	intent *memorymodels.Intent
	fact  *memorymodels.Fact
	event *models.Event
}

func packByTokenBudget(
	decisions []*memorymodels.Decision,
	intents []*memorymodels.Intent,
	facts []*memorymodels.Fact,
	events []*models.Event,
	scores map[string]int,
	budget int,
) ([]*memorymodels.Decision, []*memorymodels.Intent, []*memorymodels.Fact, []*models.Event) {
	items := make([]scoredItem, 0, len(decisions)+len(intents)+len(facts)+len(events))
	for _, d := range decisions {
		items = append(items, scoredItem{
			kind: "decision", score: scores[d.ID], text: d.Title,
			dec: d,
		})
	}
	for _, it := range intents {
		items = append(items, scoredItem{
			kind: "intent", score: scores[it.ID], text: it.Description,
			intent: it,
		})
	}
	for _, f := range facts {
		items = append(items, scoredItem{
			kind: "fact", score: scores[f.ID],
			text: f.Subject + " " + f.Predicate + " " + f.Object,
			fact: f,
		})
	}
	for _, e := range events {
		items = append(items, scoredItem{
			kind: "event", score: scores[e.ID], text: e.Title,
			event: e,
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].score != items[j].score {
			return items[i].score > items[j].score
		}
		return items[i].text < items[j].text
	})

	var outDec []*memorymodels.Decision
	var outInt []*memorymodels.Intent
	var outFact []*memorymodels.Fact
	var outEvt []*models.Event
	used := 0
	for _, item := range items {
		cost := estimateTokens(item.text)
		if used+cost > budget && used > 0 {
			continue
		}
		if used+cost > budget {
			continue
		}
		used += cost
		switch item.kind {
		case "decision":
			outDec = append(outDec, item.dec)
		case "intent":
			outInt = append(outInt, item.intent)
		case "fact":
			outFact = append(outFact, item.fact)
		case "event":
			outEvt = append(outEvt, item.event)
		}
	}
	return outDec, outInt, outFact, outEvt
}

func truncateDecisions(items []*memorymodels.Decision, limit int) []*memorymodels.Decision {
	if limit <= 0 || len(items) == 0 {
		return []*memorymodels.Decision{}
	}
	if limit >= len(items) {
		return items
	}
	return items[:limit]
}

func truncateIntents(items []*memorymodels.Intent, limit int) []*memorymodels.Intent {
	if limit <= 0 || len(items) == 0 {
		return []*memorymodels.Intent{}
	}
	if limit >= len(items) {
		return items
	}
	return items[:limit]
}

func truncateFacts(items []*memorymodels.Fact, limit int) []*memorymodels.Fact {
	if limit <= 0 || len(items) == 0 {
		return []*memorymodels.Fact{}
	}
	if limit >= len(items) {
		return items
	}
	return items[:limit]
}

func truncateEvents(items []*models.Event, limit int) []*models.Event {
	if limit <= 0 || len(items) == 0 {
		return []*models.Event{}
	}
	if limit >= len(items) {
		return items
	}
	return items[:limit]
}
