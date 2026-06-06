package models

import "time"

// Intent represents the reason or goal behind a repository decision.
type Intent struct {
	ID           string    `json:"id"`
	RepositoryID string    `json:"repository_id"`
	Description  string    `json:"description"`
	SourceID     string    `json:"source_id"`
	CreatedAt    time.Time `json:"created_at"`
}
