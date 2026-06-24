package feature

// Summary is a derived feature view (domain, events, decisions) without a persisted table yet.
type Summary struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Keywords   []string `json:"keywords"`
	Sources    []string `json:"sources"`
	EventCount int      `json:"event_count,omitempty"`
}

// ListResult is the structured output for list_features.
type ListResult struct {
	Features       []Summary `json:"features"`
	SourceServices []string  `json:"source_services"`
}
