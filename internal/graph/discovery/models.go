package discovery

// GodNode is a high-degree hub in the knowledge graph.
type GodNode struct {
	NodeID string `json:"node_id"`
	Degree int    `json:"degree"`
}

// SurprisingConnection links nodes in different communities.
type SurprisingConnection struct {
	FromNodeID    string `json:"from_node_id"`
	ToNodeID      string `json:"to_node_id"`
	EdgeType      string `json:"edge_type"`
	FromCommunity string `json:"from_community"`
	ToCommunity   string `json:"to_community"`
}

// GraphDiscoveryReport summarizes graph structure for agents.
type GraphDiscoveryReport struct {
	RepositoryID           string                 `json:"repository_id"`
	GodNodes               []GodNode              `json:"god_nodes"`
	SurprisingConnections  []SurprisingConnection `json:"surprising_connections"`
	SuggestedQuestions     []string               `json:"suggested_questions"`
}
