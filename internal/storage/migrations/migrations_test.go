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
	for v := 1; v <= 8; v++ {
		if !applied[v] {
			t.Errorf("expected migration version %d to be applied", v)
		}
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
		"scan_state",
		"memory_events",
		"memory_decisions",
		"memory_intents",
		"memory_facts",
		"memory_relationships",
		"contributors",
		"expertise",
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

	// First Rollback (rolls back version 8: create_ownership_tables)
	err = Rollback(db)
	if err != nil {
		t.Fatalf("failed to rollback migration version 8: %v", err)
	}

	applied, err = GetAppliedVersions(db)
	if err != nil {
		t.Fatalf("failed to get applied versions after first rollback: %v", err)
	}
	if applied[8] {
		t.Errorf("expected migration version 8 to be rolled back")
	}
	for v := 1; v <= 7; v++ {
		if !applied[v] {
			t.Errorf("expected migration version %d to still be applied", v)
		}
	}

	// Verify contributors and expertise tables are dropped, but memory_relationships still exists
	var name string
	for _, table := range []string{"contributors", "expertise"} {
		err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err == nil {
			t.Errorf("expected table %q to be dropped after first rollback, but it still exists", table)
		}
	}
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='memory_relationships'").Scan(&name)
	if err != nil {
		t.Error("expected table 'memory_relationships' to still exist after first rollback")
	}

	// Second Rollback (rolls back version 7: memory_relationships)
	err = Rollback(db)
	if err != nil {
		t.Fatalf("failed to rollback migration version 7: %v", err)
	}

	applied, err = GetAppliedVersions(db)
	if err != nil {
		t.Fatalf("failed to get applied versions after second rollback: %v", err)
	}
	if applied[7] {
		t.Errorf("expected migration version 7 to be rolled back")
	}
	for v := 1; v <= 6; v++ {
		if !applied[v] {
			t.Errorf("expected migration version %d to still be applied", v)
		}
	}

	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='memory_relationships'").Scan(&name)
	if err == nil {
		t.Error("expected table 'memory_relationships' to be dropped after second rollback, but it still exists")
	}
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='memory_facts'").Scan(&name)
	if err != nil {
		t.Error("expected table 'memory_facts' to still exist after second rollback")
	}

	// Third Rollback (rolls back version 6: memory_facts)
	err = Rollback(db)
	if err != nil {
		t.Fatalf("failed to rollback migration version 6: %v", err)
	}

	applied, err = GetAppliedVersions(db)
	if err != nil {
		t.Fatalf("failed to get applied versions after third rollback: %v", err)
	}
	if applied[6] {
		t.Errorf("expected migration version 6 to be rolled back")
	}
	for v := 1; v <= 5; v++ {
		if !applied[v] {
			t.Errorf("expected migration version %d to still be applied", v)
		}
	}

	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='memory_facts'").Scan(&name)
	if err == nil {
		t.Error("expected table 'memory_facts' to be dropped after third rollback, but it still exists")
	}
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='memory_intents'").Scan(&name)
	if err != nil {
		t.Error("expected table 'memory_intents' to still exist after third rollback")
	}

	// Fourth Rollback (rolls back version 5: memory_intents)
	err = Rollback(db)
	if err != nil {
		t.Fatalf("failed to rollback migration version 5: %v", err)
	}

	applied, err = GetAppliedVersions(db)
	if err != nil {
		t.Fatalf("failed to get applied versions after fourth rollback: %v", err)
	}
	if applied[5] {
		t.Errorf("expected migration version 5 to be rolled back")
	}
	for v := 1; v <= 4; v++ {
		if !applied[v] {
			t.Errorf("expected migration version %d to still be applied", v)
		}
	}

	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='memory_intents'").Scan(&name)
	if err == nil {
		t.Error("expected table 'memory_intents' to be dropped after fourth rollback, but it still exists")
	}
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='memory_decisions'").Scan(&name)
	if err != nil {
		t.Error("expected table 'memory_decisions' to still exist after fourth rollback")
	}

	// Fifth Rollback (rolls back version 4: memory_decisions)
	err = Rollback(db)
	if err != nil {
		t.Fatalf("failed to rollback migration version 4: %v", err)
	}

	applied, err = GetAppliedVersions(db)
	if err != nil {
		t.Fatalf("failed to get applied versions after fifth rollback: %v", err)
	}
	if applied[4] {
		t.Errorf("expected migration version 4 to be rolled back")
	}
	for v := 1; v <= 3; v++ {
		if !applied[v] {
			t.Errorf("expected migration version %d to still be applied", v)
		}
	}

	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='memory_decisions'").Scan(&name)
	if err == nil {
		t.Error("expected table 'memory_decisions' to be dropped after fifth rollback, but it still exists")
	}
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='memory_events'").Scan(&name)
	if err != nil {
		t.Error("expected table 'memory_events' to still exist after fifth rollback")
	}

	// Sixth Rollback (rolls back version 3: memory_events)
	err = Rollback(db)
	if err != nil {
		t.Fatalf("failed to rollback migration version 3: %v", err)
	}

	applied, err = GetAppliedVersions(db)
	if err != nil {
		t.Fatalf("failed to get applied versions after sixth rollback: %v", err)
	}
	if applied[3] {
		t.Errorf("expected migration version 3 to be rolled back")
	}
	for v := 1; v <= 2; v++ {
		if !applied[v] {
			t.Errorf("expected migration version %d to still be applied", v)
		}
	}

	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='memory_events'").Scan(&name)
	if err == nil {
		t.Error("expected table 'memory_events' to be dropped after sixth rollback, but it still exists")
	}
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='scan_state'").Scan(&name)
	if err != nil {
		t.Error("expected table 'scan_state' to still exist after sixth rollback")
	}

	// Seventh Rollback (rolls back version 2: scan_state)
	err = Rollback(db)
	if err != nil {
		t.Fatalf("failed to rollback migration version 2: %v", err)
	}

	applied, err = GetAppliedVersions(db)
	if err != nil {
		t.Fatalf("failed to get applied versions after seventh rollback: %v", err)
	}
	if applied[2] {
		t.Errorf("expected migration version 2 to be rolled back")
	}
	if !applied[1] {
		t.Errorf("expected migration version 1 to still be applied")
	}

	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='scan_state'").Scan(&name)
	if err == nil {
		t.Error("expected table 'scan_state' to be dropped after seventh rollback, but it still exists")
	}
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='repositories'").Scan(&name)
	if err != nil {
		t.Error("expected table 'repositories' to still exist after seventh rollback")
	}

	// Eighth Rollback (rolls back version 1: initial tables)
	err = Rollback(db)
	if err != nil {
		t.Fatalf("failed to rollback migration version 1: %v", err)
	}

	applied, err = GetAppliedVersions(db)
	if err != nil {
		t.Fatalf("failed to get applied versions after eighth rollback: %v", err)
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
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err == nil {
			t.Errorf("expected table %q to be dropped after eighth rollback, but it still exists", table)
		}
	}
}
