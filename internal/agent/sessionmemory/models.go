package sessionmemory

import (
	"time"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
)

const (
	PredicateSessionRemembered = "SESSION_REMEMBERED"
	PredicateSessionQA         = "SESSION_QA"
	HandoffVersion             = "v1"
)

// RememberRequest stores agent session knowledge as a traceable fact.
type RememberRequest struct {
	RepositoryID string `json:"repository_id"`
	Subject      string `json:"subject"`
	Content      string `json:"content"`
}

// WritebackRequest records a Q&A exchange as session memory.
type WritebackRequest struct {
	RepositoryID string `json:"repository_id"`
	Question     string `json:"question"`
	Answer       string `json:"answer"`
}

// HandoffBundle transfers session memory between agent sessions.
type HandoffBundle struct {
	Version       string                `json:"version"`
	RepositoryID  string                `json:"repository_id"`
	SessionID     string                `json:"session_id"`
	ExportedAt    time.Time             `json:"exported_at"`
	Facts         []*memorymodels.Fact  `json:"facts"`
	AccessRanking []string              `json:"access_ranking"`
}
