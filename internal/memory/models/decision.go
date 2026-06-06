package models

import "time"

// Decision represents an architectural decision record memory.
type Decision struct {
	ID           string    `json:"id"`
	RepositoryID string    `json:"repository_id"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	SourceID     string    `json:"source_id"`
	CreatedAt    time.Time `json:"created_at"`
}
