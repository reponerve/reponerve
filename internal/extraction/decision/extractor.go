package decision

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	models "github.com/reponerve/reponerve/pkg/models"
)

// Extractor extracts Decision memories from ADR sources.
type Extractor struct{}

// NewExtractor creates a new Decision Extractor.
func NewExtractor() *Extractor {
	return &Extractor{}
}

// Extract processes ADR sources and returns qualifying Decisions.
// Non-ADR sources are ignored.
func (e *Extractor) Extract(ctx context.Context, sources []*models.Source) ([]*memorymodels.Decision, error) {
	var decisions []*memorymodels.Decision

	for _, src := range sources {
		if src.SourceType != "adr" {
			continue
		}

		status := "Proposed"
		if src.MetadataJSON != "" {
			var meta struct {
				Status string `json:"status"`
			}
			if err := json.Unmarshal([]byte(src.MetadataJSON), &meta); err == nil && meta.Status != "" {
				status = meta.Status
			}
		}

		decisions = append(decisions, &memorymodels.Decision{
			ID:           decisionID(src.ID),
			RepositoryID: src.RepositoryID,
			Title:        src.Title,
			Status:       status,
			SourceID:     src.ID,
			CreatedAt:    time.Now(),
		})
	}

	return decisions, nil
}

// decisionID produces a stable, deterministic ID from the source ID.
func decisionID(sourceID string) string {
	h := sha256.Sum256([]byte(sourceID))
	return "decision_" + hex.EncodeToString(h[:])
}
