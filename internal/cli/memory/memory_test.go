package memorycmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"reponerve/internal/config"
	memorymodels "reponerve/internal/memory/models"
	memorystorage "reponerve/internal/memory/storage"
	"reponerve/internal/storage/migrations"
	"reponerve/internal/storage/sqlite"
	models "reponerve/pkg/models"
)

func TestMemoryListCommands(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-memory-cmd-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origWorkspace := os.Getenv("REPONERVE_WORKSPACE")
	defer func() {
		if origWorkspace != "" {
			os.Setenv("REPONERVE_WORKSPACE", origWorkspace)
		} else {
			os.Unsetenv("REPONERVE_WORKSPACE")
		}
	}()

	workspacePath := filepath.Join(tempDir, ".reponerve")
	os.Setenv("REPONERVE_WORKSPACE", workspacePath)

	cfg, err := config.Initialize(workspacePath)
	if err != nil {
		t.Fatalf("failed to initialize config: %v", err)
	}

	db, err := sqlite.Open(cfg.Storage.SQLitePath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := migrations.RunUp(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	ctx := context.Background()
	repoID1 := "repo_1"
	repoID2 := "repo_2"

	// Insert repositories
	_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repoID1, "Repo 1", tempDir, "main")
	if err != nil {
		t.Fatalf("failed to insert repository 1: %v", err)
	}
	_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repoID2, "Repo 2", tempDir, "main")
	if err != nil {
		t.Fatalf("failed to insert repository 2: %v", err)
	}

	// Insert sources
	_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_1", repoID1, "adr", "docs/adr/0001.md", "Author", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("failed to insert source 1: %v", err)
	}
	_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_2", repoID2, "commit", "commit_1", "Author", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("failed to insert source 2: %v", err)
	}

	// Insert event
	eventStore := sqlite.NewEventStore(db)
	err = eventStore.UpsertEvent(ctx, &models.Event{
		ID:           "evt_1",
		RepositoryID: repoID1,
		EventType:    "FEATURE",
		Title:        "Awesome Feature",
		Description:  "Description of feature",
		SourceID:     "src_1",
		Timestamp:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	// Insert decision
	decisionStore := memorystorage.NewSQLiteDecisionStore(db)
	err = decisionStore.UpsertDecision(ctx, &memorymodels.Decision{
		ID:           "dec_1",
		RepositoryID: repoID1,
		Title:        "Use Cache",
		Status:       "Accepted",
		SourceID:     "src_1",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to insert decision: %v", err)
	}

	// Insert intent
	intentStore := memorystorage.NewSQLiteIntentStore(db)
	err = intentStore.UpsertIntent(ctx, &memorymodels.Intent{
		ID:           "int_1",
		RepositoryID: repoID1,
		Description:  "Optimize latency",
		SourceID:     "src_1",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to insert intent: %v", err)
	}

	// Insert fact
	factStore := memorystorage.NewSQLiteFactStore(db)
	err = factStore.UpsertFact(ctx, &memorymodels.Fact{
		ID:           "fact_1",
		RepositoryID: repoID1,
		Subject:      "App",
		Predicate:    "uses",
		Object:       "Redis",
		SourceID:     "src_1",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to insert fact: %v", err)
	}

	// Insert relationship
	relStore := memorystorage.NewSQLiteRelationshipStore(db)
	err = relStore.UpsertRelationship(ctx, &memorymodels.Relationship{
		ID:           "rel_1",
		RepositoryID: repoID1,
		FromID:       "int_1",
		ToID:         "dec_1",
		Type:         "DRIVES",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to insert relationship: %v", err)
	}

	execute := func(args ...string) (string, error) {
		cmd := NewCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs(args)
		err := cmd.Execute()
		return buf.String(), err
	}

	t.Run("Command registration", func(t *testing.T) {
		out, err := execute("--help")
		if err != nil {
			t.Fatalf("failed to execute command help: %v", err)
		}
		expectedSubstrings := []string{"list", "memory"}
		for _, s := range expectedSubstrings {
			if !strings.Contains(out, s) {
				t.Errorf("expected help to contain %q, got %q", s, out)
			}
		}
	})

	t.Run("List all events", func(t *testing.T) {
		out, err := execute("list", "events")
		if err != nil {
			t.Fatalf("failed to list events: %v", err)
		}
		if !strings.Contains(out, "evt_1") || !strings.Contains(out, "Awesome Feature") {
			t.Errorf("unexpected output: %s", out)
		}
	})

	t.Run("List events by repository", func(t *testing.T) {
		out, err := execute("list", "events", "--repository", repoID1)
		if err != nil {
			t.Fatalf("failed to list events for repo1: %v", err)
		}
		if !strings.Contains(out, "evt_1") {
			t.Errorf("unexpected output: %s", out)
		}

		out2, err := execute("list", "events", "--repository", repoID2)
		if err != nil {
			t.Fatalf("failed to list events for repo2: %v", err)
		}
		if !strings.Contains(out2, "No records found.") {
			t.Errorf("expected No records found, got %s", out2)
		}
	})

	t.Run("List decisions", func(t *testing.T) {
		out, err := execute("list", "decisions")
		if err != nil {
			t.Fatalf("failed to list decisions: %v", err)
		}
		if !strings.Contains(out, "dec_1") || !strings.Contains(out, "Use Cache") {
			t.Errorf("unexpected output: %s", out)
		}
	})

	t.Run("List intents", func(t *testing.T) {
		out, err := execute("list", "intents")
		if err != nil {
			t.Fatalf("failed to list intents: %v", err)
		}
		if !strings.Contains(out, "int_1") || !strings.Contains(out, "Optimize latency") {
			t.Errorf("unexpected output: %s", out)
		}
	})

	t.Run("List facts", func(t *testing.T) {
		out, err := execute("list", "facts")
		if err != nil {
			t.Fatalf("failed to list facts: %v", err)
		}
		if !strings.Contains(out, "fact_1") || !strings.Contains(out, "App") || !strings.Contains(out, "uses") {
			t.Errorf("unexpected output: %s", out)
		}
	})

	t.Run("List relationships", func(t *testing.T) {
		out, err := execute("list", "relationships")
		if err != nil {
			t.Fatalf("failed to list relationships: %v", err)
		}
		if !strings.Contains(out, "rel_1") || !strings.Contains(out, "DRIVES") {
			t.Errorf("unexpected output: %s", out)
		}
	})

	t.Run("Empty result sets when repository doesn't match", func(t *testing.T) {
		out, err := execute("list", "decisions", "--repository", repoID2)
		if err != nil {
			t.Fatalf("failed to list decisions: %v", err)
		}
		if !strings.Contains(out, "No records found.") {
			t.Errorf("expected No records found, got %s", out)
		}
	})
}

func TestMemoryListCommandsUninitialized(t *testing.T) {
	origWorkspace := os.Getenv("REPONERVE_WORKSPACE")
	os.Setenv("REPONERVE_WORKSPACE", "/non-existent-dir-for-test")
	defer func() {
		if origWorkspace != "" {
			os.Setenv("REPONERVE_WORKSPACE", origWorkspace)
		} else {
			os.Unsetenv("REPONERVE_WORKSPACE")
		}
	}()

	cmd := NewCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"list", "events"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when workspace is uninitialized, got nil")
	}
	if !strings.Contains(err.Error(), "workspace not initialized") {
		t.Errorf("expected 'workspace not initialized' error, got %v", err)
	}
}
