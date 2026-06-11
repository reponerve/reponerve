package sqlite

import (
	"context"
	"fmt"
	"strings"

	"github.com/reponerve/reponerve/internal/storage"
)

// MemorySearchStore implements storage.MemorySearchStore for SQLite FTS5.
type MemorySearchStore struct {
	db *Database
}

// NewMemorySearchStore creates a SQLite FTS5 memory search store.
func NewMemorySearchStore(db *Database) *MemorySearchStore {
	return &MemorySearchStore{db: db}
}

// Rebuild replaces the FTS index with documents for one repository scan.
func (s *MemorySearchStore) Rebuild(ctx context.Context, repositoryID string, docs []storage.MemorySearchDocument) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin memory search rebuild transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DELETE FROM memory_search"); err != nil {
		return fmt.Errorf("failed to clear memory_search index: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO memory_search (memory_id, title, summary, content)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare memory_search insert: %w", err)
	}
	defer stmt.Close()

	for _, doc := range docs {
		if doc.RepositoryID != "" && doc.RepositoryID != repositoryID {
			continue
		}
		if _, err := stmt.ExecContext(ctx, doc.MemoryID, doc.Title, doc.EntityType, doc.Content); err != nil {
			return fmt.Errorf("failed to index memory %s: %w", doc.MemoryID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit memory search rebuild: %w", err)
	}
	return nil
}

// Search queries FTS5 for repository memory matches.
func (s *MemorySearchStore) Search(ctx context.Context, repositoryID string, terms []string, entityType string) ([]storage.MemorySearchHit, error) {
	matchQuery := buildFTSMatchQuery(terms, entityType)
	if matchQuery == "" {
		return nil, nil
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT memory_id, summary, bm25(memory_search) AS rank
		FROM memory_search
		WHERE memory_search MATCH ?
		ORDER BY rank
		LIMIT 100
	`, matchQuery)
	if err != nil {
		return nil, fmt.Errorf("memory_search query failed: %w", err)
	}
	defer rows.Close()

	var hits []storage.MemorySearchHit
	for rows.Next() {
		var hit storage.MemorySearchHit
		if err := rows.Scan(&hit.MemoryID, &hit.EntityType, &hit.Rank); err != nil {
			return nil, fmt.Errorf("failed to scan memory_search row: %w", err)
		}
		hits = append(hits, hit)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("memory_search rows error: %w", err)
	}

	_ = repositoryID
	return hits, nil
}

func buildFTSMatchQuery(terms []string, entityType string) string {
	var parts []string
	if entityType != "" {
		parts = append(parts, fmt.Sprintf("summary:%s", sanitizeFTSTerm(entityType)))
	}

	for _, term := range terms {
		clean := sanitizeFTSTerm(term)
		if clean == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("(title:%s* OR content:%s*)", clean, clean))
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " AND ")
}

func sanitizeFTSTerm(term string) string {
	term = strings.TrimSpace(term)
	term = strings.TrimSuffix(term, "*")

	var b strings.Builder
	for _, r := range term {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == '-':
			b.WriteRune(r)
		}
	}
	return b.String()
}
