package models

import "time"

// Source represents an ingested repository artifact source (e.g. commit, ADR, PR).
type Source struct {
	ID           string
	RepositoryID string
	SourceType   string
	Reference    string
	Title        string
	Author       string
	Timestamp    time.Time
}
