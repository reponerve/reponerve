package repository

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runGitCommand(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to execute git command %v: %v", args, err)
	}
}

func createTestRepo(t *testing.T, defaultBranch string) string {
	tempDir, err := os.MkdirTemp("", "reponerve-discovery-unit-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}

	runGitCommand(t, tempDir, "init")
	runGitCommand(t, tempDir, "config", "user.name", "Test")
	runGitCommand(t, tempDir, "config", "user.email", "test@test.com")

	if err := os.WriteFile(filepath.Join(tempDir, "file.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write dummy file: %v", err)
	}
	runGitCommand(t, tempDir, "add", "file.txt")
	runGitCommand(t, tempDir, "commit", "-m", "initial commit")

	if defaultBranch != "" {
		runGitCommand(t, tempDir, "branch", "-M", defaultBranch)
	}

	return tempDir
}

func TestDiscover_Success(t *testing.T) {
	tempRepo := createTestRepo(t, "main-branch")
	defer os.RemoveAll(tempRepo)

	service := NewGitDiscovery()
	ctx := context.Background()
	repo, err := service.Discover(ctx, tempRepo)
	if err != nil {
		t.Fatalf("failed to discover repository: %v", err)
	}

	expectedName := filepath.Base(tempRepo)
	if repo.Name != expectedName {
		t.Errorf("expected name to be %q, got %q", expectedName, repo.Name)
	}

	absPath, _ := filepath.Abs(tempRepo)
	if repo.Path != absPath {
		t.Errorf("expected path to be %q, got %q", absPath, repo.Path)
	}

	if repo.DefaultBranch != "main-branch" {
		t.Errorf("expected default branch to be 'main-branch', got %q", repo.DefaultBranch)
	}

	if repo.ID == "" {
		t.Errorf("expected repo ID to be set, got empty string")
	}

	if repo.UpdatedAt.IsZero() {
		t.Errorf("expected UpdatedAt to be set, got zero time")
	}
}

func TestDiscover_NotGitRepository(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-discovery-non-git-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	service := NewGitDiscovery()
	ctx := context.Background()
	_, err = service.Discover(ctx, tempDir)
	if err == nil {
		t.Error("expected error discovering a non-git directory, but it succeeded")
	}
}

func TestDiscover_EmptyRepository(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-discovery-empty-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	runGitCommand(t, tempDir, "init")
	runGitCommand(t, tempDir, "symbolic-ref", "HEAD", "refs/heads/empty-default")

	service := NewGitDiscovery()
	ctx := context.Background()
	repo, err := service.Discover(ctx, tempDir)
	if err != nil {
		t.Fatalf("failed to discover empty repository: %v", err)
	}

	if repo.DefaultBranch != "empty-default" {
		t.Errorf("expected default branch to be 'empty-default', got %q", repo.DefaultBranch)
	}
}

func TestDiscover_DetachedHEAD(t *testing.T) {
	tempRepo := createTestRepo(t, "main-branch")
	defer os.RemoveAll(tempRepo)

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = tempRepo
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to get HEAD commit: %v", err)
	}
	commitHash := strings.TrimSpace(string(out))

	runGitCommand(t, tempRepo, "checkout", commitHash)

	service := NewGitDiscovery()
	ctx := context.Background()
	repo, err := service.Discover(ctx, tempRepo)
	if err != nil {
		t.Fatalf("failed to discover repository in detached HEAD: %v", err)
	}

	if repo.DefaultBranch == "" {
		t.Errorf("expected a non-empty default branch fallback in detached HEAD state")
	}
}

func TestDiscover_InvalidPath(t *testing.T) {
	service := NewGitDiscovery()
	ctx := context.Background()
	_, err := service.Discover(ctx, "/path/does/not/exist/reponerve-invalid")
	if err == nil {
		t.Error("expected error for non-existent path, but it succeeded")
	}
}
