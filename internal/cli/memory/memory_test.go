package memorycmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/reponerve/reponerve/internal/config"
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	memorystorage "github.com/reponerve/reponerve/internal/memory/storage"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
	models "github.com/reponerve/reponerve/pkg/models"
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

	// Insert relationships
	relStore := memorystorage.NewSQLiteRelationshipStore(db)
	err = relStore.UpsertRelationship(ctx, &memorymodels.Relationship{
		ID:           "rel_1",
		RepositoryID: repoID1,
		FromID:       "int_1",
		ToID:         "dec_1",
		Type:         "INTENT_DRIVES_DECISION",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to insert relationship 1: %v", err)
	}

	err = relStore.UpsertRelationship(ctx, &memorymodels.Relationship{
		ID:           "rel_2",
		RepositoryID: repoID1,
		FromID:       "fact_1",
		ToID:         "dec_1",
		Type:         "FACT_SUPPORTS_DECISION",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to insert relationship 2: %v", err)
	}

	err = relStore.UpsertRelationship(ctx, &memorymodels.Relationship{
		ID:           "rel_3",
		RepositoryID: repoID1,
		FromID:       "dec_1",
		ToID:         "evt_1",
		Type:         "DECISION_RESULTS_IN_EVENT",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to insert relationship 3: %v", err)
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
		if !strings.Contains(out, "rel_1") || !strings.Contains(out, "INTENT_DRIVES_DECISION") {
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

	t.Run("Get commands help", func(t *testing.T) {
		out, err := execute("get", "--help")
		if err != nil {
			t.Fatalf("failed to run get help: %v", err)
		}
		if !strings.Contains(out, "event") || !strings.Contains(out, "decision") || !strings.Contains(out, "intent") || !strings.Contains(out, "fact") {
			t.Errorf("unexpected help output: %s", out)
		}
	})

	t.Run("Get event success", func(t *testing.T) {
		out, err := execute("get", "event", "evt_1")
		if err != nil {
			t.Fatalf("failed to get event: %v", err)
		}
		expectedParts := []string{"Event", "ID:", "evt_1", "Type:", "FEATURE", "Title:", "Awesome Feature", "Description:", "Description of feature", "Source:", "src_1", "Timestamp:"}
		for _, part := range expectedParts {
			if !strings.Contains(out, part) {
				t.Errorf("expected output to contain %q, got:\n%s", part, out)
			}
		}
	})

	t.Run("Get event not found", func(t *testing.T) {
		_, err := execute("get", "event", "non_existent")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "event with ID \"non_existent\" not found") {
			t.Errorf("expected not found error, got: %v", err)
		}
	})

	t.Run("Get decision success", func(t *testing.T) {
		out, err := execute("get", "decision", "dec_1")
		if err != nil {
			t.Fatalf("failed to get decision: %v", err)
		}
		expectedParts := []string{"Decision", "ID:", "dec_1", "Title:", "Use Cache", "Status:", "Accepted", "Source:", "src_1", "Created:"}
		for _, part := range expectedParts {
			if !strings.Contains(out, part) {
				t.Errorf("expected output to contain %q, got:\n%s", part, out)
			}
		}
	})

	t.Run("Get decision not found", func(t *testing.T) {
		_, err := execute("get", "decision", "non_existent")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "decision with ID \"non_existent\" not found") {
			t.Errorf("expected not found error, got: %v", err)
		}
	})

	t.Run("Get intent success", func(t *testing.T) {
		out, err := execute("get", "intent", "int_1")
		if err != nil {
			t.Fatalf("failed to get intent: %v", err)
		}
		expectedParts := []string{"Intent", "ID:", "int_1", "Description:", "Optimize latency", "Source:", "src_1", "Created:"}
		for _, part := range expectedParts {
			if !strings.Contains(out, part) {
				t.Errorf("expected output to contain %q, got:\n%s", part, out)
			}
		}
	})

	t.Run("Get intent not found", func(t *testing.T) {
		_, err := execute("get", "intent", "non_existent")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "intent with ID \"non_existent\" not found") {
			t.Errorf("expected not found error, got: %v", err)
		}
	})

	t.Run("Get fact success", func(t *testing.T) {
		out, err := execute("get", "fact", "fact_1")
		if err != nil {
			t.Fatalf("failed to get fact: %v", err)
		}
		expectedParts := []string{"Fact", "ID:", "fact_1", "Subject:", "App", "Predicate:", "uses", "Object:", "Redis", "Source:", "src_1", "Created:"}
		for _, part := range expectedParts {
			if !strings.Contains(out, part) {
				t.Errorf("expected output to contain %q, got:\n%s", part, out)
			}
		}
	})

	t.Run("Get fact not found", func(t *testing.T) {
		_, err := execute("get", "fact", "non_existent")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "fact with ID \"non_existent\" not found") {
			t.Errorf("expected not found error, got: %v", err)
		}
	})

	t.Run("Get command missing arguments", func(t *testing.T) {
		_, err := execute("get", "event")
		if err == nil {
			t.Fatal("expected error for missing argument, got nil")
		}
	})

	t.Run("Trace command help", func(t *testing.T) {
		out, err := execute("trace", "--help")
		if err != nil {
			t.Fatalf("failed to run trace help: %v", err)
		}
		if !strings.Contains(out, "decision") || !strings.Contains(out, "event") || !strings.Contains(out, "intent") {
			t.Errorf("unexpected help output: %s", out)
		}
	})

	t.Run("Trace decision success", func(t *testing.T) {
		out, err := execute("trace", "decision", "dec_1")
		if err != nil {
			t.Fatalf("failed to run trace decision: %v", err)
		}
		expectedParts := []string{"Decision", "└── Use Cache", "Intent", "└── Optimize latency", "Fact", "└── App uses Redis"}
		for _, part := range expectedParts {
			if !strings.Contains(out, part) {
				t.Errorf("expected output to contain %q, got:\n%s", part, out)
			}
		}
	})

	t.Run("Trace decision not found", func(t *testing.T) {
		_, err := execute("trace", "decision", "non_existent")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "decision with ID \"non_existent\" not found") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("Trace event success", func(t *testing.T) {
		// Event evt_1 is not linked to dec_1 via DECISION_RESULTS_IN_EVENT, but let's test traversal output
		out, err := execute("trace", "event", "evt_1")
		if err != nil {
			t.Fatalf("failed to run trace event: %v", err)
		}
		if !strings.Contains(out, "Event") || !strings.Contains(out, "└── Awesome Feature") {
			t.Errorf("unexpected output:\n%s", out)
		}
	})

	t.Run("Trace event not found", func(t *testing.T) {
		_, err := execute("trace", "event", "non_existent")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "event with ID \"non_existent\" not found") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("Trace intent success", func(t *testing.T) {
		out, err := execute("trace", "intent", "int_1")
		if err != nil {
			t.Fatalf("failed to run trace intent: %v", err)
		}
		expectedParts := []string{"Intent", "└── Optimize latency", "Decision", "└── Use Cache"}
		for _, part := range expectedParts {
			if !strings.Contains(out, part) {
				t.Errorf("expected output to contain %q, got:\n%s", part, out)
			}
		}
	})

	t.Run("Trace intent not found", func(t *testing.T) {
		_, err := execute("trace", "intent", "non_existent")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "intent with ID \"non_existent\" not found") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("Trace command missing arguments", func(t *testing.T) {
		_, err := execute("trace", "decision")
		if err == nil {
			t.Fatal("expected error for missing argument, got nil")
		}
	})

	t.Run("Explain command help", func(t *testing.T) {
		out, err := execute("explain", "--help")
		if err != nil {
			t.Fatalf("failed to run explain help: %v", err)
		}
		if !strings.Contains(out, "decision") || !strings.Contains(out, "event") {
			t.Errorf("unexpected help output: %s", out)
		}
	})

	t.Run("Explain decision success", func(t *testing.T) {
		out, err := execute("explain", "decision", "dec_1")
		if err != nil {
			t.Fatalf("failed to run explain decision: %v", err)
		}
		expectedParts := []string{"Decision:", "Use Cache", "Reason:", "Optimize latency", "Supporting Facts:", "- App uses Redis", "Resulting Events:", "- Awesome Feature"}
		for _, part := range expectedParts {
			if !strings.Contains(out, part) {
				t.Errorf("expected output to contain %q, got:\n%s", part, out)
			}
		}
	})

	t.Run("Explain decision not found", func(t *testing.T) {
		_, err := execute("explain", "decision", "non_existent")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "decision with ID \"non_existent\" not found") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("Explain event success", func(t *testing.T) {
		out, err := execute("explain", "event", "evt_1")
		if err != nil {
			t.Fatalf("failed to run explain event: %v", err)
		}
		expectedParts := []string{"Event:", "Awesome Feature", "Caused By:", "Use Cache", "Reason:", "Optimize latency"}
		for _, part := range expectedParts {
			if !strings.Contains(out, part) {
				t.Errorf("expected output to contain %q, got:\n%s", part, out)
			}
		}
	})

	t.Run("Explain event not found", func(t *testing.T) {
		_, err := execute("explain", "event", "non_existent")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "event with ID \"non_existent\" not found") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("Explain command missing arguments", func(t *testing.T) {
		_, err := execute("explain", "decision")
		if err == nil {
			t.Fatal("expected error for missing argument, got nil")
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
