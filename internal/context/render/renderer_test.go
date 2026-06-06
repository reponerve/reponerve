package render

import (
	stdcontext "context"
	"os"
	"path/filepath"
	"testing"
	"time"

	contextpkg "reponerve/internal/context"
	memorymodels "reponerve/internal/memory/models"
	memorystorage "reponerve/internal/memory/storage"
	"reponerve/internal/query/storage"
	"reponerve/internal/storage/migrations"
	"reponerve/internal/storage/sqlite"
	models "reponerve/pkg/models"
)

func TestRenderer_Unit(t *testing.T) {
	renderer := NewRenderer()
	genTime := time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC)
	repoID := "test_repo"

	t.Run("Empty context", func(t *testing.T) {
		rc := &contextpkg.RepositoryContext{
			RepositoryID: repoID,
			GeneratedAt:  genTime,
		}

		got, err := renderer.Render(rc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := "# Repository Context\n\nRepository: test_repo\n\nGenerated: 2026-06-06T12:00:00Z\n"
		if got != expected {
			t.Errorf("expected:\n%q\ngot:\n%q", expected, got)
		}
	})

	t.Run("Partial context (Decisions and Facts)", func(t *testing.T) {
		rc := &contextpkg.RepositoryContext{
			RepositoryID: repoID,
			GeneratedAt:  genTime,
			Decisions: []*memorymodels.Decision{
				{Title: "Use Redis Cache"},
			},
			Facts: []*memorymodels.Fact{
				{Subject: "Auth Service", Predicate: "USES", Object: "Redis"},
			},
		}

		got, err := renderer.Render(rc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := `# Repository Context

Repository: test_repo

Generated: 2026-06-06T12:00:00Z

## Key Decisions

* Use Redis Cache

## Key Facts

* Auth Service USES Redis
`
		if got != expected {
			t.Errorf("expected:\n%q\ngot:\n%q", expected, got)
		}
	})

	t.Run("Full context", func(t *testing.T) {
		rc := &contextpkg.RepositoryContext{
			RepositoryID: repoID,
			GeneratedAt:  genTime,
			Decisions: []*memorymodels.Decision{
				{Title: "Use Redis Cache"},
				{Title: "Adopt gRPC"},
			},
			Intents: []*memorymodels.Intent{
				{Description: "Reduce Latency"},
				{Description: "Improve Reliability"},
			},
			Facts: []*memorymodels.Fact{
				{Subject: "Auth Service", Predicate: "USES", Object: "Redis"},
				{Subject: "API Gateway", Predicate: "DEPENDS_ON", Object: "Auth Service"},
			},
			Events: []*models.Event{
				{Title: "Introduce Redis Cache"},
				{Title: "Refactor Authentication Flow"},
			},
		}

		got, err := renderer.Render(rc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := `# Repository Context

Repository: test_repo

Generated: 2026-06-06T12:00:00Z

## Key Decisions

* Use Redis Cache
* Adopt gRPC

## Key Intents

* Reduce Latency
* Improve Reliability

## Key Facts

* Auth Service USES Redis
* API Gateway DEPENDS_ON Auth Service

## Recent Events

* Introduce Redis Cache
* Refactor Authentication Flow
`
		if got != expected {
			t.Errorf("expected:\n%q\ngot:\n%q", expected, got)
		}
	})

	t.Run("Deterministic rendering", func(t *testing.T) {
		rc := &contextpkg.RepositoryContext{
			RepositoryID: repoID,
			GeneratedAt:  genTime,
			Decisions: []*memorymodels.Decision{
				{Title: "Use Redis Cache"},
			},
			Facts: []*memorymodels.Fact{
				{Subject: "Auth Service", Predicate: "USES", Object: "Redis"},
			},
		}

		got1, err := renderer.Render(rc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got2, err := renderer.Render(rc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got1 != got2 {
			t.Errorf("rendering is not deterministic:\nGot1:\n%s\n\nGot2:\n%s", got1, got2)
		}
	})
}

func TestRenderer_Integration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-renderer-integration-test-*")
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

	if err := migrations.RunUp(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	ctx := stdcontext.Background()
	repoID := "repo_context"

	// 1. Insert Repository
	_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repoID, "Repo Context", tempDir, "main")
	if err != nil {
		t.Fatalf("failed to insert repository: %v", err)
	}

	// 2. Insert Source
	_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_context", repoID, "adr", "docs/adr/0001.md", "Author", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("failed to insert source: %v", err)
	}

	// 3. Setup Stores & insert data
	eventStore := sqlite.NewEventStore(db)
	err = eventStore.UpsertEvent(ctx, &models.Event{
		ID:           "evt_1",
		RepositoryID: repoID,
		EventType:    "FEATURE",
		Title:        "Feature A",
		SourceID:     "src_context",
		Timestamp:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to write event: %v", err)
	}

	decisionStore := memorystorage.NewSQLiteDecisionStore(db)
	err = decisionStore.UpsertDecision(ctx, &memorymodels.Decision{
		ID:           "dec_1",
		RepositoryID: repoID,
		Title:        "Decision A",
		Status:       "Accepted",
		SourceID:     "src_context",
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to write decision: %v", err)
	}

	// 4. Instantiation of composed Readers, ContextReader, Generator and Renderer
	eventReader := storage.NewSQLiteEventReader(db)
	decisionReader := storage.NewSQLiteDecisionReader(db)
	intentReader := storage.NewSQLiteIntentReader(db)
	factReader := storage.NewSQLiteFactReader(db)

	contextReader := contextpkg.NewMemoryContextReader(eventReader, decisionReader, intentReader, factReader)
	generator := contextpkg.NewGenerator(contextReader)
	renderer := NewRenderer()

	// 5. Query context, generate and render
	rc, err := generator.Generate(ctx, repoID)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Override GeneratedAt for deterministic testing
	rc.GeneratedAt = time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC)

	got, err := renderer.Render(rc)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	expected := `# Repository Context

Repository: repo_context

Generated: 2026-06-06T12:00:00Z

## Key Decisions

* Decision A

## Recent Events

* Feature A
`
	if got != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, got)
	}
}
