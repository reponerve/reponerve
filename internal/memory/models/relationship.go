package models

import "time"

// Relationship represents a link between two memories.
type Relationship struct {
	ID           string    `json:"id"`
	RepositoryID string    `json:"repository_id"`
	FromID       string    `json:"from_id"`
	ToID         string    `json:"to_id"`
	Type         string    `json:"type"`
	CreatedAt    time.Time `json:"created_at"`
}
