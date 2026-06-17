package explorecmd

import (
	"strings"
	"testing"
)

func TestRenderExploreHTML(t *testing.T) {
	html, err := renderExploreHTML(explorePayload{
		RepositoryID: "repo_1",
		NodeCount:    2,
		EdgeCount:    1,
		Communities:  1,
		Nodes: []exploreNode{
			{ID: "n1", Type: "DECISION", EntityID: "d1"},
		},
		Edges: []exploreEdge{
			{ID: "e1", From: "n1", To: "n2", Type: "RELATES"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "RepoNerve Knowledge Graph") {
		t.Fatalf("missing title: %s", html)
	}
	if !strings.Contains(html, "graph-data") {
		t.Fatal("missing embedded graph data")
	}
}
