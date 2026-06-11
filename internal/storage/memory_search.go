package storage

import "context"

// MemorySearchDocument is one FTS5-indexed repository memory record.
type MemorySearchDocument struct {
	MemoryID     string
	RepositoryID string
	EntityType   string
	Title        string
	Content      string
}

// MemorySearchHit is one FTS5 match ranked by bm25.
type MemorySearchHit struct {
	MemoryID   string
	EntityType string
	Rank       float64
}

// MemorySearchReader queries the FTS5 memory_search index.
type MemorySearchReader interface {
	Search(ctx context.Context, repositoryID string, terms []string, entityType string) ([]MemorySearchHit, error)
}

// MemorySearchStore rebuilds and queries the FTS5 memory_search index.
type MemorySearchStore interface {
	MemorySearchReader
	Rebuild(ctx context.Context, repositoryID string, docs []MemorySearchDocument) error
}
