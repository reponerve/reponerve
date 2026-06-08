package relationships

import (
	"encoding/json"

	"github.com/reponerve/reponerve/internal/graph/model"
)

// DerivedRelationship represents a derived connection computed from repository memory.
// Unlike stored relationships, derived relationships are conclusions and not persisted directly.
type DerivedRelationship struct {
	Edge        *model.GraphEdge `json:"edge"`
	Evidence    json.RawMessage  `json:"evidence"`
	Explanation string           `json:"explanation"`
}
