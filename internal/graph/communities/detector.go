package communities

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/reponerve/reponerve/internal/graph/model"
)

// Detect finds connected components in an undirected view of the graph.
func Detect(repositoryID string, nodes []*model.GraphNode, edges []*model.GraphEdge) *DetectionResult {
	adj := buildAdjacency(nodes, edges)
	visited := make(map[string]bool, len(nodes))
	communities := make([]Community, 0)

	nodeIDs := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if n != nil {
			nodeIDs = append(nodeIDs, n.ID)
		}
	}
	sort.Strings(nodeIDs)

	for _, start := range nodeIDs {
		if visited[start] {
			continue
		}
		component := bfsComponent(start, adj, visited)
		sort.Strings(component)
		communities = append(communities, Community{
			ID:      communityID(repositoryID, component),
			NodeIDs: component,
			Size:    len(component),
		})
	}

	sort.Slice(communities, func(i, j int) bool {
		if communities[i].Size != communities[j].Size {
			return communities[i].Size > communities[j].Size
		}
		return communities[i].ID < communities[j].ID
	})

	return &DetectionResult{
		RepositoryID: repositoryID,
		Communities:  communities,
	}
}

func buildAdjacency(nodes []*model.GraphNode, edges []*model.GraphEdge) map[string][]string {
	adj := make(map[string][]string, len(nodes))
	for _, n := range nodes {
		if n != nil {
			adj[n.ID] = nil
		}
	}
	for _, edge := range edges {
		if edge == nil {
			continue
		}
		adj[edge.FromNodeID] = append(adj[edge.FromNodeID], edge.ToNodeID)
		adj[edge.ToNodeID] = append(adj[edge.ToNodeID], edge.FromNodeID)
	}
	for id, neighbors := range adj {
		sort.Strings(neighbors)
		adj[id] = neighbors
	}
	return adj
}

func bfsComponent(start string, adj map[string][]string, visited map[string]bool) []string {
	queue := []string{start}
	visited[start] = true
	component := make([]string, 0)

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		component = append(component, cur)
		for _, next := range adj[cur] {
			if visited[next] {
				continue
			}
			visited[next] = true
			queue = append(queue, next)
		}
	}
	return component
}

func communityID(repositoryID string, nodeIDs []string) string {
	h := sha256.New()
	_, _ = h.Write([]byte(repositoryID))
	for _, id := range nodeIDs {
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(id))
	}
	return "com_" + hex.EncodeToString(h.Sum(nil)[:8])
}

// NodeCommunityMap returns node ID to community ID.
func NodeCommunityMap(result *DetectionResult) map[string]string {
	out := make(map[string]string)
	if result == nil {
		return out
	}
	for _, c := range result.Communities {
		for _, nodeID := range c.NodeIDs {
			out[nodeID] = c.ID
		}
	}
	return out
}

// Validate ensures communities are well-formed.
func Validate(result *DetectionResult) error {
	if result == nil {
		return fmt.Errorf("result is nil")
	}
	if result.RepositoryID == "" {
		return fmt.Errorf("missing repository ID")
	}
	seen := make(map[string]string)
	for _, c := range result.Communities {
		if c.ID == "" {
			return fmt.Errorf("community missing ID")
		}
		for _, nodeID := range c.NodeIDs {
			if prev, ok := seen[nodeID]; ok {
				return fmt.Errorf("node %q assigned to communities %q and %q", nodeID, prev, c.ID)
			}
			seen[nodeID] = c.ID
		}
	}
	return nil
}
