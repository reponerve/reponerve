package traversal

import (
	"github.com/reponerve/reponerve/internal/graph/model"
)

// TraversalOptions configures graph traversal limits and filters.
type TraversalOptions struct {
	MaxDepth       int  `json:"max_depth"`
	IncludeStored  bool `json:"include_stored"`
	IncludeDerived bool `json:"include_derived"`
}

// TraversalPath represents a sequence of graph nodes and connecting edges.
type TraversalPath struct {
	Nodes []*model.GraphNode `json:"nodes"`
	Edges []*model.GraphEdge `json:"edges"`
}

// TraversalResult holds all matched traversal paths.
type TraversalResult struct {
	Paths []*TraversalPath `json:"paths"`
}
