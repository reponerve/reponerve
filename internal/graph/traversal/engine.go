package traversal

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"reponerve/internal/graph/model"
	"reponerve/internal/graph/relationships"
)

// Engine traces and queries the repository knowledge graph.
type Engine struct {
	relationshipEngine *relationships.Engine
}

// NewEngine constructs a new Traversal Engine.
func NewEngine(relationshipEngine *relationships.Engine) *Engine {
	return &Engine{
		relationshipEngine: relationshipEngine,
	}
}

// TraceGraph returns all reachable graph paths starting at startNodeID.
func (e *Engine) TraceGraph(
	ctx context.Context,
	repositoryID string,
	startNodeID string,
	options TraversalOptions,
) (*TraversalResult, error) {
	return e.traverse(ctx, repositoryID, startNodeID, options, true)
}

// FindDependencies returns outbound dependency paths starting at nodeID.
func (e *Engine) FindDependencies(
	ctx context.Context,
	repositoryID string,
	nodeID string,
	options TraversalOptions,
) (*TraversalResult, error) {
	return e.traverse(ctx, repositoryID, nodeID, options, true)
}

// FindDependents returns inbound dependency paths ending at nodeID.
func (e *Engine) FindDependents(
	ctx context.Context,
	repositoryID string,
	nodeID string,
	options TraversalOptions,
) (*TraversalResult, error) {
	return e.traverse(ctx, repositoryID, nodeID, options, false)
}

// traverse performs a deterministic BFS graph traversal either outbound or inbound.
func (e *Engine) traverse(
	ctx context.Context,
	repositoryID string,
	startNodeID string,
	options TraversalOptions,
	outbound bool,
) (*TraversalResult, error) {
	maxDepth := options.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 10
	}

	// 1. Load all nodes
	nodesByID, entityToNodeID, err := e.loadNodes(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to load graph nodes: %w", err)
	}

	// Verify start node exists
	startNode, exists := nodesByID[startNodeID]
	if !exists {
		return &TraversalResult{Paths: []*TraversalPath{}}, nil
	}

	// 2. Load and filter edges
	var edges []*model.GraphEdge

	if options.IncludeStored {
		storedEdges, err := e.loadStoredEdges(ctx, repositoryID, entityToNodeID)
		if err != nil {
			return nil, fmt.Errorf("failed to load stored edges: %w", err)
		}
		edges = append(edges, storedEdges...)
	}

	if options.IncludeDerived {
		derivedRels, err := e.relationshipEngine.Generate(ctx, repositoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate derived relationships: %w", err)
		}
		for _, dr := range derivedRels {
			if dr.Edge != nil {
				edges = append(edges, dr.Edge)
			}
		}
	}

	// 3. Build adjacency lists
	outEdges := make(map[string][]*model.GraphEdge)
	inEdges := make(map[string][]*model.GraphEdge)
	for _, edge := range edges {
		outEdges[edge.FromNodeID] = append(outEdges[edge.FromNodeID], edge)
		inEdges[edge.ToNodeID] = append(inEdges[edge.ToNodeID], edge)
	}

	// 4. Run BFS
	var resultPaths []*TraversalPath
	var queue []*TraversalPath

	if outbound {
		// Initialize queue with 1-hop outbound paths from startNodeID
		for _, edge := range outEdges[startNodeID] {
			toNode, exists := nodesByID[edge.ToNodeID]
			if !exists {
				continue
			}
			path := &TraversalPath{
				Nodes: []*model.GraphNode{startNode, toNode},
				Edges: []*model.GraphEdge{edge},
			}
			if err := validatePath(path); err == nil {
				queue = append(queue, path)
			}
		}

		// BFS Loop
		for len(queue) > 0 {
			P := queue[0]
			queue = queue[1:]

			resultPaths = append(resultPaths, P)

			if len(P.Edges) >= maxDepth {
				continue
			}

			lastNode := P.Nodes[len(P.Nodes)-1]
			for _, edge := range outEdges[lastNode.ID] {
				nextNode, exists := nodesByID[edge.ToNodeID]
				if !exists {
					continue
				}

				// Cycle handling
				hasCycle := false
				for _, n := range P.Nodes {
					if n.ID == nextNode.ID {
						hasCycle = true
						break
					}
				}
				if hasCycle {
					continue
				}

				// Construct expanded path
				nodesCopy := make([]*model.GraphNode, len(P.Nodes)+1)
				copy(nodesCopy, P.Nodes)
				nodesCopy[len(P.Nodes)] = nextNode

				edgesCopy := make([]*model.GraphEdge, len(P.Edges)+1)
				copy(edgesCopy, P.Edges)
				edgesCopy[len(P.Edges)] = edge

				PNew := &TraversalPath{
					Nodes: nodesCopy,
					Edges: edgesCopy,
				}
				if err := validatePath(PNew); err == nil {
					queue = append(queue, PNew)
				}
			}
		}
	} else {
		// Initialize queue with 1-hop inbound paths ending at startNodeID
		for _, edge := range inEdges[startNodeID] {
			fromNode, exists := nodesByID[edge.FromNodeID]
			if !exists {
				continue
			}
			path := &TraversalPath{
				Nodes: []*model.GraphNode{fromNode, startNode},
				Edges: []*model.GraphEdge{edge},
			}
			if err := validatePath(path); err == nil {
				queue = append(queue, path)
			}
		}

		// BFS Loop
		for len(queue) > 0 {
			P := queue[0]
			queue = queue[1:]

			resultPaths = append(resultPaths, P)

			if len(P.Edges) >= maxDepth {
				continue
			}

			firstNode := P.Nodes[0]
			for _, edge := range inEdges[firstNode.ID] {
				prevNode, exists := nodesByID[edge.FromNodeID]
				if !exists {
					continue
				}

				// Cycle handling
				hasCycle := false
				for _, n := range P.Nodes {
					if n.ID == prevNode.ID {
						hasCycle = true
						break
					}
				}
				if hasCycle {
					continue
				}

				// Construct prepended path
				nodesCopy := make([]*model.GraphNode, len(P.Nodes)+1)
				nodesCopy[0] = prevNode
				copy(nodesCopy[1:], P.Nodes)

				edgesCopy := make([]*model.GraphEdge, len(P.Edges)+1)
				edgesCopy[0] = edge
				copy(edgesCopy[1:], P.Edges)

				PNew := &TraversalPath{
					Nodes: nodesCopy,
					Edges: edgesCopy,
				}
				if err := validatePath(PNew); err == nil {
					queue = append(queue, PNew)
				}
			}
		}
	}

	// 5. Deterministic sorting
	sort.Slice(resultPaths, func(i, j int) bool {
		lenI := len(resultPaths[i].Edges)
		lenJ := len(resultPaths[j].Edges)
		if lenI != lenJ {
			return lenI < lenJ
		}
		startI := resultPaths[i].Nodes[0].ID
		startJ := resultPaths[j].Nodes[0].ID
		if startI != startJ {
			return startI < startJ
		}
		endI := resultPaths[i].Nodes[len(resultPaths[i].Nodes)-1].ID
		endJ := resultPaths[j].Nodes[len(resultPaths[j].Nodes)-1].ID
		return endI < endJ
	})

	return &TraversalResult{Paths: resultPaths}, nil
}

func (e *Engine) loadNodes(ctx context.Context, repoID string) (map[string]*model.GraphNode, map[string]string, error) {
	nodesByID := make(map[string]*model.GraphNode)
	entityToNodeID := make(map[string]string)

	// Decisions
	decs, err := e.relationshipEngine.DecisionReader().ListByRepository(ctx, repoID)
	if err != nil {
		return nil, nil, err
	}
	for _, d := range decs {
		nid := model.NodeID(repoID, model.NodeTypeDecision, d.ID)
		nodesByID[nid] = model.NewNode(repoID, model.NodeTypeDecision, d.ID)
		entityToNodeID[d.ID] = nid
	}

	// Intents
	intents, err := e.relationshipEngine.IntentReader().ListByRepository(ctx, repoID)
	if err != nil {
		return nil, nil, err
	}
	for _, i := range intents {
		nid := model.NodeID(repoID, model.NodeTypeIntent, i.ID)
		nodesByID[nid] = model.NewNode(repoID, model.NodeTypeIntent, i.ID)
		entityToNodeID[i.ID] = nid
	}

	// Facts
	facts, err := e.relationshipEngine.FactReader().ListByRepository(ctx, repoID)
	if err != nil {
		return nil, nil, err
	}
	for _, f := range facts {
		nid := model.NodeID(repoID, model.NodeTypeFact, f.ID)
		nodesByID[nid] = model.NewNode(repoID, model.NodeTypeFact, f.ID)
		entityToNodeID[f.ID] = nid
	}

	// Events
	events, err := e.relationshipEngine.EventReader().ListByRepository(ctx, repoID)
	if err != nil {
		return nil, nil, err
	}
	for _, ev := range events {
		nid := model.NodeID(repoID, model.NodeTypeEvent, ev.ID)
		nodesByID[nid] = model.NewNode(repoID, model.NodeTypeEvent, ev.ID)
		entityToNodeID[ev.ID] = nid
	}

	// Contributors
	contribs, err := e.relationshipEngine.ContributorReader().ListByRepository(ctx, repoID)
	if err != nil {
		return nil, nil, err
	}
	for _, c := range contribs {
		nid := model.NodeID(repoID, model.NodeTypeContributor, c.ID)
		nodesByID[nid] = model.NewNode(repoID, model.NodeTypeContributor, c.ID)
		entityToNodeID[c.ID] = nid
	}

	// Expertise
	expertise, err := e.relationshipEngine.ExpertiseReader().ListByRepository(ctx, repoID)
	if err != nil {
		return nil, nil, err
	}
	for _, exp := range expertise {
		nid := model.NodeID(repoID, model.NodeTypeExpertise, exp.ID)
		nodesByID[nid] = model.NewNode(repoID, model.NodeTypeExpertise, exp.ID)
		entityToNodeID[exp.ID] = nid
	}

	return nodesByID, entityToNodeID, nil
}

func (e *Engine) loadStoredEdges(ctx context.Context, repoID string, entityToNodeID map[string]string) ([]*model.GraphEdge, error) {
	var edges []*model.GraphEdge

	rels, err := e.relationshipEngine.RelationshipReader().ListByRepository(ctx, repoID)
	if err != nil {
		return nil, err
	}

	for _, rel := range rels {
		fromNodeID, existsFrom := entityToNodeID[rel.FromID]
		toNodeID, existsTo := entityToNodeID[rel.ToID]
		if existsFrom && existsTo {
			evidence := map[string]string{
				"relationship_id":   rel.ID,
				"relationship_type": rel.Type,
			}
			evBytes, _ := json.Marshal(evidence)

			edge := model.NewEdge(
				repoID,
				fromNodeID,
				toNodeID,
				model.EdgeType(rel.Type),
				model.CategoryStored,
				string(evBytes),
			)
			edges = append(edges, edge)
		}
	}

	// Load contributor-expertise stored relationships
	exps, err := e.relationshipEngine.ExpertiseReader().ListByRepository(ctx, repoID)
	if err != nil {
		return nil, err
	}
	for _, exp := range exps {
		if exp.ContributorID != "" {
			fromNodeID, existsFrom := entityToNodeID[exp.ID]
			toNodeID, existsTo := entityToNodeID[exp.ContributorID]
			if existsFrom && existsTo {
				evidence := map[string]string{
					"expertise_id":   exp.ID,
					"contributor_id": exp.ContributorID,
					"domain":         exp.Domain,
				}
				evBytes, _ := json.Marshal(evidence)

				edge := model.NewEdge(
					repoID,
					fromNodeID,
					toNodeID,
					"CONTRIBUTOR_EXPERT_IN_DOMAIN",
					model.CategoryStored,
					string(evBytes),
				)
				edges = append(edges, edge)
			}
		}
	}

	return edges, nil
}

func validatePath(path *TraversalPath) error {
	if path == nil {
		return fmt.Errorf("path is nil")
	}
	if len(path.Nodes) == 0 {
		return fmt.Errorf("path has no nodes")
	}
	if len(path.Edges) == 0 {
		return fmt.Errorf("path has no edges")
	}
	if len(path.Edges) != len(path.Nodes)-1 {
		return fmt.Errorf("invalid path: edge count %d != node count %d minus 1", len(path.Edges), len(path.Nodes))
	}
	for i, node := range path.Nodes {
		if node == nil {
			return fmt.Errorf("nil node at index %d", i)
		}
	}
	for i, edge := range path.Edges {
		if edge == nil {
			return fmt.Errorf("nil edge at index %d", i)
		}
	}
	return nil
}
