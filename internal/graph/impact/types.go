package impact

import (
	"github.com/reponerve/reponerve/internal/graph/traversal"
)

// ImpactPath represents a single traversal path demonstrating how a change propagates to a target node, accompanied by a reason.
type ImpactPath struct {
	Path   *traversal.TraversalPath `json:"path"`
	Reason string                   `json:"reason"`
}

// ImpactReport collects all impact paths discovered for a starting node.
type ImpactReport struct {
	ImpactPaths []*ImpactPath `json:"impact_paths"`
}
