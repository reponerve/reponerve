package mcp

import (
	stdcontext "context"
	"os"
	"path/filepath"
	"testing"

	"github.com/reponerve/reponerve/internal/context"
	"github.com/reponerve/reponerve/internal/context/render"
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	ownershipquery "github.com/reponerve/reponerve/internal/ownership/query"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
	models "github.com/reponerve/reponerve/pkg/models"
)

// --- Mock/Dummy Readers satisfying Query Reader interfaces ---

type dummyEventReader struct{}

func (d *dummyEventReader) GetByID(ctx stdcontext.Context, id string) (*models.Event, error) {
	return nil, nil
}
func (d *dummyEventReader) ListByRepository(ctx stdcontext.Context, repoID string) ([]*models.Event, error) {
	return nil, nil
}
func (d *dummyEventReader) ListAll(ctx stdcontext.Context) ([]*models.Event, error) {
	return nil, nil
}

type dummyDecisionReader struct{}

func (d *dummyDecisionReader) GetByID(ctx stdcontext.Context, id string) (*memorymodels.Decision, error) {
	return nil, nil
}
func (d *dummyDecisionReader) ListByRepository(ctx stdcontext.Context, repoID string) ([]*memorymodels.Decision, error) {
	return nil, nil
}
func (d *dummyDecisionReader) ListAll(ctx stdcontext.Context) ([]*memorymodels.Decision, error) {
	return nil, nil
}

type dummyIntentReader struct{}

func (d *dummyIntentReader) GetByID(ctx stdcontext.Context, id string) (*memorymodels.Intent, error) {
	return nil, nil
}
func (d *dummyIntentReader) ListByRepository(ctx stdcontext.Context, repoID string) ([]*memorymodels.Intent, error) {
	return nil, nil
}
func (d *dummyIntentReader) ListAll(ctx stdcontext.Context) ([]*memorymodels.Intent, error) {
	return nil, nil
}

type dummyFactReader struct{}

func (d *dummyFactReader) GetByID(ctx stdcontext.Context, id string) (*memorymodels.Fact, error) {
	return nil, nil
}
func (d *dummyFactReader) ListByRepository(ctx stdcontext.Context, repoID string) ([]*memorymodels.Fact, error) {
	return nil, nil
}
func (d *dummyFactReader) ListAll(ctx stdcontext.Context) ([]*memorymodels.Fact, error) {
	return nil, nil
}

type dummyRelationshipReader struct{}

func (d *dummyRelationshipReader) GetByID(ctx stdcontext.Context, id string) (*memorymodels.Relationship, error) {
	return nil, nil
}
func (d *dummyRelationshipReader) ListByRepository(ctx stdcontext.Context, repoID string) ([]*memorymodels.Relationship, error) {
	return nil, nil
}
func (d *dummyRelationshipReader) ListAll(ctx stdcontext.Context) ([]*memorymodels.Relationship, error) {
	return nil, nil
}

type dummyContributorReader struct{}

func (d *dummyContributorReader) GetByID(ctx stdcontext.Context, repositoryID string, id string) (*models.Contributor, error) {
	return nil, nil
}
func (d *dummyContributorReader) ListByRepository(ctx stdcontext.Context, repositoryID string) ([]*models.Contributor, error) {
	return nil, nil
}

type dummyExpertiseReader struct{}

func (d *dummyExpertiseReader) ListByRepository(ctx stdcontext.Context, repositoryID string) ([]*models.Expertise, error) {
	return nil, nil
}
func (d *dummyExpertiseReader) ListByContributor(ctx stdcontext.Context, repositoryID string, contributorID string) ([]*models.Expertise, error) {
	return nil, nil
}

type dummySourceReader struct{}

func (d *dummySourceReader) ListByRepository(ctx stdcontext.Context, repositoryID string) ([]*models.Source, error) {
	return nil, nil
}

// --- Unit Tests ---

func TestRegistry_Unit(t *testing.T) {
	t.Run("Default tools registration", func(t *testing.T) {
		r := NewRegistry()
		list := r.List()

		// Expect exactly 27 tools registered initially
		if len(list) != 27 {
			t.Errorf("expected 27 initial tools, got %d", len(list))
		}

		expectedNames := []string{
			"analyze_impact",
			"discover_knowledge",
			"explain_decision",
			"explain_event",
			"export_context",
			"find_dependencies",
			"find_dependents",
			"generate_change_plan",
			"generate_context",
			"generate_learning_path",
			"get_contributor",
			"get_decision",
			"get_event",
			"get_fact",
			"get_intent",
			"list_contributors",
			"list_decisions",
			"list_events",
			"list_expertise",
			"list_facts",
			"list_intents",
			"recommend_reviewers",
			"trace_contributor",
			"trace_decision",
			"trace_event",
			"trace_graph",
			"trace_path",
		}
		for i, name := range expectedNames {
			if list[i].Name != name {
				t.Errorf("expected tool %d name to be %q, got %q", i, name, list[i].Name)
			}
		}
	})

	t.Run("Successful custom tool registration and lookup", func(t *testing.T) {
		r := NewRegistry()
		tool := ToolDefinition{
			Name:        "custom_tool",
			Description: "A custom test tool",
		}

		err := r.Register(tool)
		if err != nil {
			t.Fatalf("unexpected error registering tool: %v", err)
		}

		got, exists := r.Get("custom_tool")
		if !exists {
			t.Fatal("expected tool 'custom_tool' to exist in registry")
		}
		if got.Description != tool.Description {
			t.Errorf("expected description %q, got %q", tool.Description, got.Description)
		}

		// Verify listing has 28 tools (27 defaults + 1 custom) sorted alphabetically
		list := r.List()
		if len(list) != 28 {
			t.Errorf("expected 28 tools after registration, got %d", len(list))
		}
	})

	t.Run("Duplicate tool registration prevention", func(t *testing.T) {
		r := NewRegistry()
		tool := ToolDefinition{
			Name:        "list_decisions",
			Description: "Duplicate description",
		}

		err := r.Register(tool)
		if err == nil {
			t.Fatal("expected duplicate registration to return error, got nil")
		}
	})

	t.Run("Empty tool name registration prevention", func(t *testing.T) {
		r := NewRegistry()
		tool := ToolDefinition{
			Name:        "",
			Description: "Empty name",
		}

		err := r.Register(tool)
		if err == nil {
			t.Fatal("expected empty name registration to return error, got nil")
		}
	})
}

func TestService_Unit(t *testing.T) {
	t.Run("Service construction with dummies", func(t *testing.T) {
		er := &dummyEventReader{}
		dr := &dummyDecisionReader{}
		ir := &dummyIntentReader{}
		fr := &dummyFactReader{}
		rr := &dummyRelationshipReader{}
		cr := &dummyContributorReader{}
		expr := &dummyExpertiseReader{}
		sr := &dummySourceReader{}

		ctxReader := context.NewMemoryContextReader(er, dr, ir, fr)
		generator := context.NewGenerator(ctxReader)
		renderer := render.NewRenderer()
		ownershipReader := ownershipquery.NewReader(cr, expr, sr, dr, fr, er)

		svc := NewService(dr, ir, fr, er, rr, generator, renderer, ownershipReader, nil, nil, nil, nil, nil, nil)
		if svc.DecisionReader != dr {
			t.Error("Service DecisionReader dependency not set correctly")
		}
		if svc.IntentReader != ir {
			t.Error("Service IntentReader dependency not set correctly")
		}
		if svc.FactReader != fr {
			t.Error("Service FactReader dependency not set correctly")
		}
		if svc.EventReader != er {
			t.Error("Service EventReader dependency not set correctly")
		}
		if svc.RelationshipReader != rr {
			t.Error("Service RelationshipReader dependency not set correctly")
		}
		if svc.Generator != generator {
			t.Error("Service Generator dependency not set correctly")
		}
		if svc.Renderer != renderer {
			t.Error("Service Renderer dependency not set correctly")
		}
		if svc.OwnershipReader != ownershipReader {
			t.Error("Service OwnershipReader dependency not set correctly")
		}
	})
}

// --- Integration Tests ---

func TestService_Integration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-mcp-integration-test-*")
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

	// Create SQLite readers
	er := storage.NewSQLiteEventReader(db)
	dr := storage.NewSQLiteDecisionReader(db)
	ir := storage.NewSQLiteIntentReader(db)
	fr := storage.NewSQLiteFactReader(db)
	rr := storage.NewSQLiteRelationshipReader(db)
	cr := storage.NewSQLiteContributorReader(db)
	expr := storage.NewSQLiteExpertiseReader(db)
	sr := storage.NewSQLiteSourceReader(db)

	ctxReader := context.NewMemoryContextReader(er, dr, ir, fr)
	generator := context.NewGenerator(ctxReader)
	renderer := render.NewRenderer()
	ownershipReader := ownershipquery.NewReader(cr, expr, sr, dr, fr, er)

	// Instantiate Service
	svc := NewService(dr, ir, fr, er, rr, generator, renderer, ownershipReader, nil, nil, nil, nil, nil, nil)

	// Verify dependencies are set and correctly typed
	if svc.DecisionReader == nil || svc.IntentReader == nil || svc.FactReader == nil ||
		svc.EventReader == nil || svc.RelationshipReader == nil || svc.Generator == nil || svc.Renderer == nil || svc.OwnershipReader == nil {
		t.Fatal("one or more service dependencies are nil")
	}

	// Validate Registry -> Service mapping
	r := NewRegistry()
	tools := r.List()

	expectedTools := map[string]bool{
		"explain_decision":       true,
		"explain_event":          true,
		"export_context":         true,
		"generate_context":       true,
		"get_contributor":        true,
		"get_decision":           true,
		"get_event":              true,
		"get_fact":               true,
		"get_intent":             true,
		"list_contributors":      true,
		"list_decisions":         true,
		"list_events":            true,
		"list_expertise":         true,
		"list_facts":             true,
		"list_intents":           true,
		"recommend_reviewers":    true,
		"trace_contributor":      true,
		"trace_decision":         true,
		"trace_event":            true,
		"trace_graph":            true,
		"trace_path":             true,
		"analyze_impact":         true,
		"find_dependencies":      true,
		"find_dependents":        true,
		"discover_knowledge":     true,
		"generate_learning_path": true,
		"generate_change_plan":   true,
	}

	for _, tool := range tools {
		if !expectedTools[tool.Name] {
			t.Errorf("unexpected tool %q in registry", tool.Name)
		}
	}
}
