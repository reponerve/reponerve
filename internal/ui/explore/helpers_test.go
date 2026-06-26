package exploreui

import (
	"strings"
	"testing"
)

func TestRenderHTML(t *testing.T) {
	html, err := RenderHTML(&Payload{
		RepositoryID: "repo_1",
		TotalNodes:   2,
		TotalEdges:   1,
		Stats:        Stats{Communities: 1},
		Nodes: []NodeView{
			{ID: "n1", Type: "DECISION", EntityID: "d1", Degree: 1},
		},
		Edges: []EdgeView{
			{ID: "e1", From: "n1", To: "n2", Type: "RELATES", Category: "STORED", Evidence: `{"source":"test"}`},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "RepoNerve Explore") {
		t.Fatalf("missing title")
	}
	if !strings.Contains(html, "cytoscape") {
		t.Fatal("missing cytoscape")
	}
	if strings.Contains(html, `const nodes = "[`) {
		t.Fatal("graph nodes rendered as a string literal")
	}
	if !strings.Contains(html, `const nodes = [{"id":"n1","type":"DECISION"}];`) {
		t.Fatal("graph nodes were not rendered as a JavaScript array")
	}
}

func TestFilterNodes(t *testing.T) {
	nodes := []NodeView{
		{ID: "a", Type: "DECISION", EntityID: "auth"},
		{ID: "b", Type: "EVENT", EntityID: "login"},
	}
	out := FilterNodes(nodes, "DECISION", "")
	if len(out) != 1 || out[0].ID != "a" {
		t.Fatalf("got %+v", out)
	}
	out = FilterNodes(nodes, "", "login")
	if len(out) != 1 || out[0].ID != "b" {
		t.Fatalf("got %+v", out)
	}
}

func TestNodeDetailFor(t *testing.T) {
	payload := &Payload{
		Nodes: []NodeView{{ID: "n1", Type: "DECISION", EntityID: "d1"}},
		Edges: []EdgeView{{ID: "e1", From: "n1", To: "n2", Type: "LINKS", Category: "DERIVED", Evidence: "{}"}},
	}
	d, err := NodeDetailFor(payload, "n1")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Outgoing) != 1 || len(d.Hints) == 0 {
		t.Fatalf("detail: %+v", d)
	}
}
