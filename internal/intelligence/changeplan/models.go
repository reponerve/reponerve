package changeplan

import (
	"encoding/json"
	"fmt"
)

// ChangePlanItem represents a specific repository entity that should be reviewed prior to a change.
type ChangePlanItem struct {
	EntityType   string `json:"entity_type"`
	EntityID     string `json:"entity_id"`
	Priority     int    `json:"priority"`
	EvidenceJSON string `json:"evidence_json"`
	Explanation  string `json:"explanation"`
}

// ChangePlan holds the list of prioritized change plan items.
type ChangePlan struct {
	Items []*ChangePlanItem `json:"items"`
}

// ValidateItem validates that the item properties are populated and correct.
func ValidateItem(item *ChangePlanItem) error {
	if item == nil {
		return fmt.Errorf("item is nil")
	}
	if item.EntityType == "" {
		return fmt.Errorf("missing entity type")
	}
	if item.EntityType != "DECISION" &&
		item.EntityType != "FACT" &&
		item.EntityType != "EVENT" &&
		item.EntityType != "CONTRIBUTOR" {
		return fmt.Errorf("invalid entity type: %q", item.EntityType)
	}
	if item.EntityID == "" {
		return fmt.Errorf("missing entity ID")
	}
	if item.Priority <= 0 {
		return fmt.Errorf("invalid priority: %d (must be > 0)", item.Priority)
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
