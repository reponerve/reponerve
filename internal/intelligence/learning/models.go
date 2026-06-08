package learning

import (
	"encoding/json"
	"fmt"
)

// Supported entity types.
const (
	EntityTypeDecision    = "DECISION"
	EntityTypeFact        = "FACT"
	EntityTypeEvent       = "EVENT"
	EntityTypeContributor = "CONTRIBUTOR"
)

// LearningStep represents an ordered step in a learning path.
type LearningStep struct {
	EntityType   string `json:"entity_type"`
	EntityID     string `json:"entity_id"`
	Position     int    `json:"position"`
	EvidenceJSON string `json:"evidence_json"`
	Explanation  string `json:"explanation"`
}

// LearningPath holds the list of ordered learning steps.
type LearningPath struct {
	Steps []*LearningStep `json:"steps"`
}

// ValidateStep validates that the step properties are correct.
func ValidateStep(step *LearningStep) error {
	if step == nil {
		return fmt.Errorf("step is nil")
	}
	if step.EntityType == "" {
		return fmt.Errorf("missing entity type")
	}
	if step.EntityType != EntityTypeDecision &&
		step.EntityType != EntityTypeFact &&
		step.EntityType != EntityTypeEvent &&
		step.EntityType != EntityTypeContributor {
		return fmt.Errorf("invalid entity type: %q", step.EntityType)
	}
	if step.EntityID == "" {
		return fmt.Errorf("missing entity ID")
	}
	if step.Position <= 0 {
		return fmt.Errorf("invalid position: %d (must be > 0)", step.Position)
	}
	if step.EvidenceJSON == "" {
		return fmt.Errorf("missing evidence")
	}
	if !json.Valid([]byte(step.EvidenceJSON)) {
		return fmt.Errorf("evidence must be valid JSON")
	}
	if step.Explanation == "" {
		return fmt.Errorf("missing explanation")
	}
	return nil
}
