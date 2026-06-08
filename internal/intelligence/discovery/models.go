package discovery

import (
	"encoding/json"
	"fmt"
)

// Supported entity types for v1.
const (
	EntityTypeDecision    = "DECISION"
	EntityTypeFact        = "FACT"
	EntityTypeEvent       = "EVENT"
	EntityTypeContributor = "CONTRIBUTOR"
)

// DiscoveryItem represents a recommended knowledge entity surfaced by the engine.
type DiscoveryItem struct {
	EntityType   string  `json:"entity_type"`
	EntityID     string  `json:"entity_id"`
	Score        float64 `json:"score"`
	EvidenceJSON string  `json:"evidence_json"`
	Explanation  string  `json:"explanation"`
}

// KnowledgeDiscoveryReport collects all discovery recommendations.
type KnowledgeDiscoveryReport struct {
	Items []*DiscoveryItem `json:"items"`
}

// ValidateItem validates that the item properties are populated and correct.
func ValidateItem(item *DiscoveryItem) error {
	if item == nil {
		return fmt.Errorf("item is nil")
	}
	if item.EntityType == "" {
		return fmt.Errorf("missing entity type")
	}
	if item.EntityType != EntityTypeDecision &&
		item.EntityType != EntityTypeFact &&
		item.EntityType != EntityTypeEvent &&
		item.EntityType != EntityTypeContributor {
		return fmt.Errorf("invalid entity type: %q", item.EntityType)
	}
	if item.EntityID == "" {
		return fmt.Errorf("missing entity ID")
	}
	if item.Score < 0 {
		return fmt.Errorf("invalid score: %f (must be non-negative)", item.Score)
	}
	if item.EvidenceJSON == "" {
		return fmt.Errorf("missing evidence")
	}
	if !json.Valid([]byte(item.EvidenceJSON)) {
		return fmt.Errorf("evidence must be valid JSON")
	}
	if item.Explanation == "" {
		return fmt.Errorf("missing explanation")
	}
	return nil
}
