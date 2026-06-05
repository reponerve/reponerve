package models

import "time"

type ADR struct {
	ID           string
	RepositoryID string
	Title        string
	Status       string
	Path         string
	Content      string
	CreatedAt    time.Time
}
