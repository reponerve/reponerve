package extraction

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/reponerve/reponerve/pkg/models"
)

// Extractor extracts Contributor records from commit sources.
type Extractor struct {
}

// NewExtractor creates a new Contributor Extractor.
func NewExtractor() *Extractor {
	return &Extractor{}
}

// authorRegex matches conventional Git author format "Name <email>"
var authorRegex = regexp.MustCompile(`^([^<]+)\s*<([^>]+)>$`)

type contribData struct {
	repositoryID string
	name         string
	email        string
	firstSeen    time.Time
	lastSeen     time.Time
	commitCount  int
}

// Extract parses commit sources, aggregates statistics, groups contributors, and returns them deterministically.
func (e *Extractor) Extract(ctx context.Context, sources []*models.Source) ([]*models.Contributor, error) {
	contribMap := make(map[string]*contribData)

	for _, src := range sources {
		if src.SourceType != "commit" {
			continue
		}

		name := strings.TrimSpace(src.Author)
		email := ""

		matches := authorRegex.FindStringSubmatch(src.Author)
		if len(matches) == 3 {
			name = strings.TrimSpace(matches[1])
			email = strings.TrimSpace(matches[2])
		} else if strings.Contains(src.Author, "@") && !strings.Contains(src.Author, " ") {
			email = strings.TrimSpace(src.Author)
			name = ""
		}

		if name == "" && email == "" {
			continue
		}

		// Grouping key: group by email if present, otherwise group by name.
		var mapKey string
		if email != "" {
			mapKey = src.RepositoryID + ":email:" + email
		} else {
			mapKey = src.RepositoryID + ":name:" + name
		}

		data, exists := contribMap[mapKey]
		if !exists {
			data = &contribData{
				repositoryID: src.RepositoryID,
				name:         name,
				email:        email,
				firstSeen:    src.Timestamp,
				lastSeen:     src.Timestamp,
				commitCount:  0,
			}
			contribMap[mapKey] = data
		}

		data.commitCount++
		if src.Timestamp.Before(data.firstSeen) {
			data.firstSeen = src.Timestamp
		}
		if src.Timestamp.After(data.lastSeen) {
			data.lastSeen = src.Timestamp
		}

		// Keep the first non-empty name, or if name is updated to a longer one.
		if data.name == "" && name != "" {
			data.name = name
		} else if name != "" && len(name) > len(data.name) {
			data.name = name
		}
	}

	var contributors []*models.Contributor
	for _, data := range contribMap {
		id := contributorID(data.repositoryID, data.name, data.email)
		contributors = append(contributors, &models.Contributor{
			ID:           id,
			RepositoryID: data.repositoryID,
			Name:         data.name,
			Email:        data.email,
			FirstSeen:    data.firstSeen,
			LastSeen:     data.lastSeen,
			CommitCount:  data.commitCount,
		})
	}

	// Sort by ID to ensure deterministic output order.
	sort.Slice(contributors, func(i, j int) bool {
		return contributors[i].ID < contributors[j].ID
	})

	return contributors, nil
}

func contributorID(repositoryID, name, email string) string {
	var input string
	if email != "" {
		input = repositoryID + ":" + email
	} else {
		input = repositoryID + ":" + name
	}
	h := sha256.Sum256([]byte(input))
	return "ctr_" + hex.EncodeToString(h[:])
}
