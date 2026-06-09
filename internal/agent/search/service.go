package agentsearch

import (
	stdcontext "context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/reponerve/reponerve/internal/intelligence/discovery"
	"github.com/reponerve/reponerve/internal/query/storage"
	models "github.com/reponerve/reponerve/pkg/models"
)

var supportedPrefixes = map[string]bool{
	"type":   true,
	"domain": true,
}

var typeFilterMap = map[string]string{
	"decision":     EntityTypeDecision,
	"fact":         EntityTypeFact,
	"event":        EntityTypeEvent,
	"contributor":  EntityTypeContributor,
	"expertise":    EntityTypeExpertise,
	"relationship": EntityTypeRelationship,
}

var sourcePrecedence = map[string]int{
	SourceMemory:    4,
	SourceOwnership: 3,
	SourceGraph:     2,
	SourceDiscovery: 1,
}

type parsedQuery struct {
	typeFilter   string
	domainFilter string
	terms        []string
}

type fieldSpec struct {
	name  string
	value string
	weak  bool
}

type matchEvidence struct {
	MatchType string `json:"match_type"`
	Field     string `json:"field"`
}

// Service provides deterministic repository knowledge retrieval.
type Service struct {
	decisionReader     storage.DecisionReader
	factReader         storage.FactReader
	eventReader        storage.EventReader
	relationshipReader storage.RelationshipReader
	contributorReader  storage.ContributorReader
	expertiseReader    storage.ExpertiseReader
	discoveryService   *discovery.Service
}

// NewService constructs a new Repository Search Service.
func NewService(
	dr storage.DecisionReader,
	fr storage.FactReader,
	er storage.EventReader,
	rr storage.RelationshipReader,
	cr storage.ContributorReader,
	expr storage.ExpertiseReader,
	discoverySvc *discovery.Service,
) *Service {
	return &Service{
		decisionReader:     dr,
		factReader:         fr,
		eventReader:        er,
		relationshipReader: rr,
		contributorReader:  cr,
		expertiseReader:    expr,
		discoveryService:   discoverySvc,
	}
}

// Search retrieves repository knowledge matching the given query.
func (s *Service) Search(ctx stdcontext.Context, repositoryID string, query string) (*SearchResult, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	parsed, err := parseQuery(query)
	if err != nil {
		return nil, err
	}

	var hits []*SearchHit

	if s.shouldSearchType(parsed.typeFilter, EntityTypeDecision) {
		decisionHits, err := s.searchDecisions(ctx, repositoryID, parsed)
		if err != nil {
			return nil, fmt.Errorf("failed to search decisions: %w", err)
		}
		hits = append(hits, decisionHits...)
	}

	if s.shouldSearchType(parsed.typeFilter, EntityTypeFact) {
		factHits, err := s.searchFacts(ctx, repositoryID, parsed)
		if err != nil {
			return nil, fmt.Errorf("failed to search facts: %w", err)
		}
		hits = append(hits, factHits...)
	}

	if s.shouldSearchType(parsed.typeFilter, EntityTypeEvent) {
		eventHits, err := s.searchEvents(ctx, repositoryID, parsed)
		if err != nil {
			return nil, fmt.Errorf("failed to search events: %w", err)
		}
		hits = append(hits, eventHits...)
	}

	if s.shouldSearchType(parsed.typeFilter, EntityTypeContributor) && parsed.domainFilter == "" {
		contribHits, err := s.searchContributors(ctx, repositoryID, parsed)
		if err != nil {
			return nil, fmt.Errorf("failed to search contributors: %w", err)
		}
		hits = append(hits, contribHits...)
	}

	if s.shouldSearchType(parsed.typeFilter, EntityTypeExpertise) || parsed.domainFilter != "" {
		expertiseHits, err := s.searchExpertise(ctx, repositoryID, parsed)
		if err != nil {
			return nil, fmt.Errorf("failed to search expertise: %w", err)
		}
		hits = append(hits, expertiseHits...)
	}

	if s.shouldSearchType(parsed.typeFilter, EntityTypeRelationship) && parsed.domainFilter == "" {
		relHits, err := s.searchRelationships(ctx, repositoryID, parsed)
		if err != nil {
			return nil, fmt.Errorf("failed to search relationships: %w", err)
		}
		hits = append(hits, relHits...)
	}

	if s.discoveryService != nil && parsed.domainFilter == "" {
		discoveryHits, err := s.searchDiscovery(ctx, repositoryID, parsed)
		if err != nil {
			return nil, fmt.Errorf("failed to search discovery: %w", err)
		}
		hits = append(hits, discoveryHits...)
	}

	hits = deduplicateHits(hits)
	sortHits(hits)

	if hits == nil {
		hits = make([]*SearchHit, 0)
	}

	result := &SearchResult{
		RepositoryID: repositoryID,
		Query:        query,
		Hits:         hits,
	}

	if err := ValidateResult(result); err != nil {
		return nil, fmt.Errorf("generated search result is invalid: %w", err)
	}

	return result, nil
}

func parseQuery(query string) (*parsedQuery, error) {
	pq := &parsedQuery{}
	tokens := strings.Fields(strings.TrimSpace(query))

	for _, token := range tokens {
		colonIdx := strings.Index(token, ":")
		if colonIdx <= 0 {
			pq.terms = append(pq.terms, token)
			continue
		}

		prefix := strings.ToLower(token[:colonIdx])
		value := token[colonIdx+1:]

		if !supportedPrefixes[prefix] {
			return nil, fmt.Errorf("unknown prefix: %s", prefix)
		}

		switch prefix {
		case "type":
			entityType, ok := typeFilterMap[strings.ToLower(value)]
			if !ok {
				return nil, fmt.Errorf("unknown type filter: %s", value)
			}
			pq.typeFilter = entityType
		case "domain":
			if value == "" {
				return nil, fmt.Errorf("domain filter value cannot be empty")
			}
			pq.domainFilter = value
		}
	}

	return pq, nil
}

func (s *Service) shouldSearchType(typeFilter, entityType string) bool {
	return typeFilter == "" || typeFilter == entityType
}

func (s *Service) searchDecisions(ctx stdcontext.Context, repositoryID string, pq *parsedQuery) ([]*SearchHit, error) {
	decisions, err := s.decisionReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	var hits []*SearchHit
	for _, d := range decisions {
		fields := []fieldSpec{
			{"id", d.ID, false},
			{"title", d.Title, false},
			{"status", d.Status, true},
		}

		if hit := matchEntity(EntityTypeDecision, d.ID, SourceMemory, fields, pq); hit != nil {
			hits = append(hits, hit)
		} else if pq.typeFilter == EntityTypeDecision && len(pq.terms) == 0 && pq.domainFilter == "" {
			hits = append(hits, typeFilterHit(EntityTypeDecision, d.ID, SourceMemory))
		}
	}
	return hits, nil
}

func (s *Service) searchFacts(ctx stdcontext.Context, repositoryID string, pq *parsedQuery) ([]*SearchHit, error) {
	facts, err := s.factReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	var hits []*SearchHit
	for _, f := range facts {
		fields := []fieldSpec{
			{"subject", f.Subject, false},
			{"predicate", f.Predicate, false},
			{"object", f.Object, false},
			{"id", f.ID, true},
		}

		if hit := matchEntity(EntityTypeFact, f.ID, SourceMemory, fields, pq); hit != nil {
			hits = append(hits, hit)
		} else if pq.typeFilter == EntityTypeFact && len(pq.terms) == 0 && pq.domainFilter == "" {
			hits = append(hits, typeFilterHit(EntityTypeFact, f.ID, SourceMemory))
		}
	}
	return hits, nil
}

func (s *Service) searchEvents(ctx stdcontext.Context, repositoryID string, pq *parsedQuery) ([]*SearchHit, error) {
	events, err := s.eventReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	var hits []*SearchHit
	for _, ev := range events {
		fields := []fieldSpec{
			{"id", ev.ID, false},
			{"description", ev.Description, false},
			{"title", ev.Title, true},
		}

		if hit := matchEntity(EntityTypeEvent, ev.ID, SourceMemory, fields, pq); hit != nil {
			hits = append(hits, hit)
		} else if pq.typeFilter == EntityTypeEvent && len(pq.terms) == 0 && pq.domainFilter == "" {
			hits = append(hits, typeFilterHit(EntityTypeEvent, ev.ID, SourceMemory))
		}
	}
	return hits, nil
}

func (s *Service) searchContributors(ctx stdcontext.Context, repositoryID string, pq *parsedQuery) ([]*SearchHit, error) {
	contribs, err := s.contributorReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	var hits []*SearchHit
	for _, c := range contribs {
		fields := []fieldSpec{
			{"name", c.Name, false},
			{"email", c.Email, false},
			{"id", c.ID, true},
		}

		if hit := matchEntity(EntityTypeContributor, c.ID, SourceOwnership, fields, pq); hit != nil {
			hits = append(hits, hit)
		} else if pq.typeFilter == EntityTypeContributor && len(pq.terms) == 0 {
			hits = append(hits, typeFilterHit(EntityTypeContributor, c.ID, SourceOwnership))
		}
	}
	return hits, nil
}

func (s *Service) searchExpertise(ctx stdcontext.Context, repositoryID string, pq *parsedQuery) ([]*SearchHit, error) {
	exps, err := s.expertiseReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	contribMap, err := s.buildContributorMap(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	var hits []*SearchHit
	for _, exp := range exps {
		if pq.domainFilter != "" && !domainMatches(exp.Domain, pq.domainFilter) {
			continue
		}

		fields := []fieldSpec{
			{"domain", exp.Domain, false},
			{"id", exp.ID, true},
			{"contributor_id", exp.ContributorID, true},
		}

		if c, ok := contribMap[exp.ContributorID]; ok {
			fields = append(fields,
				fieldSpec{"contributor_name", c.Name, false},
				fieldSpec{"contributor_email", c.Email, false},
			)
		}

		searchPQ := pq
		if pq.domainFilter != "" && len(pq.terms) == 0 {
			searchPQ = &parsedQuery{
				typeFilter:   pq.typeFilter,
				domainFilter: pq.domainFilter,
				terms:        []string{pq.domainFilter},
			}
		}

		if hit := matchEntity(EntityTypeExpertise, exp.ID, SourceOwnership, fields, searchPQ); hit != nil {
			hits = append(hits, hit)
		} else if pq.typeFilter == EntityTypeExpertise && len(pq.terms) == 0 && pq.domainFilter == "" {
			hits = append(hits, typeFilterHit(EntityTypeExpertise, exp.ID, SourceOwnership))
		}
	}
	return hits, nil
}

func (s *Service) searchRelationships(ctx stdcontext.Context, repositoryID string, pq *parsedQuery) ([]*SearchHit, error) {
	rels, err := s.relationshipReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	var hits []*SearchHit
	for _, rel := range rels {
		fields := []fieldSpec{
			{"id", rel.ID, false},
			{"from_id", rel.FromID, false},
			{"to_id", rel.ToID, false},
			{"type", rel.Type, false},
		}

		if hit := matchEntity(EntityTypeRelationship, rel.ID, SourceGraph, fields, pq); hit != nil {
			hits = append(hits, hit)
		} else if pq.typeFilter == EntityTypeRelationship && len(pq.terms) == 0 {
			hits = append(hits, typeFilterHit(EntityTypeRelationship, rel.ID, SourceGraph))
		}
	}
	return hits, nil
}

func (s *Service) searchDiscovery(ctx stdcontext.Context, repositoryID string, pq *parsedQuery) ([]*SearchHit, error) {
	if len(pq.terms) == 0 && pq.typeFilter == "" {
		return nil, nil
	}

	report, err := s.discoveryService.Discover(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	var hits []*SearchHit
	for _, item := range report.Items {
		if pq.typeFilter != "" && item.EntityType != pq.typeFilter {
			continue
		}

		fields, entityID, err := s.resolveDiscoveryFields(ctx, repositoryID, item)
		if err != nil {
			continue
		}

		if hit := matchEntity(item.EntityType, entityID, SourceDiscovery, fields, pq); hit != nil {
			hits = append(hits, hit)
		}
	}
	return hits, nil
}

func (s *Service) resolveDiscoveryFields(ctx stdcontext.Context, repositoryID string, item *discovery.DiscoveryItem) ([]fieldSpec, string, error) {
	switch item.EntityType {
	case discovery.EntityTypeDecision:
		d, err := s.decisionReader.GetByID(ctx, item.EntityID)
		if err != nil {
			return nil, "", err
		}
		return []fieldSpec{
			{"id", d.ID, false},
			{"title", d.Title, false},
			{"status", d.Status, true},
		}, d.ID, nil

	case discovery.EntityTypeFact:
		f, err := s.factReader.GetByID(ctx, item.EntityID)
		if err != nil {
			return nil, "", err
		}
		return []fieldSpec{
			{"subject", f.Subject, false},
			{"predicate", f.Predicate, false},
			{"object", f.Object, false},
			{"id", f.ID, true},
		}, f.ID, nil

	case discovery.EntityTypeEvent:
		ev, err := s.eventReader.GetByID(ctx, item.EntityID)
		if err != nil {
			return nil, "", err
		}
		return []fieldSpec{
			{"id", ev.ID, false},
			{"description", ev.Description, false},
			{"title", ev.Title, true},
		}, ev.ID, nil

	case discovery.EntityTypeContributor:
		c, err := s.contributorReader.GetByID(ctx, repositoryID, item.EntityID)
		if err != nil {
			return nil, "", err
		}
		return []fieldSpec{
			{"name", c.Name, false},
			{"email", c.Email, false},
			{"id", c.ID, true},
		}, c.ID, nil

	default:
		return nil, "", fmt.Errorf("unsupported discovery entity type: %s", item.EntityType)
	}
}

func (s *Service) buildContributorMap(ctx stdcontext.Context, repositoryID string) (map[string]*models.Contributor, error) {
	contribs, err := s.contributorReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	m := make(map[string]*models.Contributor, len(contribs))
	for _, c := range contribs {
		m[c.ID] = c
	}
	return m, nil
}

func matchEntity(entityType, entityID, source string, fields []fieldSpec, pq *parsedQuery) *SearchHit {
	if len(pq.terms) == 0 {
		return nil
	}

	bestScore := -1
	var bestEvidence matchEvidence

	for _, term := range pq.terms {
		for _, f := range fields {
			score, matchType, ok := scoreField(f.value, term, f.weak)
			if !ok {
				continue
			}
			if score > bestScore {
				bestScore = score
				bestEvidence = matchEvidence{MatchType: matchType, Field: f.name}
			}
		}
	}

	if bestScore < 0 {
		return nil
	}

	evidenceJSON, err := json.Marshal(bestEvidence)
	if err != nil {
		return nil
	}

	return &SearchHit{
		EntityType:   entityType,
		EntityID:     entityID,
		Source:       source,
		MatchScore:   bestScore,
		EvidenceJSON: string(evidenceJSON),
	}
}

func typeFilterHit(entityType, entityID, source string) *SearchHit {
	evidenceJSON, _ := json.Marshal(matchEvidence{MatchType: "weak", Field: "type_filter"})
	return &SearchHit{
		EntityType:   entityType,
		EntityID:     entityID,
		Source:       source,
		MatchScore:   ScoreWeak,
		EvidenceJSON: string(evidenceJSON),
	}
}

func scoreField(field, term string, weak bool) (int, string, bool) {
	field = strings.TrimSpace(field)
	term = strings.TrimSpace(term)
	if field == "" || term == "" {
		return 0, "", false
	}

	isPrefixQuery := strings.HasSuffix(term, "*")
	cleanTerm := term
	if isPrefixQuery {
		cleanTerm = strings.TrimSuffix(term, "*")
	}
	if cleanTerm == "" {
		return 0, "", false
	}

	fieldLower := strings.ToLower(field)
	termLower := strings.ToLower(cleanTerm)

	if fieldLower == termLower {
		return ScoreExact, "exact", true
	}
	if isPrefixQuery || strings.HasPrefix(fieldLower, termLower) {
		return ScorePrefix, "prefix", true
	}
	if strings.Contains(fieldLower, termLower) {
		if weak {
			return ScoreWeak, "weak", true
		}
		return ScorePartial, "partial", true
	}
	return 0, "", false
}

func domainMatches(domain, filter string) bool {
	domainLower := strings.ToLower(domain)
	filterLower := strings.ToLower(filter)
	return domainLower == filterLower ||
		strings.HasPrefix(domainLower, filterLower) ||
		strings.Contains(domainLower, filterLower)
}

func deduplicateHits(hits []*SearchHit) []*SearchHit {
	type key struct {
		entityType string
		entityID   string
	}

	best := make(map[key]*SearchHit)
	for _, hit := range hits {
		k := key{hit.EntityType, hit.EntityID}
		existing, ok := best[k]
		if !ok {
			best[k] = hit
			continue
		}
		if hit.MatchScore > existing.MatchScore {
			best[k] = hit
			continue
		}
		if hit.MatchScore == existing.MatchScore && sourcePrecedence[hit.Source] > sourcePrecedence[existing.Source] {
			best[k] = hit
		}
	}

	result := make([]*SearchHit, 0, len(best))
	for _, hit := range best {
		result = append(result, hit)
	}
	return result
}

func sortHits(hits []*SearchHit) {
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].MatchScore != hits[j].MatchScore {
			return hits[i].MatchScore > hits[j].MatchScore
		}
		if hits[i].EntityType != hits[j].EntityType {
			return hits[i].EntityType < hits[j].EntityType
		}
		return hits[i].EntityID < hits[j].EntityID
	})
}
