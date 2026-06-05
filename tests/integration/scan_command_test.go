package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"reponerve/internal/cli"
	"reponerve/internal/storage/sqlite"
)

func TestScanCommandIntegration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-scan-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up Git repository
	runGitCommand(t, tempDir, "init")
	runGitCommand(t, tempDir, "config", "user.name", "E2E Tester")
	runGitCommand(t, tempDir, "config", "user.email", "e2e@reponerve.com")

	// Create a commit
	if err := os.WriteFile(filepath.Join(tempDir, "code.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	runGitCommand(t, tempDir, "add", "code.go")
	runGitCommand(t, tempDir, "commit", "-m", "initial repository commit")
	runGitCommand(t, tempDir, "branch", "-M", "main")

	// Create an ADR
	adrDir := filepath.Join(tempDir, "docs", "adr")
	if err := os.MkdirAll(adrDir, 0755); err != nil {
		t.Fatalf("failed to create ADR folder: %v", err)
	}
	adrContent := `# 1. Use Go

## Status

Accepted

Go is chosen.`
	if err := os.WriteFile(filepath.Join(adrDir, "0001-use-go.md"), []byte(adrContent), 0644); err != nil {
		t.Fatalf("failed to write ADR file: %v", err)
	}

	// Change working directory to tempDir
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current dir: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Set workspace env var
	workspaceDir := filepath.Join(tempDir, ".reponerve")
	os.Setenv("REPONERVE_WORKSPACE", workspaceDir)
	defer os.Unsetenv("REPONERVE_WORKSPACE")

	// Run "init" command
	initBuf := new(bytes.Buffer)
	initCmd := cli.NewRootCmd()
	initCmd.SetOut(initBuf)
	initCmd.SetErr(initBuf)
	initCmd.SetArgs([]string{"init"})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("reponerve init failed: %v, output: %s", err, initBuf.String())
	}

	// Run "scan" command
	scanBuf := new(bytes.Buffer)
	scanCmd := cli.NewRootCmd()
	scanCmd.SetOut(scanBuf)
	scanCmd.SetErr(scanBuf)
	scanCmd.SetArgs([]string{"scan"})
	if err := scanCmd.Execute(); err != nil {
		t.Fatalf("reponerve scan failed: %v, output: %s", err, scanBuf.String())
	}

	output := scanBuf.String()
	expectedLines := []string{
		"Scanning repository...",
		"✓ Repository discovered",
		"✓ 1 commits indexed",
		"✓ 1 ADRs indexed",
		"Scan completed.",
	}
	for _, expected := range expectedLines {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got %q", expected, output)
		}
	}

	// Verify database contents
	db, err := sqlite.Open(filepath.Join(workspaceDir, "memory.db"))
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	var commitCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sources WHERE source_type = 'commit'").Scan(&commitCount)
	if err != nil {
		t.Fatalf("failed to query commit count: %v", err)
	}
	if commitCount != 1 {
		t.Errorf("expected 1 commit in db, got %d", commitCount)
	}

	var adrCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sources WHERE source_type = 'adr'").Scan(&adrCount)
	if err != nil {
		t.Fatalf("failed to query adr count: %v", err)
	}
	if adrCount != 1 {
		t.Errorf("expected 1 ADR in db, got %d", adrCount)
	}
}
