package code

import (
	"sort"

	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

const (
	relCalls      = "CALLS"
	relReferences = "REFERENCES"
	relDependsOn  = "DEPENDS_ON"
	relImports    = "IMPORTS"
)

var dependencyRelationshipTypes = map[string]bool{
	relCalls:      true,
	relReferences: true,
	relDependsOn:  true,
	relImports:    true,
}

// BuildCallGraphFromRelationships builds a deterministic outbound CALLS graph.
func BuildCallGraphFromRelationships(rootEntityID string, allRelationships []*codemodels.CodeRelationship, maxDepth int) *codemodels.CallGraph {
	if maxDepth <= 0 {
		maxDepth = 10
	}

	outbound := make(map[string][]*codemodels.CodeRelationship)
	for _, rel := range allRelationships {
		if rel.RelationshipType != relCalls {
			continue
		}
		outbound[rel.FromEntityID] = append(outbound[rel.FromEntityID], rel)
	}
	for fromID := range outbound {
		sort.Slice(outbound[fromID], func(i, j int) bool {
			return outbound[fromID][i].ToEntityID < outbound[fromID][j].ToEntityID
		})
	}

	graph := &codemodels.CallGraph{RootEntityID: rootEntityID}
	visited := map[string]struct{}{rootEntityID: {}}
	queue := []struct {
		entityID string
		depth    int
	}{{rootEntityID, 0}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current.depth >= maxDepth {
			continue
		}
		for _, rel := range outbound[current.entityID] {
			graph.Edges = append(graph.Edges, &codemodels.CallGraphEdge{
				FromEntityID: rel.FromEntityID,
				ToEntityID:   rel.ToEntityID,
				Relationship: rel,
			})
			if _, seen := visited[rel.ToEntityID]; seen {
				continue
			}
			visited[rel.ToEntityID] = struct{}{}
			queue = append(queue, struct {
				entityID string
				depth    int
			}{rel.ToEntityID, current.depth + 1})
		}
	}

	sort.Slice(graph.Edges, func(i, j int) bool {
		if graph.Edges[i].FromEntityID != graph.Edges[j].FromEntityID {
			return graph.Edges[i].FromEntityID < graph.Edges[j].FromEntityID
		}
		return graph.Edges[i].ToEntityID < graph.Edges[j].ToEntityID
	})

	return graph
}

// CollectSymbolDependencies returns outbound dependency relationships for a symbol.
func CollectSymbolDependencies(rootEntityID string, outbound []*codemodels.CodeRelationship) []*codemodels.CodeRelationship {
	var deps []*codemodels.CodeRelationship
	for _, rel := range outbound {
		if dependencyRelationshipTypes[rel.RelationshipType] {
			deps = append(deps, rel)
		}
	}
	sort.Slice(deps, func(i, j int) bool {
		if deps[i].RelationshipType != deps[j].RelationshipType {
			return deps[i].RelationshipType < deps[j].RelationshipType
		}
		return deps[i].ToEntityID < deps[j].ToEntityID
	})
	return deps
}
