package explorecmd

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
)

type explorePayload struct {
	RepositoryID string        `json:"repository_id"`
	NodeCount    int           `json:"node_count"`
	EdgeCount    int           `json:"edge_count"`
	Communities  int           `json:"communities"`
	GodNodes     int           `json:"god_nodes"`
	Surprises    int           `json:"surprises"`
	Nodes        []exploreNode `json:"nodes"`
	Edges        []exploreEdge `json:"edges"`
}

type exploreNode struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	EntityID string `json:"entity_id"`
}

type exploreEdge struct {
	ID   string `json:"id"`
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

func renderExploreHTML(payload explorePayload) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	escaped := html.EscapeString(string(data))
	var b strings.Builder
	b.WriteString("<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>RepoNerve Graph</title>")
	b.WriteString("<style>body{font-family:system-ui,sans-serif;margin:1.5rem}pre{white-space:pre-wrap;background:#111;color:#eee;padding:1rem;border-radius:8px}</style>")
	b.WriteString("</head><body>")
	b.WriteString("<h1>RepoNerve Knowledge Graph</h1>")
	fmt.Fprintf(&b, "<p>Repository: %s · Nodes: %d · Edges: %d · Communities: %d · God nodes: %d · Surprises: %d</p>",
		html.EscapeString(payload.RepositoryID), payload.NodeCount, payload.EdgeCount, payload.Communities, payload.GodNodes, payload.Surprises)
	b.WriteString("<h2>Graph Data</h2><pre id=\"graph-data\">")
	b.WriteString(escaped)
	b.WriteString("</pre></body></html>")
	return b.String(), nil
}
