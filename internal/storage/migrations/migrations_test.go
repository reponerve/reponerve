package migrations

import (
	"os"
	"path/filepath"
	"testing"

	"reponerve/internal/storage/sqlite"
)

func TestMigrations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-migrations-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	applied, err := GetAppliedVersions(db)
	if err != nil {
		t.Fatalf("failed to get applied versions: %v", err)
	}
	if len(applied) != 0 {
		t.Errorf("expected 0 migrations applied, got %d", len(applied))
	}

	err = RunUp(db)
	if err != nil {
		t.Fatalf("failed to run migrations up: %v", err)
	}

	applied, err = GetAppliedVersions(db)
	if err != nil {
		t.Fatalf("failed to get applied versions after RunUp: %v", err)
	}
	if !applied[1] {
		t.Errorf("expected migration version 1 to be applied")
	}

	tables := []string{
		"schema_migrations",
		"repositories",
		"sources",
		"memories",
		"facts",
		"events",
		"decisions",
		"ownerships",
		"intents",
		"relationships",
		"evidence",
		"memory_search",
	}
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("expected table %q to exist, got error: %v", table, err)
		}
	}

	err = RunUp(db)
	if err != nil {
		t.Fatalf("failed to re-run migrations up: %v", err)
	}

	err = Rollback(db)
	if err != nil {
		t.Fatalf("failed to rollback migration: %v", err)
	}

	applied, err = GetAppliedVersions(db)
	if err != nil {
		t.Fatalf("failed to get applied versions after rollback: %v", err)
	}
	if applied[1] {
		t.Errorf("expected migration version 1 to be rolled back")
	}

	for _, table := range []string{
		"repositories",
		"sources",
		"memories",
		"facts",
		"events",
		"decisions",
		"ownerships",
		"intents",
		"relationships",
		"evidence",
		"memory_search",
	} {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err == nil {
			t.Errorf("expected table %q to be dropped after rollback, but it still exists", table)
		}
	}
}
