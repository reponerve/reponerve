package linker

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	memorymodels "reponerve/internal/memory/models"
	models "reponerve/pkg/models"
)

// LinkInput holds the slices of extracted memories to link.
type LinkInput struct {
	Events    []*models.Event
	Decisions []*memorymodels.Decision
	Intents   []*memorymodels.Intent
	Facts     []*memorymodels.Fact
}

// Linker traverses extracted memories and creates relationships between them.
type Linker struct{}

// NewLinker creates a new Linker instance.
func NewLinker() *Linker {
	return &Linker{}
}

// Link evaluates memories to produce deterministic Relationship memories.
func (l *Linker) Link(ctx context.Context, input LinkInput) ([]*memorymodels.Relationship, error) {
	var relationships []*memorymodels.Relationship

	// 1. INTENT_DRIVES_DECISION: Intent.SourceID == Decision.SourceID
	for _, it := range input.Intents {
		for _, dec := range input.Decisions {
			if it.SourceID != "" && it.SourceID == dec.SourceID {
				relType := "INTENT_DRIVES_DECISION"
				relationships = append(relationships, &memorymodels.Relationship{
					ID:           relationshipID(it.ID, dec.ID, relType),
					RepositoryID: dec.RepositoryID,
					FromID:       it.ID,
					ToID:         dec.ID,
					Type:         relType,
					CreatedAt:    time.Now(),
				})
			}
		}
	}

	// 2. DECISION_RESULTS_IN_EVENT: Title similarity between Decision and Event
	for _, dec := range input.Decisions {
		decClean := getCleanWords(dec.Title)
		if len(decClean) == 0 {
			continue
		}
		for _, evt := range input.Events {
			evtClean := getCleanWords(evt.Title)
			if len(evtClean) == 0 {
				continue
			}

			// Check if they share at least one non-stop word
			matches := false
			for w := range decClean {
				if evtClean[w] {
					matches = true
					break
				}
			}

			if matches {
				relType := "DECISION_RESULTS_IN_EVENT"
				relationships = append(relationships, &memorymodels.Relationship{
					ID:           relationshipID(dec.ID, evt.ID, relType),
					RepositoryID: dec.RepositoryID,
					FromID:       dec.ID,
					ToID:         evt.ID,
					Type:         relType,
					CreatedAt:    time.Now(),
				})
			}
		}
	}

	// 3. FACT_SUPPORTS_DECISION: Fact Subject or Object overlaps with Decision Title
	for _, f := range input.Facts {
		factClean := make(map[string]bool)
		for w := range getCleanWords(f.Subject) {
			factClean[w] = true
		}
		for w := range getCleanWords(f.Object) {
			factClean[w] = true
		}

		if len(factClean) == 0 {
			continue
		}

		for _, dec := range input.Decisions {
			decClean := getCleanWords(dec.Title)
			if len(decClean) == 0 {
				continue
			}

			// Check for keyword overlap
			matches := false
			for w := range factClean {
				if decClean[w] {
					matches = true
					break
				}
			}

			if matches {
				relType := "FACT_SUPPORTS_DECISION"
				relationships = append(relationships, &memorymodels.Relationship{
					ID:           relationshipID(f.ID, dec.ID, relType),
					RepositoryID: f.RepositoryID,
					FromID:       f.ID,
					ToID:         dec.ID,
					Type:         relType,
					CreatedAt:    time.Now(),
				})
			}
		}
	}

	return relationships, nil
}

func normalizeTitle(title string) string {
	title = strings.ToLower(title)
	var sb strings.Builder
	for _, r := range title {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
		} else {
			sb.WriteRune(' ')
		}
	}
	return sb.String()
}

func getCleanWords(title string) map[string]bool {
	normalized := normalizeTitle(title)
	words := strings.Fields(normalized)
	stopWords := map[string]bool{
		"use": true, "uses": true, "using": true,
		"introduce": true, "introduces": true, "introducing": true,
		"implement": true, "implements": true, "implementing": true,
		"add": true, "adds": true, "adding": true, "new": true,
		"to": true, "the": true, "a": true, "an": true, "for": true,
		"in": true, "of": true, "and": true, "or": true, "with": true,
		"by": true, "on": true, "at": true, "from": true,
		"feat": true, "fix": true, "chore": true, "docs": true, "refactor": true,
	}
	clean := make(map[string]bool)
	for _, w := range words {
		if !stopWords[w] && len(w) > 0 {
			clean[w] = true
		}
	}
	return clean
}

func relationshipID(fromID, toID, relType string) string {
	h := sha256.Sum256([]byte(fromID + ":" + toID + ":" + relType))
	return "relationship_" + hex.EncodeToString(h[:])
}
