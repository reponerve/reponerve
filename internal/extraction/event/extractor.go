package event

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"reponerve/pkg/models"
)

// EventType constants defined by extraction-rules-v1.md.
const (
	EventTypeFeatureIntroduced    = "FEATURE_INTRODUCED"
	EventTypeDefectResolved       = "DEFECT_RESOLVED"
	EventTypeCodeRefactored       = "CODE_REFACTORED"
	EventTypeDocumentationUpdated = "DOCUMENTATION_UPDATED"
	EventTypeMaintenancePerformed = "MAINTENANCE_PERFORMED"
)

// commitPrefixes maps lowercase commit prefixes to their EventType.
var commitPrefixes = []struct {
	prefix    string
	eventType string
}{
	{"feat:", EventTypeFeatureIntroduced},
	{"feature:", EventTypeFeatureIntroduced},
	{"fix:", EventTypeDefectResolved},
	{"bugfix:", EventTypeDefectResolved},
	{"refactor:", EventTypeCodeRefactored},
	{"docs:", EventTypeDocumentationUpdated},
	{"chore:", EventTypeMaintenancePerformed},
}

// Extractor extracts Events from commit sources.
type Extractor struct{}

// NewExtractor creates a new Event Extractor.
func NewExtractor() *Extractor {
	return &Extractor{}
}

// Extract processes commit sources and returns qualifying Events.
// Non-matching commits are skipped. Each qualifying commit produces exactly one Event.
func (e *Extractor) Extract(_ context.Context, sources []*models.Source) ([]*models.Event, error) {
	var events []*models.Event

	for _, src := range sources {
		if src.SourceType != "commit" {
			continue
		}

		eventType, ok := classifyCommit(src.Title)
		if !ok {
			continue
		}

		title := deriveTitle(src.Title)
		if title == "" {
			title = src.Title
		}

		events = append(events, &models.Event{
			ID:           eventID(src.RepositoryID, src.ID),
			RepositoryID: src.RepositoryID,
			EventType:    eventType,
			Title:        title,
			Description:  "",
			SourceID:     src.ID,
			Timestamp:    src.Timestamp,
		})
	}

	return events, nil
}

// classifyCommit returns the EventType for a commit title, or false if unrecognized.
func classifyCommit(title string) (string, bool) {
	idx := strings.Index(title, ":")
	if idx < 0 {
		return "", false
	}
	prefixPart := strings.TrimSpace(title[:idx])

	// Handle scope: e.g. "feat(cache)" -> "feat"
	if pIdx := strings.Index(prefixPart, "("); pIdx >= 0 {
		prefixPart = strings.TrimSpace(prefixPart[:pIdx])
	}

	lowerType := strings.ToLower(prefixPart)

	// Match against our recognized types
	switch lowerType {
	case "feat", "feature":
		return EventTypeFeatureIntroduced, true
	case "fix", "bugfix":
		return EventTypeDefectResolved, true
	case "refactor":
		return EventTypeCodeRefactored, true
	case "docs":
		return EventTypeDocumentationUpdated, true
	case "chore":
		return EventTypeMaintenancePerformed, true
	}
	return "", false
}

// deriveTitle strips the conventional commit prefix (and optional scope) from the
// commit title and returns a cleaned, title-cased string.
//
// Examples:
//   "feat(cache): introduce redis cache" → "Introduce Redis Cache"
//   "fix: resolve null pointer"          → "Resolve Null Pointer"
func deriveTitle(commitTitle string) string {
	// Find the colon separator
	idx := strings.Index(commitTitle, ":")
	if idx < 0 {
		return toTitleCase(strings.TrimSpace(commitTitle))
	}

	rest := strings.TrimSpace(commitTitle[idx+1:])
	if rest == "" {
		return ""
	}

	return toTitleCase(rest)
}

// toTitleCase capitalises the first letter of each word.
func toTitleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) == 0 {
			continue
		}
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}

// eventID produces a stable, deterministic ID from the repository ID and source ID.
func eventID(repositoryID, sourceID string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("event:%s:%s", repositoryID, sourceID)))
	return "evt_" + hex.EncodeToString(h[:])
}
