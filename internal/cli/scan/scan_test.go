package scancmd

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/reponerve/reponerve/internal/config"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

func runGit(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to run git command %v: %v", args, err)
	}
}

func TestScanCommand_ErrorsIfNoWorkspace(t *testing.T) {
	// Ensure we point to a non-existent workspace
	tempDir, err := os.MkdirTemp("", "reponerve-scan-cli-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("REPONERVE_WORKSPACE", filepath.Join(tempDir, "nonexistent"))
	defer os.Unsetenv("REPONERVE_WORKSPACE")

	cmd := NewCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err = cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("expected scan command to fail when workspace is not initialized")
	}

	if !strings.Contains(err.Error(), "workspace not initialized") {
		t.Errorf("expected error message to contain 'workspace not initialized', got %q", err.Error())
	}
}

func TestScanCommand_Success(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-scan-cli-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo in tempDir
	runGit(t, tempDir, "init")
	runGit(t, tempDir, "config", "user.name", "Test User")
	runGit(t, tempDir, "config", "user.email", "test@reponerve.com")
	if err := os.WriteFile(filepath.Join(tempDir, "file.txt"), []byte("commit 1"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	runGit(t, tempDir, "add", "file.txt")
	runGit(t, tempDir, "commit", "-m", "first commit")
	runGit(t, tempDir, "branch", "-M", "main")

	// Set workspace
	workspaceDir := filepath.Join(tempDir, ".reponerve")
	os.Setenv("REPONERVE_WORKSPACE", workspaceDir)
	defer os.Unsetenv("REPONERVE_WORKSPACE")

	// Initialize config
	cfg, err := config.Initialize(workspaceDir)
	if err != nil {
		t.Fatalf("failed to initialize config: %v", err)
	}

	// Update configuration repository path to point to tempDir instead of default "."
	// We can update cfg and write it back or just change working directory.
	// Changing working directory to tempDir is simple and robust.
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Run migrations to prepare DB
	db, err := sqlite.Open(cfg.Storage.SQLitePath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	if err := migrations.RunUp(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}
	db.Close()

	cmd := NewCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err = cmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("scan command failed: %v", err)
	}

	output := buf.String()
	expectedSubstrings := []string{
		"Scanning repository...",
		"✓ Repository discovered",
		"✓ 1 commits indexed",
		"✓ 0 ADRs indexed",
		"Scan completed.",
	}
	for _, sub := range expectedSubstrings {
		if !strings.Contains(output, sub) {
			t.Errorf("expected output to contain %q, got %q", sub, output)
		}
	}
}
