package exploreui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/reponerve/reponerve/internal/graph/communities"
	graphdiscovery "github.com/reponerve/reponerve/internal/graph/discovery"
	"github.com/reponerve/reponerve/internal/graph/model"
	"github.com/reponerve/reponerve/internal/graph/relationships"
	"github.com/reponerve/reponerve/internal/graph/traversal"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/scanner/repository"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

const MaxDisplayNodes = 200

// Payload is the explore UI view model.
type Payload struct {
	RepositoryID string
	Stats        Stats
	Nodes        []NodeView
	Edges        []EdgeView
	TotalNodes   int
	TotalEdges   int
}

// Stats summarizes graph analysis.
type Stats struct {
	Communities int
	GodNodes    int
	Surprises   int
}

// NodeView is a capped graph node for display.
type NodeView struct {
	ID       string
	Type     string
	EntityID string
	Degree   int
}

// EdgeView is an edge between displayed nodes.
type EdgeView struct {
	ID       string
	From     string
	To       string
	Type     string
	Category string
	Evidence string
}

// NodeDetail is the htmx evidence panel model.
type NodeDetail struct {
	Node     NodeView
	Incoming []EdgeView
	Outgoing []EdgeView
	Hints    []string
}

// Loader loads graph payloads from workspace memory.
type Loader struct {
	DB           *sqlite.Database
	RepoPath     string
	Discovery    repository.Discovery
	Traversal    *traversal.Engine
}

// Load reads the repository graph and builds a capped UI payload.
func (l *Loader) Load(ctx context.Context) (*Payload, error) {
	if l.Discovery == nil {
		l.Discovery = repository.NewGitDiscovery()
	}
	repo, err := l.Discovery.Discover(ctx, l.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("discover repository: %w", err)
	}
	if l.Traversal == nil {
		decisionReader := storage.NewSQLiteDecisionReader(l.DB)
		intentReader := storage.NewSQLiteIntentReader(l.DB)
		factReader := storage.NewSQLiteFactReader(l.DB)
		eventReader := storage.NewSQLiteEventReader(l.DB)
		relationshipReader := storage.NewSQLiteRelationshipReader(l.DB)
		contribReader := storage.NewSQLiteContributorReader(l.DB)
		expertiseReader := storage.NewSQLiteExpertiseReader(l.DB)
		sourceReader := storage.NewSQLiteSourceReader(l.DB)
		relEngine := relationships.NewEngine(
			decisionReader, intentReader, factReader, eventReader,
			relationshipReader, contribReader, expertiseReader, sourceReader,
		)
		l.Traversal = traversal.NewEngine(relEngine)
	}

	snapshot, err := l.Traversal.LoadGraphSnapshot(ctx, repo.ID, traversal.TraversalOptions{
		IncludeStored:  true,
		IncludeDerived: true,
	})
	if err != nil {
		return nil, fmt.Errorf("load graph: %w", err)
	}

	communityResult := communities.Detect(repo.ID, snapshot.Nodes, snapshot.Edges)
	report, err := graphdiscovery.Analyze(repo.ID, snapshot.Nodes, snapshot.Edges, communityResult)
	if err != nil {
		return nil, fmt.Errorf("analyze graph: %w", err)
	}

	degree := degreeByNode(snapshot.Edges)
	nodes := make([]NodeView, 0, len(snapshot.Nodes))
	for _, n := range snapshot.Nodes {
		nodes = append(nodes, NodeView{
			ID:       n.ID,
			Type:     string(n.NodeType),
			EntityID: n.EntityID,
			Degree:   degree[n.ID],
		})
	}
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Degree != nodes[j].Degree {
			return nodes[i].Degree > nodes[j].Degree
		}
		if nodes[i].Type != nodes[j].Type {
			return nodes[i].Type < nodes[j].Type
		}
		return nodes[i].ID < nodes[j].ID
	})

	display := nodes
	if len(display) > MaxDisplayNodes {
		display = append([]NodeView(nil), display[:MaxDisplayNodes]...)
	}
	allowed := make(map[string]struct{}, len(display))
	for _, n := range display {
		allowed[n.ID] = struct{}{}
	}

	edges := make([]EdgeView, 0)
	for _, e := range snapshot.Edges {
		if _, ok := allowed[e.FromNodeID]; !ok {
			continue
		}
		if _, ok := allowed[e.ToNodeID]; !ok {
			continue
		}
		edges = append(edges, edgeViewFromModel(e))
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Type != edges[j].Type {
			return edges[i].Type < edges[j].Type
		}
		return edges[i].ID < edges[j].ID
	})

	return &Payload{
		RepositoryID: repo.ID,
		Stats: Stats{
			Communities: len(communityResult.Communities),
			GodNodes:    len(report.GodNodes),
			Surprises:   len(report.SurprisingConnections),
		},
		Nodes:      display,
		Edges:      edges,
		TotalNodes: len(snapshot.Nodes),
		TotalEdges: len(snapshot.Edges),
	}, nil
}

// NodeDetailFor returns evidence panel data for one node.
func NodeDetailFor(payload *Payload, nodeID string) (*NodeDetail, error) {
	var node *NodeView
	for i := range payload.Nodes {
		if payload.Nodes[i].ID == nodeID {
			node = &payload.Nodes[i]
			break
		}
	}
	if node == nil {
		return nil, fmt.Errorf("node %q not in display set (cap %d)", nodeID, MaxDisplayNodes)
	}
	detail := &NodeDetail{
		Node:  *node,
		Hints: hintsForType(node.Type),
	}
	for _, e := range payload.Edges {
		if e.To == nodeID {
			detail.Incoming = append(detail.Incoming, e)
		}
		if e.From == nodeID {
			detail.Outgoing = append(detail.Outgoing, e)
		}
	}
	return detail, nil
}

// FilterNodes returns nodes matching type and substring query.
func FilterNodes(nodes []NodeView, nodeType, query string) []NodeView {
	nodeType = strings.ToUpper(strings.TrimSpace(nodeType))
	query = strings.ToLower(strings.TrimSpace(query))
	var out []NodeView
	for _, n := range nodes {
		if nodeType != "" && nodeType != "ALL" && n.Type != nodeType {
			continue
		}
		if query != "" {
			hay := strings.ToLower(n.ID + " " + n.Type + " " + n.EntityID)
			if !strings.Contains(hay, query) {
				continue
			}
		}
		out = append(out, n)
	}
	return out
}

func degreeByNode(edges []*model.GraphEdge) map[string]int {
	deg := map[string]int{}
	for _, e := range edges {
		deg[e.FromNodeID]++
		deg[e.ToNodeID]++
	}
	return deg
}

func edgeViewFromModel(e *model.GraphEdge) EdgeView {
	evidence := e.EvidenceJSON
	if len(evidence) > 240 {
		evidence = evidence[:240] + "…"
	}
	return EdgeView{
		ID:       e.ID,
		From:     e.FromNodeID,
		To:       e.ToNodeID,
		Type:     string(e.EdgeType),
		Category: string(e.Category),
		Evidence: evidence,
	}
}

func hintsForType(nodeType string) []string {
	switch strings.ToUpper(nodeType) {
	case "DECISION":
		return []string{"explain_decision", "trace_decision", "analyze_topic_impact"}
	case "EVENT":
		return []string{"explain_event", "trace_event"}
	case "FACT":
		return []string{"get_fact", "analyze_topic_impact"}
	case "CONTRIBUTOR":
		return []string{"get_contributor", "trace_contributor"}
	default:
		return []string{"ask", "query_graph"}
	}
}
