package traversal

import (
	"context"
	"fmt"
	"sort"

	"github.com/reponerve/reponerve/internal/graph/model"
)

// BudgetTraversalOptions configures token-budget graph exploration.
type BudgetTraversalOptions struct {
	TraversalOptions
	TokenBudget int `json:"token_budget"`
}

// BudgetTraversalResult is a bounded subgraph reachable from a start node.
type BudgetTraversalResult struct {
	StartNodeID string             `json:"start_node_id"`
	Nodes       []*model.GraphNode `json:"nodes"`
	Edges       []*model.GraphEdge `json:"edges"`
	TokensUsed  int                `json:"tokens_used"`
}

// TraverseWithBudget performs BFS from startNodeID until the token budget is exhausted.
func (e *Engine) TraverseWithBudget(
	ctx context.Context,
	repositoryID string,
	startNodeID string,
	options BudgetTraversalOptions,
) (*BudgetTraversalResult, error) {
	if startNodeID == "" {
		return nil, fmt.Errorf("start node ID is required")
	}
	if options.TokenBudget <= 0 {
		return nil, fmt.Errorf("token budget must be positive")
	}

	opts := options.TraversalOptions
	if !opts.IncludeStored && !opts.IncludeDerived {
		opts.IncludeStored = true
		opts.IncludeDerived = true
	}
	snapshot, err := e.LoadGraphSnapshot(ctx, repositoryID, opts)
	if err != nil {
		return nil, err
	}

	nodeByID := make(map[string]*model.GraphNode, len(snapshot.Nodes))
	for _, n := range snapshot.Nodes {
		nodeByID[n.ID] = n
	}
	if _, ok := nodeByID[startNodeID]; !ok {
		return &BudgetTraversalResult{StartNodeID: startNodeID}, nil
	}

	adj := make(map[string][]*model.GraphEdge)
	for _, edge := range snapshot.Edges {
		adj[edge.FromNodeID] = append(adj[edge.FromNodeID], edge)
		adj[edge.ToNodeID] = append(adj[edge.ToNodeID], edge)
	}
	for id := range adj {
		sort.Slice(adj[id], func(i, j int) bool {
			return adj[id][i].ID < adj[id][j].ID
		})
	}

	visited := map[string]bool{startNodeID: true}
	queue := []string{startNodeID}
	selectedNodes := []*model.GraphNode{nodeByID[startNodeID]}
	selectedEdges := make([]*model.GraphEdge, 0)
	tokensUsed := estimateNodeTokens(nodeByID[startNodeID])

	for len(queue) > 0 && tokensUsed < options.TokenBudget {
		cur := queue[0]
		queue = queue[1:]
		for _, edge := range adj[cur] {
			other := edge.ToNodeID
			if other == cur {
				other = edge.FromNodeID
			}
			edgeCost := estimateEdgeTokens(edge)
			nodeCost := 0
			if !visited[other] {
				nodeCost = estimateNodeTokens(nodeByID[other])
			}
			if tokensUsed+edgeCost+nodeCost > options.TokenBudget {
				continue
			}
			tokensUsed += edgeCost
			selectedEdges = append(selectedEdges, edge)
			if !visited[other] {
				visited[other] = true
				tokensUsed += nodeCost
				selectedNodes = append(selectedNodes, nodeByID[other])
				queue = append(queue, other)
			}
		}
	}

	sort.Slice(selectedNodes, func(i, j int) bool { return selectedNodes[i].ID < selectedNodes[j].ID })
	sort.Slice(selectedEdges, func(i, j int) bool { return selectedEdges[i].ID < selectedEdges[j].ID })

	return &BudgetTraversalResult{
		StartNodeID: startNodeID,
		Nodes:       selectedNodes,
		Edges:       selectedEdges,
		TokensUsed:  tokensUsed,
	}, nil
}

func estimateNodeTokens(node *model.GraphNode) int {
	if node == nil {
		return 1
	}
	text := string(node.NodeType) + node.EntityID
	return (len(text) + 3) / 4
}

func estimateEdgeTokens(edge *model.GraphEdge) int {
	if edge == nil {
		return 1
	}
	text := string(edge.EdgeType) + edge.FromNodeID + edge.ToNodeID
	return (len(text) + 3) / 4
}
