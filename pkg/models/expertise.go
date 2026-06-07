package models

// Expertise represents evidence-backed familiarity of a contributor with a domain.
type Expertise struct {
	ID            string
	RepositoryID  string
	ContributorID string
	Domain        string
	Score         float64
	EvidenceJSON  string
}
