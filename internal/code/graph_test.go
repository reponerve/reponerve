package code

import (
	"testing"

	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

func TestBuildCallGraphFromRelationships(t *testing.T) {
	rels := []*codemodels.CodeRelationship{
		{FromEntityID: "a", ToEntityID: "b", RelationshipType: "CALLS"},
		{FromEntityID: "b", ToEntityID: "c", RelationshipType: "CALLS"},
		{FromEntityID: "a", ToEntityID: "d", RelationshipType: "IMPORTS"},
	}
	graph := BuildCallGraphFromRelationships("a", rels, 5)
	if len(graph.Edges) != 2 {
		t.Fatalf("expected 2 call edges, got %d", len(graph.Edges))
	}
}

func TestCollectSymbolDependencies(t *testing.T) {
	outbound := []*codemodels.CodeRelationship{
		{RelationshipType: "IMPORTS", ToEntityID: "pkg"},
		{RelationshipType: "DEFINED_IN_FILE", ToEntityID: "file"},
	}
	deps := CollectSymbolDependencies("sym", outbound)
	if len(deps) != 1 || deps[0].RelationshipType != "IMPORTS" {
		t.Fatalf("expected only IMPORTS dependency, got %+v", deps)
	}
}
