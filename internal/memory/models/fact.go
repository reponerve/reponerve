package models

import "time"

// Fact represents a deterministic repository truth memory.
type Fact struct {
	ID           string    `json:"id"`
	RepositoryID string    `json:"repository_id"`
	Subject      string    `json:"subject"`
	Predicate    string    `json:"predicate"`
	Object       string    `json:"object"`
	SourceID     string    `json:"source_id"`
	CreatedAt    time.Time `json:"created_at"`
}
