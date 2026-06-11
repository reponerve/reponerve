package development

import (
	"context"
	"fmt"
	"sort"
	"strings"

	agentsearch "github.com/reponerve/reponerve/internal/agent/search"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/internal/query/storage"
)

// Router resolves natural-language topics across repository and code authorities.
type Router struct {
	searchService  *agentsearch.Service
	codeEntityReader storage.CodeEntityReader
	repoCodeReader   storage.RepositoryCodeRelationshipReader
}

// NewRouter creates a topic resolver.
func NewRouter(
	searchService *agentsearch.Service,
	codeEntityReader storage.CodeEntityReader,
	repoCodeReader storage.RepositoryCodeRelationshipReader,
) *Router {
	return &Router{
		searchService:    searchService,
		codeEntityReader: codeEntityReader,
		repoCodeReader:   repoCodeReader,
	}
}

// ResolveTopic normalizes input and resolves repository and code matches in parallel.
func (r *Router) ResolveTopic(ctx context.Context, repositoryID, input string) (*ResolvedTopic, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	normalized := normalizeTopic(input)
	if normalized == "" {
		return nil, fmt.Errorf("topic cannot be empty")
	}

	topic := &ResolvedTopic{
		Input:            normalized,
		RepositoryHitIDs: make(map[string]struct{}),
		CodeEntityIDs:    make(map[string]struct{}),
	}

	searchResult, err := r.searchService.Search(ctx, repositoryID, normalized)
	if err != nil {
		return nil, fmt.Errorf("repository search: %w", err)
	}
	for _, hit := range searchResult.Hits {
		topic.RepositoryHitIDs[hit.EntityID] = struct{}{}
	}

	codeMatches, err := r.searchCodeEntities(ctx, repositoryID, normalized)
	if err != nil {
		return nil, fmt.Errorf("code search: %w", err)
	}
	for _, entity := range codeMatches {
		topic.CodeEntityIDs[entity.ID] = struct{}{}
	}

	if err := r.expandRepositoryCodeLinks(ctx, repositoryID, topic); err != nil {
		return nil, err
	}

	topic.PrimaryEntityType = classifyPrimaryEntityType(topic)
	topic.MatchEvidence = buildMatchEvidence(normalized, len(searchResult.Hits), len(codeMatches))
	return topic, nil
}

func (r *Router) searchCodeEntities(ctx context.Context, repositoryID, topic string) ([]*codemodels.CodeEntity, error) {
	terms := topicTerms(topic)
	if len(terms) == 0 {
		return nil, nil
	}

	all, err := r.codeEntityReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	type scored struct {
		entity *codemodels.CodeEntity
		score  int
	}
	var matches []scored
	for _, e := range all {
		score := scoreCodeEntity(e, terms)
		if score > 0 {
			matches = append(matches, scored{entity: e, score: score})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		if matches[i].entity.EntityType != matches[j].entity.EntityType {
			return matches[i].entity.EntityType < matches[j].entity.EntityType
		}
		if matches[i].entity.QualifiedName != matches[j].entity.QualifiedName {
			return matches[i].entity.QualifiedName < matches[j].entity.QualifiedName
		}
		return matches[i].entity.ID < matches[j].entity.ID
	})

	out := make([]*codemodels.CodeEntity, 0, len(matches))
	for _, m := range matches {
		out = append(out, m.entity)
	}
	return out, nil
}

func (r *Router) expandRepositoryCodeLinks(ctx context.Context, repositoryID string, topic *ResolvedTopic) error {
	allLinks, err := r.repoCodeReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return err
	}

	seen := make(map[string]struct{})
	for _, link := range allLinks {
		_, repoHit := topic.RepositoryHitIDs[link.RepositoryEntityID]
		_, codeHit := topic.CodeEntityIDs[link.CodeEntityID]
		if !repoHit && !codeHit {
			continue
		}
		key := link.ID
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		topic.RepositoryCodeLinks = append(topic.RepositoryCodeLinks, link)
		topic.RepositoryHitIDs[link.RepositoryEntityID] = struct{}{}
		topic.CodeEntityIDs[link.CodeEntityID] = struct{}{}
	}
	sort.Slice(topic.RepositoryCodeLinks, func(i, j int) bool {
		a, b := topic.RepositoryCodeLinks[i], topic.RepositoryCodeLinks[j]
		if a.RelationshipType != b.RelationshipType {
			return a.RelationshipType < b.RelationshipType
		}
		if a.RepositoryEntityID != b.RepositoryEntityID {
			return a.RepositoryEntityID < b.RepositoryEntityID
		}
		return a.CodeEntityID < b.CodeEntityID
	})
	return nil
}

func normalizeTopic(input string) string {
	input = strings.TrimSpace(input)
	input = strings.Join(strings.Fields(input), " ")
	return strings.ToLower(input)
}

func topicTerms(topic string) []string {
	parts := strings.Fields(topic)
	var terms []string
	for _, p := range parts {
		if len(p) < 2 {
			continue
		}
		terms = append(terms, p)
	}
	return terms
}

func scoreCodeEntity(e *codemodels.CodeEntity, terms []string) int {
	name := strings.ToLower(e.Name)
	qualified := strings.ToLower(e.QualifiedName)
	filePath := strings.ToLower(e.FilePath)
	score := 0
	for _, term := range terms {
		switch {
		case name == term:
			score += 100
		case qualified == term:
			score += 100
		case strings.HasSuffix(qualified, "."+term):
			score += 80
		case strings.Contains(name, term):
			score += 50
		case strings.Contains(qualified, term):
			score += 40
		case strings.Contains(filePath, term):
			score += 25
		}
	}
	return score
}

func classifyPrimaryEntityType(topic *ResolvedTopic) string {
	hasRepo := len(topic.RepositoryHitIDs) > 0
	hasCode := len(topic.CodeEntityIDs) > 0
	switch {
	case hasRepo && hasCode:
		return "mixed"
	case hasCode:
		return "code"
	case hasRepo:
		return "repository"
	default:
		return "none"
	}
}

func buildMatchEvidence(topic string, repoHits, codeHits int) string {
	return fmt.Sprintf("topic=%q repository_hits=%d code_hits=%d", topic, repoHits, codeHits)
}
