package exploreui

import (
	"bytes"
	"context"
)

type cyNode struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type cyEdge struct {
	ID   string `json:"id"`
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

func graphNodes(data *Payload) []cyNode {
	out := make([]cyNode, len(data.Nodes))
	for i, n := range data.Nodes {
		out[i] = cyNode{ID: n.ID, Type: n.Type}
	}
	return out
}

func graphEdges(data *Payload) []cyEdge {
	out := make([]cyEdge, len(data.Edges))
	for i, e := range data.Edges {
		out[i] = cyEdge{ID: e.ID, From: e.From, To: e.To, Type: e.Type}
	}
	return out
}

// RenderHTML renders the explore page to HTML string.
func RenderHTML(data *Payload) (string, error) {
	var buf bytes.Buffer
	if err := ExplorePage(data).Render(context.Background(), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
