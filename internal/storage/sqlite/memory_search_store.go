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

	for _, doc := range docs {
		if doc.RepositoryID != "" && doc.RepositoryID != repositoryID {
			continue
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM memory_search WHERE memory_id = ?`, doc.MemoryID); err != nil {
			return fmt.Errorf("failed to delete memory_search row %s: %w", doc.MemoryID, err)
		}
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
		content := scopedSearchContent(doc.RepositoryID, doc.Content)
		if _, err := stmt.ExecContext(ctx, doc.MemoryID, doc.Title, doc.EntityType, content); err != nil {
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
	matchQuery := buildFTSMatchQuery(terms, entityType, repositoryID)
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

	return hits, nil
}

func scopedSearchContent(repositoryID, content string) string {
	token := repoScopeToken(repositoryID)
	if token == "" {
		return content
	}
	return token + " " + content
}

func repoScopeToken(repositoryID string) string {
	token := sanitizeFTSTerm(repositoryID)
	if token == "" {
		return ""
	}
	return "repo" + token
}

func buildFTSMatchQuery(terms []string, entityType, repositoryID string) string {
	var parts []string
	if token := repoScopeToken(repositoryID); token != "" {
		parts = append(parts, token)
	}
	if entityType != "" {
		if token := sanitizeFTSTerm(entityType); token != "" {
			parts = append(parts, fmt.Sprintf("summary:%s", token))
		}
	}

	for _, term := range terms {
		for _, token := range tokenizeFTSTerms(term) {
			parts = append(parts, fmt.Sprintf("(title:%s* OR content:%s*)", token, token))
		}
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
		case r == '_':
			b.WriteRune(r)
		}
	}
	return b.String()
}

// tokenizeFTSTerms splits compound terms so FTS5 does not treat '-' as NOT.
func tokenizeFTSTerms(term string) []string {
	normalized := strings.ReplaceAll(strings.TrimSpace(term), "-", " ")
	fields := strings.Fields(normalized)
	if len(fields) == 0 {
		if token := sanitizeFTSTerm(term); token != "" {
			return []string{token}
		}
		return nil
	}
	seen := make(map[string]struct{}, len(fields))
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		token := sanitizeFTSTerm(f)
		if token == "" {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		out = append(out, token)
	}
	return out
}

