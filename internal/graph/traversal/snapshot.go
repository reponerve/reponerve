package traversal

import (
	"context"
	"fmt"
	"sort"

	"github.com/reponerve/reponerve/internal/graph/model"
)

// GraphSnapshot is a deterministic node/edge view of repository knowledge graph.
type GraphSnapshot struct {
	RepositoryID string             `json:"repository_id"`
	Nodes        []*model.GraphNode `json:"nodes"`
	Edges        []*model.GraphEdge `json:"edges"`
}

// LoadGraphSnapshot loads graph nodes and edges for exploration and community detection.
func (e *Engine) LoadGraphSnapshot(ctx context.Context, repositoryID string, options TraversalOptions) (*GraphSnapshot, error) {
	nodesByID, entityToNodeID, err := e.loadNodes(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("load nodes: %w", err)
	}

	var edges []*model.GraphEdge
	if options.IncludeStored {
		stored, err := e.loadStoredEdges(ctx, repositoryID, entityToNodeID)
		if err != nil {
			return nil, fmt.Errorf("load stored edges: %w", err)
		}
		edges = append(edges, stored...)
	}
	if options.IncludeDerived {
		derived, err := e.relationshipEngine.Generate(ctx, repositoryID)
		if err != nil {
			return nil, fmt.Errorf("generate derived edges: %w", err)
		}
		for _, dr := range derived {
			if dr != nil && dr.Edge != nil {
				edges = append(edges, dr.Edge)
			}
		}
	}

	nodes := make([]*model.GraphNode, 0, len(nodesByID))
	for _, n := range nodesByID {
		nodes = append(nodes, n)
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID < nodes[j].ID
	})
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].ID < edges[j].ID
	})

	return &GraphSnapshot{
		RepositoryID: repositoryID,
		Nodes:        nodes,
		Edges:        edges,
	}, nil
}
