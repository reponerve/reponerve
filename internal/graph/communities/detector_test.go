package communities

import (
	"testing"

	"github.com/reponerve/reponerve/internal/graph/model"
)

func TestDetectConnectedComponents(t *testing.T) {
	repoID := "repo_1"
	nodes := []*model.GraphNode{
		model.NewNode(repoID, model.NodeTypeDecision, "d1"),
		model.NewNode(repoID, model.NodeTypeDecision, "d2"),
		model.NewNode(repoID, model.NodeTypeFact, "f1"),
	}
	edges := []*model.GraphEdge{
		model.NewEdge(repoID, nodes[0].ID, nodes[1].ID, "RELATES", model.CategoryStored, `{"relationship_id":"r1"}`),
	}

	result := Detect(repoID, nodes, edges)
	if err := Validate(result); err != nil {
		t.Fatal(err)
	}
	if len(result.Communities) != 2 {
		t.Fatalf("expected 2 communities, got %d", len(result.Communities))
	}
	if result.Communities[0].Size != 2 {
		t.Fatalf("expected largest community size 2, got %d", result.Communities[0].Size)
	}
}

func TestDetectDeterministicOrdering(t *testing.T) {
	repoID := "repo_1"
	nodes := []*model.GraphNode{
		model.NewNode(repoID, model.NodeTypeDecision, "d1"),
		model.NewNode(repoID, model.NodeTypeFact, "f1"),
	}
	a := Detect(repoID, nodes, nil)
	b := Detect(repoID, nodes, nil)
	if a.Communities[0].ID != b.Communities[0].ID {
		t.Fatalf("community IDs not deterministic")
	}
}
