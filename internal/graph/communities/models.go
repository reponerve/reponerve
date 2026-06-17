package communities

// Community is a deterministically ordered connected component in the knowledge graph.
type Community struct {
	ID      string   `json:"id"`
	NodeIDs []string `json:"node_ids"`
	Size    int      `json:"size"`
}

// DetectionResult holds communities for a repository graph snapshot.
type DetectionResult struct {
	RepositoryID string      `json:"repository_id"`
	Communities  []Community `json:"communities"`
}
