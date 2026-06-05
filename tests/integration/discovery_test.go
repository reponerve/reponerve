package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"reponerve/internal/scanner/repository"
	"reponerve/internal/storage/migrations"
	"reponerve/internal/storage/sqlite"
)

func runGitCommand(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to execute git command %v: %v", args, err)
	}
}

func createTestGitRepo(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "reponerve-discovery-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}

	runGitCommand(t, tempDir, "init")
	runGitCommand(t, tempDir, "config", "user.name", "Test")
	runGitCommand(t, tempDir, "config", "user.email", "test@test.com")

	if err := os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("# Integration"), 0644); err != nil {
		t.Fatalf("failed to write README: %v", err)
	}
	runGitCommand(t, tempDir, "add", "README.md")
	runGitCommand(t, tempDir, "commit", "-m", "initial commit")
	runGitCommand(t, tempDir, "branch", "-M", "main-integration")

	return tempDir
}

func TestDiscoveryIntegration(t *testing.T) {
	repoPath := createTestGitRepo(t)
	defer os.RemoveAll(repoPath)

	dbPath := filepath.Join(repoPath, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite connection: %v", err)
	}
	defer db.Close()

	err = migrations.RunUp(db)
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	ctx := context.Background()
	service := repository.NewGitDiscovery(db)
	repo, err := service.Discover(ctx, repoPath)
	if err != nil {
		t.Fatalf("failed to discover repository: %v", err)
	}

	err = service.Store(ctx, repo)
	if err != nil {
		t.Fatalf("failed to store repository metadata: %v", err)
	}

	var id, name, path, defaultBranch string
	err = db.QueryRow("SELECT id, name, path, default_branch FROM repositories WHERE id = ?", repo.ID).
		Scan(&id, &name, &path, &defaultBranch)
	if err != nil {
		t.Fatalf("failed to query repository metadata from DB: %v", err)
	}

	if id != repo.ID {
		t.Errorf("expected ID %q, got %q", repo.ID, id)
	}
	if name != repo.Name {
		t.Errorf("expected Name %q, got %q", repo.Name, name)
	}
	if path != repo.Path {
		t.Errorf("expected Path %q, got %q", repo.Path, path)
	}
	if defaultBranch != "main-integration" {
		t.Errorf("expected DefaultBranch 'main-integration', got %q", defaultBranch)
	}

	repo.Name = "updated-integration-name"
	repo.DefaultBranch = "updated-integration-branch"
	err = service.Store(ctx, repo)
	if err != nil {
		t.Fatalf("failed to upsert repository metadata: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM repositories").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query repo record count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected exactly 1 repository record after upsert, got %d", count)
	}

	var updatedName, updatedBranch string
	err = db.QueryRow("SELECT name, default_branch FROM repositories WHERE id = ?", repo.ID).
		Scan(&updatedName, &updatedBranch)
	if err != nil {
		t.Fatalf("failed to query updated metadata: %v", err)
	}

	if updatedName != "updated-integration-name" {
		t.Errorf("expected updated name 'updated-integration-name', got %q", updatedName)
	}
	if updatedBranch != "updated-integration-branch" {
		t.Errorf("expected updated default branch 'updated-integration-branch', got %q", updatedBranch)
	}
}
