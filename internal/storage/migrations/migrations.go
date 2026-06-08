package migrations

import (
	"database/sql"
	"fmt"
	"sort"

	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// Migration represents a versioned database schema migration.
type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

// allMigrations is the registry of all database migrations.
// Version 1 sets up the complete database schema defined in the RepoNerve Data Model.
var allMigrations = []Migration{
	{
		Version: 1,
		Name:    "create_initial_schema",
		Up: `
			CREATE TABLE IF NOT EXISTS repositories (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				path TEXT NOT NULL,
				default_branch TEXT,
				created_at DATETIME NOT NULL,
				updated_at DATETIME NOT NULL
			);

			CREATE TABLE IF NOT EXISTS sources (
				id TEXT PRIMARY KEY,
				repository_id TEXT NOT NULL,
				source_type TEXT NOT NULL,
				reference TEXT NOT NULL,
				title TEXT,
				author TEXT,
				timestamp DATETIME,
				metadata_json TEXT,
				created_at DATETIME NOT NULL,
				FOREIGN KEY (repository_id) REFERENCES repositories(id)
			);

			CREATE TABLE IF NOT EXISTS memories (
				id TEXT PRIMARY KEY,
				repository_id TEXT NOT NULL,
				memory_type TEXT NOT NULL,
				title TEXT NOT NULL,
				summary TEXT,
				confidence TEXT NOT NULL,
				metadata_json TEXT,
				created_at DATETIME NOT NULL,
				updated_at DATETIME NOT NULL,
				FOREIGN KEY (repository_id) REFERENCES repositories(id)
			);

			CREATE TABLE IF NOT EXISTS facts (
				memory_id TEXT PRIMARY KEY,
				subject TEXT NOT NULL,
				predicate TEXT NOT NULL,
				object TEXT NOT NULL,
				FOREIGN KEY (memory_id) REFERENCES memories(id)
			);

			CREATE TABLE IF NOT EXISTS events (
				memory_id TEXT PRIMARY KEY,
				event_type TEXT NOT NULL,
				event_timestamp DATETIME,
				actor TEXT,
				resource TEXT,
				FOREIGN KEY (memory_id) REFERENCES memories(id)
			);

			CREATE TABLE IF NOT EXISTS decisions (
				memory_id TEXT PRIMARY KEY,
				reason TEXT,
				alternatives TEXT,
				tradeoffs TEXT,
				decision_maker TEXT,
				outcome TEXT,
				FOREIGN KEY (memory_id) REFERENCES memories(id)
			);

			CREATE TABLE IF NOT EXISTS ownerships (
				memory_id TEXT PRIMARY KEY,
				resource TEXT NOT NULL,
				owner_type TEXT NOT NULL,
				owner TEXT NOT NULL,
				start_date DATETIME,
				end_date DATETIME,
				FOREIGN KEY (memory_id) REFERENCES memories(id)
			);

			CREATE TABLE IF NOT EXISTS intents (
				memory_id TEXT PRIMARY KEY,
				goal TEXT NOT NULL,
				description TEXT,
				outcome TEXT,
				FOREIGN KEY (memory_id) REFERENCES memories(id)
			);

			CREATE TABLE IF NOT EXISTS relationships (
				id TEXT PRIMARY KEY,
				source_memory_id TEXT NOT NULL,
				relation TEXT NOT NULL,
				target_memory_id TEXT NOT NULL,
				confidence TEXT,
				FOREIGN KEY (source_memory_id) REFERENCES memories(id),
				FOREIGN KEY (target_memory_id) REFERENCES memories(id)
			);

			CREATE TABLE IF NOT EXISTS evidence (
				id TEXT PRIMARY KEY,
				memory_id TEXT NOT NULL,
				source_id TEXT NOT NULL,
				confidence TEXT,
				explanation TEXT,
				FOREIGN KEY (memory_id) REFERENCES memories(id),
				FOREIGN KEY (source_id) REFERENCES sources(id)
			);

			CREATE VIRTUAL TABLE IF NOT EXISTS memory_search USING fts5 (
				memory_id,
				title,
				summary,
				content
			);

			CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(memory_type);
			CREATE INDEX IF NOT EXISTS idx_sources_type ON sources(source_type);
			CREATE INDEX IF NOT EXISTS idx_evidence_memory ON evidence(memory_id);
			CREATE INDEX IF NOT EXISTS idx_relationships_source ON relationships(source_memory_id);
			CREATE INDEX IF NOT EXISTS idx_relationships_target ON relationships(target_memory_id);
		`,
		Down: `
			DROP INDEX IF EXISTS idx_relationships_target;
			DROP INDEX IF EXISTS idx_relationships_source;
			DROP INDEX IF EXISTS idx_evidence_memory;
			DROP INDEX IF EXISTS idx_sources_type;
			DROP INDEX IF EXISTS idx_memories_type;
			DROP TABLE IF EXISTS memory_search;
			DROP TABLE IF EXISTS evidence;
			DROP TABLE IF EXISTS relationships;
			DROP TABLE IF EXISTS intents;
			DROP TABLE IF EXISTS ownerships;
			DROP TABLE IF EXISTS decisions;
			DROP TABLE IF EXISTS events;
			DROP TABLE IF EXISTS facts;
			DROP TABLE IF EXISTS memories;
			DROP TABLE IF EXISTS sources;
			DROP TABLE IF EXISTS repositories;
		`,
	},
	{
		Version: 2,
		Name:    "add_scan_state_table",
		Up: `
			CREATE TABLE IF NOT EXISTS scan_state (
				repository_id TEXT PRIMARY KEY,
				last_scan_commit TEXT NOT NULL,
				updated_at DATETIME NOT NULL,
				FOREIGN KEY (repository_id) REFERENCES repositories(id)
			);
		`,
		Down: `
			DROP TABLE IF EXISTS scan_state;
		`,
	},
	{
		Version: 3,
		Name:    "create_memory_events_table",
		Up: `
			CREATE TABLE IF NOT EXISTS memory_events (
				id TEXT PRIMARY KEY,
				repository_id TEXT NOT NULL,
				event_type TEXT NOT NULL,
				title TEXT NOT NULL,
				description TEXT,
				source_id TEXT NOT NULL,
				timestamp DATETIME NOT NULL,
				created_at DATETIME NOT NULL,
				FOREIGN KEY (repository_id) REFERENCES repositories(id),
				FOREIGN KEY (source_id) REFERENCES sources(id)
			);

			CREATE INDEX IF NOT EXISTS idx_memory_events_repository_id
			ON memory_events(repository_id);

			CREATE INDEX IF NOT EXISTS idx_memory_events_source_id
			ON memory_events(source_id);
		`,
		Down: `
			DROP INDEX IF EXISTS idx_memory_events_source_id;
			DROP INDEX IF EXISTS idx_memory_events_repository_id;
			DROP TABLE IF EXISTS memory_events;
		`,
	},
	{
		Version: 4,
		Name:    "create_memory_decisions_table",
		Up: `
			CREATE TABLE IF NOT EXISTS memory_decisions (
				id TEXT PRIMARY KEY,
				repository_id TEXT NOT NULL,
				title TEXT NOT NULL,
				status TEXT NOT NULL,
				source_id TEXT NOT NULL,
				created_at DATETIME NOT NULL,
				FOREIGN KEY (repository_id) REFERENCES repositories(id),
				FOREIGN KEY (source_id) REFERENCES sources(id)
			);

			CREATE INDEX IF NOT EXISTS idx_memory_decisions_repository_id
			ON memory_decisions(repository_id);

			CREATE INDEX IF NOT EXISTS idx_memory_decisions_source_id
			ON memory_decisions(source_id);
		`,
		Down: `
			DROP INDEX IF EXISTS idx_memory_decisions_source_id;
			DROP INDEX IF EXISTS idx_memory_decisions_repository_id;
			DROP TABLE IF EXISTS memory_decisions;
		`,
	},
	{
		Version: 5,
		Name:    "create_memory_intents_table",
		Up: `
			CREATE TABLE IF NOT EXISTS memory_intents (
				id TEXT PRIMARY KEY,
				repository_id TEXT NOT NULL,
				description TEXT NOT NULL,
				source_id TEXT NOT NULL,
				created_at DATETIME NOT NULL,
				FOREIGN KEY (repository_id) REFERENCES repositories(id),
				FOREIGN KEY (source_id) REFERENCES sources(id)
			);

			CREATE INDEX IF NOT EXISTS idx_memory_intents_repository_id
			ON memory_intents(repository_id);

			CREATE INDEX IF NOT EXISTS idx_memory_intents_source_id
			ON memory_intents(source_id);
		`,
		Down: `
			DROP INDEX IF EXISTS idx_memory_intents_source_id;
			DROP INDEX IF EXISTS idx_memory_intents_repository_id;
			DROP TABLE IF EXISTS memory_intents;
		`,
	},
	{
		Version: 6,
		Name:    "create_memory_facts_table",
		Up: `
			CREATE TABLE IF NOT EXISTS memory_facts (
				id TEXT PRIMARY KEY,
				repository_id TEXT NOT NULL,
				subject TEXT NOT NULL,
				predicate TEXT NOT NULL,
				object TEXT NOT NULL,
				source_id TEXT NOT NULL,
				created_at DATETIME NOT NULL,
				FOREIGN KEY (repository_id) REFERENCES repositories(id),
				FOREIGN KEY (source_id) REFERENCES sources(id)
			);

			CREATE INDEX IF NOT EXISTS idx_memory_facts_repository_id
			ON memory_facts(repository_id);

			CREATE INDEX IF NOT EXISTS idx_memory_facts_source_id
			ON memory_facts(source_id);
		`,
		Down: `
			DROP INDEX IF EXISTS idx_memory_facts_source_id;
			DROP INDEX IF EXISTS idx_memory_facts_repository_id;
			DROP TABLE IF EXISTS memory_facts;
		`,
	},
	{
		Version: 7,
		Name:    "create_memory_relationships_table",
		Up: `
			CREATE TABLE IF NOT EXISTS memory_relationships (
				id TEXT PRIMARY KEY,
				repository_id TEXT NOT NULL,
				from_id TEXT NOT NULL,
				to_id TEXT NOT NULL,
				relationship_type TEXT NOT NULL,
				created_at DATETIME NOT NULL,
				FOREIGN KEY (repository_id) REFERENCES repositories(id)
			);

			CREATE INDEX IF NOT EXISTS idx_memory_relationships_repository_id
			ON memory_relationships(repository_id);

			CREATE INDEX IF NOT EXISTS idx_memory_relationships_from_id
			ON memory_relationships(from_id);

			CREATE INDEX IF NOT EXISTS idx_memory_relationships_to_id
			ON memory_relationships(to_id);
		`,
		Down: `
			DROP INDEX IF EXISTS idx_memory_relationships_to_id;
			DROP INDEX IF EXISTS idx_memory_relationships_from_id;
			DROP INDEX IF EXISTS idx_memory_relationships_repository_id;
			DROP TABLE IF EXISTS memory_relationships;
		`,
	},
	{
		Version: 8,
		Name:    "create_ownership_tables",
		Up: `
			CREATE TABLE IF NOT EXISTS contributors (
				id TEXT PRIMARY KEY,
				repository_id TEXT NOT NULL,
				name TEXT NOT NULL,
				email TEXT NOT NULL,
				first_seen DATETIME NOT NULL,
				last_seen DATETIME NOT NULL,
				commit_count INTEGER NOT NULL,
				FOREIGN KEY (repository_id) REFERENCES repositories(id)
			);

			CREATE INDEX IF NOT EXISTS idx_contributors_repository_id ON contributors(repository_id);
			CREATE INDEX IF NOT EXISTS idx_contributors_email ON contributors(email);

			CREATE TABLE IF NOT EXISTS expertise (
				id TEXT PRIMARY KEY,
				repository_id TEXT NOT NULL,
				contributor_id TEXT NOT NULL,
				domain TEXT NOT NULL,
				score REAL NOT NULL,
				evidence_json TEXT,
				FOREIGN KEY (repository_id) REFERENCES repositories(id),
				FOREIGN KEY (contributor_id) REFERENCES contributors(id)
			);

			CREATE INDEX IF NOT EXISTS idx_expertise_repository_id ON expertise(repository_id);
			CREATE INDEX IF NOT EXISTS idx_expertise_contributor_id ON expertise(contributor_id);
		`,
		Down: `
			DROP INDEX IF EXISTS idx_expertise_contributor_id;
			DROP INDEX IF EXISTS idx_expertise_repository_id;
			DROP TABLE IF EXISTS expertise;
			DROP INDEX IF EXISTS idx_contributors_email;
			DROP INDEX IF EXISTS idx_contributors_repository_id;
			DROP TABLE IF EXISTS contributors;
		`,
	},
}

// GetAppliedVersions returns the list of applied migration versions from the database.
func GetAppliedVersions(db *sqlite.Database) (map[int]bool, error) {
	applied := make(map[int]bool)

	var name string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&name)
	if err == sql.ErrNoRows {
		return applied, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to check schema_migrations existence: %w", err)
	}

	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[v] = true
	}

	return applied, nil
}

// RunUp executes all pending migrations in order of version.
func RunUp(db *sqlite.Database) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit schema_migrations setup: %w", err)
	}

	applied, err := GetAppliedVersions(db)
	if err != nil {
		return err
	}

	var pending []Migration
	for _, m := range allMigrations {
		if !applied[m.Version] {
			pending = append(pending, m)
		}
	}

	sort.Slice(pending, func(i, j int) bool {
		return pending[i].Version < pending[j].Version
	})

	for _, m := range pending {
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for version %d: %w", m.Version, err)
		}
		defer tx.Rollback()

		if _, err := tx.Exec(m.Up); err != nil {
			return fmt.Errorf("failed to execute migration up for version %d (%s): %w", m.Version, m.Name, err)
		}

		_, err = tx.Exec("INSERT INTO schema_migrations (version, name) VALUES (?, ?)", m.Version, m.Name)
		if err != nil {
			return fmt.Errorf("failed to record migration application for version %d: %w", m.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration transaction for version %d: %w", m.Version, err)
		}
	}

	return nil
}

// Rollback executes the Down SQL of the highest version applied migration.
func Rollback(db *sqlite.Database) error {
	applied, err := GetAppliedVersions(db)
	if err != nil {
		return err
	}

	if len(applied) == 0 {
		return fmt.Errorf("no migrations applied to rollback")
	}

	var versions []int
	for v := range applied {
		versions = append(versions, v)
	}
	sort.Ints(versions)
	highestVersion := versions[len(versions)-1]

	var targetMigration *Migration
	for _, m := range allMigrations {
		if m.Version == highestVersion {
			targetMigration = &m
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration specification for version %d not found in registry", highestVersion)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for rollback: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(targetMigration.Down); err != nil {
		return fmt.Errorf("failed to execute rollback for version %d (%s): %w", targetMigration.Version, targetMigration.Name, err)
	}

	_, err = tx.Exec("DELETE FROM schema_migrations WHERE version = ?", targetMigration.Version)
	if err != nil {
		return fmt.Errorf("failed to delete migration record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback transaction: %w", err)
	}

	return nil
}
