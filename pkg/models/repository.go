package models

import "time"

type Repository struct {
	ID            string
	Name          string
	Path          string
	DefaultBranch string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
