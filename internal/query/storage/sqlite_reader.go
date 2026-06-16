package storage

import (
	"context"
	"fmt"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
	models "github.com/reponerve/reponerve/pkg/models"
)

// SQLiteEventReader implements EventReader for SQLite.
type SQLiteEventReader struct {
	db *sqlite.Database
}

func NewSQLiteEventReader(db *sqlite.Database) *SQLiteEventReader {
	return &SQLiteEventReader{db: db}
}

func (r *SQLiteEventReader) GetByID(ctx context.Context, id string) (*models.Event, error) {
	var event models.Event
	var ts sqlite.FlexibleTime
	query := `
		SELECT id, repository_id, event_type, title, COALESCE(description, ''), source_id, timestamp
		FROM memory_events
		WHERE id = ?
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.RepositoryID,
		&event.EventType,
		&event.Title,
		&event.Description,
		&event.SourceID,
		&ts,
	)
	if err != nil {
		return nil, err
	}
	event.Timestamp = ts.Time
	return &event, nil
}

func (r *SQLiteEventReader) ListByRepository(ctx context.Context, repositoryID string) ([]*models.Event, error) {
	query := `
		SELECT id, repository_id, event_type, title, COALESCE(description, ''), source_id, timestamp
		FROM memory_events
		WHERE repository_id = ?
	`
	rows, err := r.db.QueryContext(ctx, query, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list events by repository: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		var ts sqlite.FlexibleTime
		err := rows.Scan(
			&event.ID,
			&event.RepositoryID,
			&event.EventType,
			&event.Title,
			&event.Description,
			&event.SourceID,
			&ts,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		event.Timestamp = ts.Time
		events = append(events, &event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return events, nil
}

func (r *SQLiteEventReader) ListAll(ctx context.Context) ([]*models.Event, error) {
	query := `
		SELECT id, repository_id, event_type, title, COALESCE(description, ''), source_id, timestamp
		FROM memory_events
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all events: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		var ts sqlite.FlexibleTime
		err := rows.Scan(
			&event.ID,
			&event.RepositoryID,
			&event.EventType,
			&event.Title,
			&event.Description,
			&event.SourceID,
			&ts,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		event.Timestamp = ts.Time
		events = append(events, &event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return events, nil
}

// SQLiteDecisionReader implements DecisionReader for SQLite.
type SQLiteDecisionReader struct {
	db *sqlite.Database
}

func NewSQLiteDecisionReader(db *sqlite.Database) *SQLiteDecisionReader {
	return &SQLiteDecisionReader{db: db}
}

func (r *SQLiteDecisionReader) GetByID(ctx context.Context, id string) (*memorymodels.Decision, error) {
	var dec memorymodels.Decision
	var createdAt sqlite.FlexibleTime
	query := `
		SELECT id, repository_id, title, status, source_id, created_at
		FROM memory_decisions
		WHERE id = ?
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&dec.ID,
		&dec.RepositoryID,
		&dec.Title,
		&dec.Status,
		&dec.SourceID,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}
	dec.CreatedAt = flexibleTime(createdAt)
	return &dec, nil
}

func (r *SQLiteDecisionReader) ListByRepository(ctx context.Context, repositoryID string) ([]*memorymodels.Decision, error) {
	query := `
		SELECT id, repository_id, title, status, source_id, created_at
		FROM memory_decisions
		WHERE repository_id = ?
	`
	rows, err := r.db.QueryContext(ctx, query, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list decisions by repository: %w", err)
	}
	defer rows.Close()

	var decisions []*memorymodels.Decision
	for rows.Next() {
		var dec memorymodels.Decision
		var createdAt sqlite.FlexibleTime
		err := rows.Scan(
			&dec.ID,
			&dec.RepositoryID,
			&dec.Title,
			&dec.Status,
			&dec.SourceID,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan decision: %w", err)
		}
		dec.CreatedAt = flexibleTime(createdAt)
		decisions = append(decisions, &dec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return decisions, nil
}

func (r *SQLiteDecisionReader) ListAll(ctx context.Context) ([]*memorymodels.Decision, error) {
	query := `
		SELECT id, repository_id, title, status, source_id, created_at
		FROM memory_decisions
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all decisions: %w", err)
	}
	defer rows.Close()

	var decisions []*memorymodels.Decision
	for rows.Next() {
		var dec memorymodels.Decision
		var createdAt sqlite.FlexibleTime
		err := rows.Scan(
			&dec.ID,
			&dec.RepositoryID,
			&dec.Title,
			&dec.Status,
			&dec.SourceID,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan decision: %w", err)
		}
		dec.CreatedAt = flexibleTime(createdAt)
		decisions = append(decisions, &dec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return decisions, nil
}

// SQLiteIntentReader implements IntentReader for SQLite.
type SQLiteIntentReader struct {
	db *sqlite.Database
}

func NewSQLiteIntentReader(db *sqlite.Database) *SQLiteIntentReader {
	return &SQLiteIntentReader{db: db}
}

func (r *SQLiteIntentReader) GetByID(ctx context.Context, id string) (*memorymodels.Intent, error) {
	var intent memorymodels.Intent
	var createdAt sqlite.FlexibleTime
	query := `
		SELECT id, repository_id, description, source_id, created_at
		FROM memory_intents
		WHERE id = ?
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&intent.ID,
		&intent.RepositoryID,
		&intent.Description,
		&intent.SourceID,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}
	intent.CreatedAt = flexibleTime(createdAt)
	return &intent, nil
}

func (r *SQLiteIntentReader) ListByRepository(ctx context.Context, repositoryID string) ([]*memorymodels.Intent, error) {
	query := `
		SELECT id, repository_id, description, source_id, created_at
		FROM memory_intents
		WHERE repository_id = ?
	`
	rows, err := r.db.QueryContext(ctx, query, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list intents by repository: %w", err)
	}
	defer rows.Close()

	var intents []*memorymodels.Intent
	for rows.Next() {
		var intent memorymodels.Intent
		var createdAt sqlite.FlexibleTime
		err := rows.Scan(
			&intent.ID,
			&intent.RepositoryID,
			&intent.Description,
			&intent.SourceID,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan intent: %w", err)
		}
		intent.CreatedAt = flexibleTime(createdAt)
		intents = append(intents, &intent)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return intents, nil
}

func (r *SQLiteIntentReader) ListAll(ctx context.Context) ([]*memorymodels.Intent, error) {
	query := `
		SELECT id, repository_id, description, source_id, created_at
		FROM memory_intents
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all intents: %w", err)
	}
	defer rows.Close()

	var intents []*memorymodels.Intent
	for rows.Next() {
		var intent memorymodels.Intent
		var createdAt sqlite.FlexibleTime
		err := rows.Scan(
			&intent.ID,
			&intent.RepositoryID,
			&intent.Description,
			&intent.SourceID,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan intent: %w", err)
		}
		intent.CreatedAt = flexibleTime(createdAt)
		intents = append(intents, &intent)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return intents, nil
}

// SQLiteFactReader implements FactReader for SQLite.
type SQLiteFactReader struct {
	db *sqlite.Database
}

func NewSQLiteFactReader(db *sqlite.Database) *SQLiteFactReader {
	return &SQLiteFactReader{db: db}
}

func (r *SQLiteFactReader) GetByID(ctx context.Context, id string) (*memorymodels.Fact, error) {
	var fact memorymodels.Fact
	var createdAt sqlite.FlexibleTime
	query := `
		SELECT id, repository_id, subject, predicate, object, source_id, created_at
		FROM memory_facts
		WHERE id = ?
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&fact.ID,
		&fact.RepositoryID,
		&fact.Subject,
		&fact.Predicate,
		&fact.Object,
		&fact.SourceID,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}
	fact.CreatedAt = flexibleTime(createdAt)
	return &fact, nil
}

func (r *SQLiteFactReader) ListByRepository(ctx context.Context, repositoryID string) ([]*memorymodels.Fact, error) {
	query := `
		SELECT id, repository_id, subject, predicate, object, source_id, created_at
		FROM memory_facts
		WHERE repository_id = ?
	`
	rows, err := r.db.QueryContext(ctx, query, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list facts by repository: %w", err)
	}
	defer rows.Close()

	var facts []*memorymodels.Fact
	for rows.Next() {
		var fact memorymodels.Fact
		var createdAt sqlite.FlexibleTime
		err := rows.Scan(
			&fact.ID,
			&fact.RepositoryID,
			&fact.Subject,
			&fact.Predicate,
			&fact.Object,
			&fact.SourceID,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan fact: %w", err)
		}
		fact.CreatedAt = flexibleTime(createdAt)
		facts = append(facts, &fact)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return facts, nil
}

func (r *SQLiteFactReader) ListAll(ctx context.Context) ([]*memorymodels.Fact, error) {
	query := `
		SELECT id, repository_id, subject, predicate, object, source_id, created_at
		FROM memory_facts
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all facts: %w", err)
	}
	defer rows.Close()

	var facts []*memorymodels.Fact
	for rows.Next() {
		var fact memorymodels.Fact
		var createdAt sqlite.FlexibleTime
		err := rows.Scan(
			&fact.ID,
			&fact.RepositoryID,
			&fact.Subject,
			&fact.Predicate,
			&fact.Object,
			&fact.SourceID,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan fact: %w", err)
		}
		fact.CreatedAt = flexibleTime(createdAt)
		facts = append(facts, &fact)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return facts, nil
}

// SQLiteRelationshipReader implements RelationshipReader for SQLite.
type SQLiteRelationshipReader struct {
	db *sqlite.Database
}

func NewSQLiteRelationshipReader(db *sqlite.Database) *SQLiteRelationshipReader {
	return &SQLiteRelationshipReader{db: db}
}

func (r *SQLiteRelationshipReader) GetByID(ctx context.Context, id string) (*memorymodels.Relationship, error) {
	var rel memorymodels.Relationship
	var createdAt sqlite.FlexibleTime
	query := `
		SELECT id, repository_id, from_id, to_id, relationship_type, created_at
		FROM memory_relationships
		WHERE id = ?
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rel.ID,
		&rel.RepositoryID,
		&rel.FromID,
		&rel.ToID,
		&rel.Type,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}
	rel.CreatedAt = flexibleTime(createdAt)
	return &rel, nil
}

func (r *SQLiteRelationshipReader) ListByRepository(ctx context.Context, repositoryID string) ([]*memorymodels.Relationship, error) {
	query := `
		SELECT id, repository_id, from_id, to_id, relationship_type, created_at
		FROM memory_relationships
		WHERE repository_id = ?
	`
	rows, err := r.db.QueryContext(ctx, query, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list relationships by repository: %w", err)
	}
	defer rows.Close()

	var relationships []*memorymodels.Relationship
	for rows.Next() {
		var rel memorymodels.Relationship
		var createdAt sqlite.FlexibleTime
		err := rows.Scan(
			&rel.ID,
			&rel.RepositoryID,
			&rel.FromID,
			&rel.ToID,
			&rel.Type,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan relationship: %w", err)
		}
		rel.CreatedAt = flexibleTime(createdAt)
		relationships = append(relationships, &rel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return relationships, nil
}

func (r *SQLiteRelationshipReader) ListAll(ctx context.Context) ([]*memorymodels.Relationship, error) {
	query := `
		SELECT id, repository_id, from_id, to_id, relationship_type, created_at
		FROM memory_relationships
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all relationships: %w", err)
	}
	defer rows.Close()

	var relationships []*memorymodels.Relationship
	for rows.Next() {
		var rel memorymodels.Relationship
		var createdAt sqlite.FlexibleTime
		err := rows.Scan(
			&rel.ID,
			&rel.RepositoryID,
			&rel.FromID,
			&rel.ToID,
			&rel.Type,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan relationship: %w", err)
		}
		rel.CreatedAt = flexibleTime(createdAt)
		relationships = append(relationships, &rel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return relationships, nil
}

// SQLiteContributorReader implements ContributorReader for SQLite.
type SQLiteContributorReader struct {
	db *sqlite.Database
}

func NewSQLiteContributorReader(db *sqlite.Database) *SQLiteContributorReader {
	return &SQLiteContributorReader{db: db}
}

func (r *SQLiteContributorReader) GetByID(ctx context.Context, repositoryID string, id string) (*models.Contributor, error) {
	var c models.Contributor
	var firstSeen, lastSeen sqlite.FlexibleTime
	query := `
		SELECT id, repository_id, name, email, first_seen, last_seen, commit_count
		FROM contributors
		WHERE repository_id = ? AND id = ?
	`
	err := r.db.QueryRowContext(ctx, query, repositoryID, id).Scan(
		&c.ID,
		&c.RepositoryID,
		&c.Name,
		&c.Email,
		&firstSeen,
		&lastSeen,
		&c.CommitCount,
	)
	if err != nil {
		return nil, err
	}
	c.FirstSeen = flexibleTime(firstSeen)
	c.LastSeen = flexibleTime(lastSeen)
	return &c, nil
}

func (r *SQLiteContributorReader) ListByRepository(ctx context.Context, repositoryID string) ([]*models.Contributor, error) {
	query := `
		SELECT id, repository_id, name, email, first_seen, last_seen, commit_count
		FROM contributors
		WHERE repository_id = ?
	`
	rows, err := r.db.QueryContext(ctx, query, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list contributors: %w", err)
	}
	defer rows.Close()

	var list []*models.Contributor
	for rows.Next() {
		var c models.Contributor
		var firstSeen, lastSeen sqlite.FlexibleTime
		err := rows.Scan(
			&c.ID,
			&c.RepositoryID,
			&c.Name,
			&c.Email,
			&firstSeen,
			&lastSeen,
			&c.CommitCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contributor: %w", err)
		}
		c.FirstSeen = flexibleTime(firstSeen)
		c.LastSeen = flexibleTime(lastSeen)
		list = append(list, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return list, nil
}

// SQLiteExpertiseReader implements ExpertiseReader for SQLite.
type SQLiteExpertiseReader struct {
	db *sqlite.Database
}

func NewSQLiteExpertiseReader(db *sqlite.Database) *SQLiteExpertiseReader {
	return &SQLiteExpertiseReader{db: db}
}

func (r *SQLiteExpertiseReader) ListByRepository(ctx context.Context, repositoryID string) ([]*models.Expertise, error) {
	query := `
		SELECT id, repository_id, contributor_id, domain, score, COALESCE(evidence_json, '')
		FROM expertise
		WHERE repository_id = ?
	`
	rows, err := r.db.QueryContext(ctx, query, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list expertise: %w", err)
	}
	defer rows.Close()

	var list []*models.Expertise
	for rows.Next() {
		var e models.Expertise
		err := rows.Scan(
			&e.ID,
			&e.RepositoryID,
			&e.ContributorID,
			&e.Domain,
			&e.Score,
			&e.EvidenceJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expertise: %w", err)
		}
		list = append(list, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return list, nil
}

func (r *SQLiteExpertiseReader) ListByContributor(ctx context.Context, repositoryID string, contributorID string) ([]*models.Expertise, error) {
	query := `
		SELECT id, repository_id, contributor_id, domain, score, COALESCE(evidence_json, '')
		FROM expertise
		WHERE repository_id = ? AND contributor_id = ?
	`
	rows, err := r.db.QueryContext(ctx, query, repositoryID, contributorID)
	if err != nil {
		return nil, fmt.Errorf("failed to list contributor expertise: %w", err)
	}
	defer rows.Close()

	var list []*models.Expertise
	for rows.Next() {
		var e models.Expertise
		err := rows.Scan(
			&e.ID,
			&e.RepositoryID,
			&e.ContributorID,
			&e.Domain,
			&e.Score,
			&e.EvidenceJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expertise: %w", err)
		}
		list = append(list, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return list, nil
}

// SQLiteSourceReader implements SourceReader for SQLite.
type SQLiteSourceReader struct {
	db *sqlite.Database
}

func NewSQLiteSourceReader(db *sqlite.Database) *SQLiteSourceReader {
	return &SQLiteSourceReader{db: db}
}

func (r *SQLiteSourceReader) ListByRepository(ctx context.Context, repositoryID string) ([]*models.Source, error) {
	query := `
		SELECT id, repository_id, source_type, reference, title, author, timestamp, COALESCE(metadata_json, '')
		FROM sources
		WHERE repository_id = ?
	`
	rows, err := r.db.QueryContext(ctx, query, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sources: %w", err)
	}
	defer rows.Close()

	var list []*models.Source
	for rows.Next() {
		var src models.Source
		var ts sqlite.FlexibleTime
		err := rows.Scan(
			&src.ID,
			&src.RepositoryID,
			&src.SourceType,
			&src.Reference,
			&src.Title,
			&src.Author,
			&ts,
			&src.MetadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan source: %w", err)
		}
		src.Timestamp = ts.Time
		list = append(list, &src)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return list, nil
}
