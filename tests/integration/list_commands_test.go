package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"reponerve/internal/cli"
)

func TestListCommandsIntegration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-list-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up Git repository
	runGitCommand(t, tempDir, "init")
	runGitCommand(t, tempDir, "config", "user.name", "E2E Tester")
	runGitCommand(t, tempDir, "config", "user.email", "e2e@reponerve.com")

	// Create code and commit to extract event/intent
	if err := os.WriteFile(filepath.Join(tempDir, "code.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	runGitCommand(t, tempDir, "add", "code.go")
	runGitCommand(t, tempDir, "commit", "-m", "feat: initial repository commit to optimize storage")
	runGitCommand(t, tempDir, "branch", "-M", "main")

	// Create an ADR to extract decision/fact/relationship/intent
	adrDir := filepath.Join(tempDir, "docs", "adr")
	if err := os.MkdirAll(adrDir, 0755); err != nil {
		t.Fatalf("failed to create ADR folder: %v", err)
	}
	adrContent := `# 1. Use Go

## Status

Accepted

We need to simplify configuration and optimize deployment. Authentication Service uses Redis. API Gateway uses Go.`
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

	execute := func(args ...string) (string, error) {
		cmd := cli.NewRootCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs(args)
		err := cmd.Execute()
		return buf.String(), err
	}

	t.Run("Events Integration", func(t *testing.T) {
		out, err := execute("memory", "list", "events")
		if err != nil {
			t.Fatalf("failed to run list events: %v", err)
		}
		if !strings.Contains(strings.ToLower(out), "initial repository commit to optimize storage") {
			t.Errorf("expected commit title in output, got:\n%s", out)
		}
	})

	t.Run("Decisions Integration", func(t *testing.T) {
		out, err := execute("memory", "list", "decisions")
		if err != nil {
			t.Fatalf("failed to run list decisions: %v", err)
		}
		if !strings.Contains(out, "1. Use Go") || !strings.Contains(out, "Accepted") {
			t.Errorf("expected decision in output, got:\n%s", out)
		}
	})

	t.Run("Intents Integration", func(t *testing.T) {
		out, err := execute("memory", "list", "intents")
		if err != nil {
			t.Fatalf("failed to run list intents: %v", err)
		}
		if !strings.Contains(out, "Optimize Storage") || !strings.Contains(out, "Simplify Configuration") {
			t.Errorf("expected intents in output, got:\n%s", out)
		}
	})

	t.Run("Facts Integration", func(t *testing.T) {
		out, err := execute("memory", "list", "facts")
		if err != nil {
			t.Fatalf("failed to run list facts: %v", err)
		}
		if !strings.Contains(out, "Authentication Service") || !strings.Contains(out, "USES") || !strings.Contains(out, "Redis") {
			t.Errorf("expected facts in output, got:\n%s", out)
		}
	})

	t.Run("Relationships Integration", func(t *testing.T) {
		out, err := execute("memory", "list", "relationships")
		if err != nil {
			t.Fatalf("failed to run list relationships: %v", err)
		}
		if !strings.Contains(out, "INTENT_DRIVES_DECISION") && !strings.Contains(out, "FACT_SUPPORTS_DECISION") {
			t.Errorf("expected relationships in output, got:\n%s", out)
		}
	})
}
