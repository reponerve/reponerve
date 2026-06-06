package intent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	memorymodels "reponerve/internal/memory/models"
	models "reponerve/pkg/models"
)

var keywords = []string{
	"improve", "reduce", "increase", "optimize", "enhance",
	"simplify", "minimize", "accelerate", "stabilize", "secure",
}

// Extractor extracts Intent memories from ADR and commit sources.
type Extractor struct{}

// NewExtractor creates a new Intent Extractor.
func NewExtractor() *Extractor {
	return &Extractor{}
}

// Extract processes ADR and commit sources to parse matching statements containing keywords.
func (e *Extractor) Extract(ctx context.Context, sources []*models.Source) ([]*memorymodels.Intent, error) {
	var intents []*memorymodels.Intent

	for _, src := range sources {
		text := ""
		if src.SourceType == "adr" {
			if src.MetadataJSON != "" {
				var meta struct {
					Content string `json:"content"`
				}
				if err := json.Unmarshal([]byte(src.MetadataJSON), &meta); err == nil && meta.Content != "" {
					text = meta.Content
				}
			}
			if text == "" {
				text = src.Title
			}
		} else if src.SourceType == "commit" {
			text = src.Title
		} else {
			continue
		}

		segments := splitIntoSegments(text)
		for _, segment := range segments {
			lowerSeg := strings.ToLower(segment)
			firstKeywordIdx := -1

			for _, kw := range keywords {
				idx := strings.Index(lowerSeg, kw)
				if idx >= 0 {
					if firstKeywordIdx == -1 || idx < firstKeywordIdx {
						firstKeywordIdx = idx
					}
				}
			}

			if firstKeywordIdx >= 0 {
				phrase := segment[firstKeywordIdx:]
				phrase = strings.TrimSpace(phrase)
				phrase = strings.TrimRight(phrase, ".,;!?")
				phrase = cleanMarkdown(phrase)

				if phrase == "" {
					continue
				}

				desc := toTitleCase(phrase)
				intents = append(intents, &memorymodels.Intent{
					ID:           intentID(src.ID, desc),
					RepositoryID: src.RepositoryID,
					Description:  desc,
					SourceID:     src.ID,
					CreatedAt:    time.Now(),
				})
			}
		}
	}

	return intents, nil
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

	if strings.HasPrefix(s, "- ") || strings.HasPrefix(s, "* ") || strings.HasPrefix(s, "+ ") {
		s = s[2:]
		s = strings.TrimSpace(s)
	}

	if idx := strings.Index(s, ". "); idx > 0 {
		isNumeric := true
		for _, char := range s[:idx] {
			if char < '0' || char > '9' {
				isNumeric = false
				break
			}
		}
		if isNumeric {
			s = s[idx+2:]
			s = strings.TrimSpace(s)
		}
	}

	s = strings.Trim(s, "*_`")

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

func intentID(sourceID, description string) string {
	h := sha256.Sum256([]byte(sourceID + ":" + strings.ToLower(description)))
	return "intent_" + hex.EncodeToString(h[:])
}
