package exploreui

import (
	"bytes"
	"context"
	"encoding/json"
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

func graphNodesJSON(data *Payload) string {
	out := make([]cyNode, len(data.Nodes))
	for i, n := range data.Nodes {
		out[i] = cyNode{ID: n.ID, Type: n.Type}
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func graphEdgesJSON(data *Payload) string {
	out := make([]cyEdge, len(data.Edges))
	for i, e := range data.Edges {
		out[i] = cyEdge{ID: e.ID, From: e.From, To: e.To, Type: e.Type}
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// RenderHTML renders the explore page to HTML string.
func RenderHTML(data *Payload) (string, error) {
	var buf bytes.Buffer
	if err := ExplorePage(data).Render(context.Background(), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
