package agentsearch

import (
	"encoding/json"
	"fmt"
)

// Supported source values for SearchHit.
const (
	SourceMemory    = "memory"
	SourceOwnership = "ownership"
	SourceGraph     = "graph"
	SourceDiscovery = "discovery"
)

// Entity type constants for search hits.
const (
	EntityTypeDecision    = "DECISION"
	EntityTypeFact        = "FACT"
	EntityTypeEvent       = "EVENT"
	EntityTypeContributor = "CONTRIBUTOR"
	EntityTypeExpertise   = "EXPERTISE"
	EntityTypeRelationship = "RELATIONSHIP"
)

// Match score constants represent retrieval relevance only.
const (
	ScoreExact   = 100
	ScorePrefix  = 75
	ScorePartial = 50
	ScoreWeak    = 25
)

var validSources = map[string]bool{
	SourceMemory:    true,
	SourceOwnership: true,
	SourceGraph:     true,
	SourceDiscovery: true,
}

var validEntityTypes = map[string]bool{
	EntityTypeDecision:     true,
	EntityTypeFact:         true,
	EntityTypeEvent:        true,
	EntityTypeContributor:  true,
	EntityTypeExpertise:    true,
	EntityTypeRelationship: true,
}

// SearchHit represents one deterministic repository knowledge retrieval match.
type SearchHit struct {
	EntityType   string `json:"entity_type"`
	EntityID     string `json:"entity_id"`
	Source       string `json:"source"`
	MatchScore   int    `json:"match_score"`
	EvidenceJSON string `json:"evidence_json"`
}

// SearchResult collects all hits for a single search query.
type SearchResult struct {
	RepositoryID string       `json:"repository_id"`
	Query        string       `json:"query"`
	Hits         []*SearchHit `json:"hits"`
}

// ValidateHit ensures a single hit is structurally valid.
func ValidateHit(hit *SearchHit) error {
	if hit == nil {
		return fmt.Errorf("hit is nil")
	}
	if hit.EntityType == "" {
		return fmt.Errorf("missing entity type")
	}
	if !validEntityTypes[hit.EntityType] {
		return fmt.Errorf("invalid entity type: %q", hit.EntityType)
	}
	if hit.EntityID == "" {
		return fmt.Errorf("missing entity ID")
	}
	if hit.Source == "" {
		return fmt.Errorf("missing source")
	}
	if !validSources[hit.Source] {
		return fmt.Errorf("unsupported source %q (must be one of: memory, ownership, graph, discovery)", hit.Source)
	}
	if hit.MatchScore < 0 {
		return fmt.Errorf("invalid match score: %d (must be non-negative)", hit.MatchScore)
	}
	if hit.EvidenceJSON == "" {
		return fmt.Errorf("missing evidence")
	}
	if !json.Valid([]byte(hit.EvidenceJSON)) {
		return fmt.Errorf("evidence must be valid JSON")
	}
	return nil
}

// ValidateResult ensures a search result is structurally valid.
func ValidateResult(result *SearchResult) error {
	if result == nil {
		return fmt.Errorf("result is nil")
	}
	if result.RepositoryID == "" {
		return fmt.Errorf("missing repository ID")
	}
	if result.Query == "" {
		return fmt.Errorf("missing query")
	}
	if result.Hits == nil {
		return fmt.Errorf("hits is nil")
	}
	for i, hit := range result.Hits {
		if err := ValidateHit(hit); err != nil {
			return fmt.Errorf("hit %d: %w", i, err)
		}
	}
	return nil
}
