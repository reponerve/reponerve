package models

import "time"

// Contributor represents an identifiable repository participant.
type Contributor struct {
	ID           string
	RepositoryID string
	Name         string
	Email        string
	FirstSeen    time.Time
	LastSeen     time.Time
	CommitCount  int
}
