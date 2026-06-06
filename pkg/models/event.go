package models

import "time"

// Event represents something that happened in the repository.
// Events are extracted from commit sources.
type Event struct {
	ID           string
	RepositoryID string
	EventType    string
	Title        string
	Description  string
	SourceID     string
	Timestamp    time.Time
}
