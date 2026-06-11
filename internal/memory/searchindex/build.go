package searchindex

import (
	"strings"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	"github.com/reponerve/reponerve/internal/storage"
	models "github.com/reponerve/reponerve/pkg/models"
)

// Input holds extracted repository memories to index for search.
type Input struct {
	RepositoryID string
	Events       []*models.Event
	Decisions    []*memorymodels.Decision
	Facts        []*memorymodels.Fact
}

// BuildDocuments converts extracted memories into FTS5 documents.
func BuildDocuments(in Input) []storage.MemorySearchDocument {
	var docs []storage.MemorySearchDocument

	for _, ev := range in.Events {
		if ev == nil || ev.RepositoryID != in.RepositoryID {
			continue
		}
		docs = append(docs, storage.MemorySearchDocument{
			MemoryID:     ev.ID,
			RepositoryID: ev.RepositoryID,
			EntityType:   "EVENT",
			Title:        ev.Title,
			Content:      joinNonEmpty(ev.Description, ev.EventType),
		})
	}

	for _, d := range in.Decisions {
		if d == nil || d.RepositoryID != in.RepositoryID {
			continue
		}
		docs = append(docs, storage.MemorySearchDocument{
			MemoryID:     d.ID,
			RepositoryID: d.RepositoryID,
			EntityType:   "DECISION",
			Title:        d.Title,
			Content:      d.Status,
		})
	}

	for _, f := range in.Facts {
		if f == nil || f.RepositoryID != in.RepositoryID {
			continue
		}
		docs = append(docs, storage.MemorySearchDocument{
			MemoryID:     f.ID,
			RepositoryID: f.RepositoryID,
			EntityType:   "FACT",
			Title:        f.Subject,
			Content:      joinNonEmpty(f.Predicate, f.Object),
		})
	}

	return docs
}

func joinNonEmpty(parts ...string) string {
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, " ")
}
