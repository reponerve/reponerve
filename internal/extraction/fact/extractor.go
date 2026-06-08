package fact

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	models "github.com/reponerve/reponerve/pkg/models"
)

var (
	usesRegex      = regexp.MustCompile(`(?i)(.+?)\buses\b(.+)`)
	dependsOnRegex = regexp.MustCompile(`(?i)(.+?)\bdepends\s+on\b(.+)`)
	callsRegex     = regexp.MustCompile(`(?i)(.+?)\bcalls\b(.+)`)
	storesInRegex  = regexp.MustCompile(`(?i)(.+?)\bstores\s+data\s+in\b(.+)`)
)

// Extractor extracts Fact memories from ADR sources.
type Extractor struct{}

// NewExtractor creates a new Fact Extractor.
func NewExtractor() *Extractor {
	return &Extractor{}
}

// Extract processes ADR sources to parse relationship patterns.
func (e *Extractor) Extract(ctx context.Context, sources []*models.Source) ([]*memorymodels.Fact, error) {
	var facts []*memorymodels.Fact

	for _, src := range sources {
		if src.SourceType != "adr" {
			continue
		}

		text := ""
		if src.MetadataJSON != "" {
			var meta struct {
				Content string `json:"content"`
			}
			if err := json.Unmarshal([]byte(src.MetadataJSON), &meta); err == nil && meta.Content != "" {
				text = meta.Content
			}
		}
		if text == "" {
			continue
		}

		segments := splitIntoSegments(text)
		for _, segment := range segments {
			var subject, predicate, object string

			if matches := storesInRegex.FindStringSubmatch(segment); len(matches) == 3 {
				subject = cleanMarkdown(matches[1])
				predicate = "STORES_IN"
				object = cleanMarkdown(matches[2])
			} else if matches := dependsOnRegex.FindStringSubmatch(segment); len(matches) == 3 {
				subject = cleanMarkdown(matches[1])
				predicate = "DEPENDS_ON"
				object = cleanMarkdown(matches[2])
			} else if matches := usesRegex.FindStringSubmatch(segment); len(matches) == 3 {
				subject = cleanMarkdown(matches[1])
				predicate = "USES"
				object = cleanMarkdown(matches[2])
			} else if matches := callsRegex.FindStringSubmatch(segment); len(matches) == 3 {
				subject = cleanMarkdown(matches[1])
				predicate = "CALLS"
				object = cleanMarkdown(matches[2])
			}

			if subject != "" && object != "" {
				subject = toTitleCase(subject)
				object = toTitleCase(object)

				if subject == "" || object == "" {
					continue
				}

				facts = append(facts, &memorymodels.Fact{
					ID:           factID(src.ID, subject, predicate, object),
					RepositoryID: src.RepositoryID,
					Subject:      subject,
					Predicate:    predicate,
					Object:       object,
					SourceID:     src.ID,
					CreatedAt:    time.Now(),
				})
			}
		}
	}

	return facts, nil
}

func splitIntoSegments(text string) []string {
	var segments []string
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.ReplaceAll(line, "!", ".")
		line = strings.ReplaceAll(line, "?", ".")
		line = strings.ReplaceAll(line, ";", ".")
		line = strings.ReplaceAll(line, ",", ".")

		lowerLine := strings.ToLower(line)
		for _, conj := range []string{" and ", " but ", " also "} {
			for {
				idx := strings.Index(lowerLine, conj)
				if idx < 0 {
					break
				}
				line = line[:idx] + ". " + line[idx+len(conj):]
				lowerLine = lowerLine[:idx] + ". " + lowerLine[idx+len(conj):]
			}
		}

		parts := strings.Split(line, ".")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				segments = append(segments, part)
			}
		}
	}
	return segments
}

func cleanMarkdown(s string) string {
	for strings.HasPrefix(s, "#") {
		s = strings.TrimLeft(s, "#")
		s = strings.TrimSpace(s)
	}

	s = strings.TrimPrefix(s, "- ")
	s = strings.TrimPrefix(s, "* ")
	s = strings.TrimPrefix(s, "+ ")
	s = strings.TrimSpace(s)

	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "__", "")
	s = strings.ReplaceAll(s, "*", "")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "`", "")

	return strings.TrimSpace(s)
}

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

func factID(sourceID, subject, predicate, object string) string {
	h := sha256.Sum256([]byte(sourceID + subject + predicate + object))
	return "fact_" + hex.EncodeToString(h[:])
}
