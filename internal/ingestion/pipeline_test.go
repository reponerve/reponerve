package ingestion

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	memorystorage "github.com/reponerve/reponerve/internal/memory/storage"
	"github.com/reponerve/reponerve/internal/scanner/repository"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
	"github.com/reponerve/reponerve/pkg/models"
)

func containsStr(s, substr string) bool {
	return strings.Contains(s, substr)
}

type mockScanner struct {
	scanFunc func(ctx context.Context, repo *models.Repository) ([]*models.Source, error)
}

func (m *mockScanner) Scan(ctx context.Context, repo *models.Repository) ([]*models.Source, error) {
	return m.scanFunc(ctx, repo)
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()
	if len(r.Scanners()) != 0 {
		t.Fatalf("expected empty registry, got %d scanners", len(r.Scanners()))
	}

	s1 := &mockScanner{}
	r.Register("mock", s1)

	if len(r.Scanners()) != 1 {
		t.Fatalf("expected 1 scanner, got %d", len(r.Scanners()))
	}
	if r.Scanners()[0].Name != "mock" {
		t.Errorf("expected scanner name %q, got %q", "mock", r.Scanners()[0].Name)
	}
	if r.Scanners()[0].Scanner != s1 {
		t.Error("registered scanner pointer does not match")
	}
}

func TestPipeline_Execute_Success(t *testing.T) {
	r := NewRegistry()

	src1 := &models.Source{ID: "src1"}
	s1 := &mockScanner{
		scanFunc: func(ctx context.Context, repo *models.Repository) ([]*models.Source, error) {
			return []*models.Source{src1}, nil
		},
	}

	src2 := &models.Source{ID: "src2"}
	s2 := &mockScanner{
		scanFunc: func(ctx context.Context, repo *models.Repository) ([]*models.Source, error) {
			return []*models.Source{src2}, nil
		},
	}

	r.Register("mock1", s1)
	r.Register("mock2", s2)

	p := NewPipeline(r)
	repo := &models.Repository{ID: "test-repo"}

	sources, err := p.Execute(context.Background(), repo)
	if err != nil {
		t.Fatalf("expected successful execution, got err: %v", err)
	}

	if len(sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(sources))
	}
	if sources[0].ID != "src1" || sources[1].ID != "src2" {
		t.Errorf("sources returned out of order or incorrect: %v", sources)
	}
}

func TestPipeline_Execute_Error(t *testing.T) {
	r := NewRegistry()

	s1 := &mockScanner{
		scanFunc: func(ctx context.Context, repo *models.Repository) ([]*models.Source, error) {
			return nil, errors.New("scan failed")
		},
	}
	r.Register("failing-mock", s1)

	p := NewPipeline(r)
	repo := &models.Repository{ID: "test-repo"}

	_, err := p.Execute(context.Background(), repo)
	if err == nil {
		t.Fatal("expected pipeline to return error when scanner fails")
	}
	if err.Error() == "" {
		t.Errorf("expected non-empty error message, got empty string")
	}
	expectedSubstr := `scanner "failing-mock" failed`
	if !containsStr(err.Error(), expectedSubstr) {
		t.Errorf("expected error to contain %q, got: %v", expectedSubstr, err)
	}
}

func TestCoordinator_Run(t *testing.T) {
	// Setup temporary workspace & Git repository for discovery
	tempDir, err := os.MkdirTemp("", "reponerve-coordinator-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git: %v", err)
	}

	// Create a commit so that HEAD exists
	if err := os.WriteFile(filepath.Join(tempDir, "dummy.txt"), []byte("dummy"), 0644); err != nil {
		t.Fatalf("failed to write dummy file: %v", err)
	}
	runCmd := func(args ...string) {
		c := exec.Command("git", args...)
		c.Dir = tempDir
		if err := c.Run(); err != nil {
			t.Fatalf("failed to run git %v: %v", args, err)
		}
	}
	runCmd("config", "user.name", "Test User")
	runCmd("config", "user.email", "test@example.com")
	runCmd("add", "dummy.txt")
	runCmd("commit", "-m", "initial commit")

	// Create SQLite database
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := migrations.RunUp(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	discovery := repository.NewGitDiscovery()
	reg := NewRegistry()
	pipeline := NewPipeline(reg)

	// Register a dummy scanner that records custom execution time
	dummySource := &models.Source{
		ID:         "dummy",
		SourceType: "dummy_type",
		Reference:  "dummy_ref",
		Timestamp:  time.Now(),
	}
	reg.Register("dummy", &mockScanner{
		scanFunc: func(ctx context.Context, repo *models.Repository) ([]*models.Source, error) {
			dummySource.RepositoryID = repo.ID
			return []*models.Source{dummySource}, nil
		},
	})

	repoStore := sqlite.NewRepositoryStore(db)
	sourceStore := sqlite.NewSourceStore(db)
	scanStateStore := sqlite.NewScanStateStore(db)
	eventStore := sqlite.NewEventStore(db)
	decisionStore := memorystorage.NewSQLiteDecisionStore(db)
	intentStore := memorystorage.NewSQLiteIntentStore(db)
	factStore := memorystorage.NewSQLiteFactStore(db)
	relationshipStore := memorystorage.NewSQLiteRelationshipStore(db)
	contributorStore := sqlite.NewSQLiteContributorStore(db)
	expertiseStore := sqlite.NewSQLiteExpertiseStore(db)
	coord := NewCoordinator(discovery, repoStore, sourceStore, scanStateStore, eventStore, decisionStore, intentStore, factStore, relationshipStore, contributorStore, expertiseStore, nil, nil, pipeline)
	ctx := context.Background()

	result, err := coord.Run(ctx, tempDir)
	if err != nil {
		t.Fatalf("coordinator run failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected ScanResult, got nil")
	}
	if result.CommitsIndexed != 0 {
		t.Errorf("expected 0 commits indexed, got %d", result.CommitsIndexed)
	}
	if result.ADRsIndexed != 0 {
		t.Errorf("expected 0 ADRs indexed, got %d", result.ADRsIndexed)
	}
	if result.Duration <= 0 {
		t.Errorf("expected non-zero duration, got %v", result.Duration)
	}

	// Verify repository is stored in database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM repositories WHERE id = ?", result.RepositoryID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query repository count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected repository record in database, got count %d", count)
	}

	// Verify source is stored in database
	var sourceCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sources WHERE id = ?", "dummy").Scan(&sourceCount)
	if err != nil {
		t.Fatalf("failed to query source count: %v", err)
	}
	if sourceCount != 1 {
		t.Errorf("expected source record in database, got count %d", sourceCount)
	}

	// Verify scan state is stored in database
	var stateCount int
	err = db.QueryRow("SELECT COUNT(*) FROM scan_state WHERE repository_id = ?", result.RepositoryID).Scan(&stateCount)
	if err != nil {
		t.Fatalf("failed to query scan state count: %v", err)
	}
	if stateCount != 1 {
		t.Errorf("expected scan state in database, got count %d", stateCount)
	}
}
