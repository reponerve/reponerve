package discovery

import (
	"testing"

	"github.com/reponerve/reponerve/internal/graph/communities"
	"github.com/reponerve/reponerve/internal/graph/model"
)

func TestAnalyzeSurprisingConnections(t *testing.T) {
	repoID := "repo_1"
	nodes := []*model.GraphNode{
		model.NewNode(repoID, model.NodeTypeDecision, "d1"),
		model.NewNode(repoID, model.NodeTypeDecision, "d2"),
		model.NewNode(repoID, model.NodeTypeFact, "f1"),
	}
	edges := []*model.GraphEdge{
		model.NewEdge(repoID, nodes[0].ID, nodes[1].ID, "RELATES", model.CategoryStored, `{"relationship_id":"r1"}`),
		model.NewEdge(repoID, nodes[0].ID, nodes[2].ID, "BRIDGE", model.CategoryDerived, `{"match_type":"keyword"}`),
	}
	communities := communities.Detect(repoID, nodes, edges)

	report, err := Analyze(repoID, nodes, edges, communities)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.GodNodes) == 0 {
		t.Fatal("expected god node on d1")
	}
	if len(report.SurprisingConnections) == 0 {
		t.Fatal("expected cross-community surprise edge")
	}
}
