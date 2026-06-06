package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"reponerve/internal/cli"
)

func TestExplainCommandsIntegration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-explain-integration-*")
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

	// Create an ADR to extract decision/fact/intent/relationship
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

	// Retrieve IDs from lists
	listEvtsOut, err := execute("memory", "list", "events")
	if err != nil {
		t.Fatalf("list events failed: %v", err)
	}
	var eventID string
	for _, line := range strings.Split(listEvtsOut, "\n") {
		if strings.Contains(strings.ToLower(line), "initial repository commit") {
			parts := strings.Split(line, "|")
			eventID = strings.TrimSpace(parts[0])
			break
		}
	}

	listDecsOut, err := execute("memory", "list", "decisions")
	if err != nil {
		t.Fatalf("list decisions failed: %v", err)
	}
	var decisionID string
	for _, line := range strings.Split(listDecsOut, "\n") {
		if strings.Contains(line, "1. Use Go") {
			parts := strings.Split(line, "|")
			decisionID = strings.TrimSpace(parts[0])
			break
		}
	}
	if decisionID == "" {
		t.Fatalf("could not find decision ID in list output: %s", listDecsOut)
	}

	t.Run("Explain Decision Integration", func(t *testing.T) {
		out, err := execute("memory", "explain", "decision", decisionID)
		if err != nil {
			t.Fatalf("failed to explain decision: %v", err)
		}
		expectedParts := []string{"Decision:", "1. Use Go", "Reason:", "Simplify Configuration", "Supporting Facts:", "API Gateway USES Go"}
		for _, part := range expectedParts {
			if !strings.Contains(out, part) {
				t.Errorf("expected output to contain %q, got:\n%s", part, out)
			}
		}
	})

	t.Run("Explain Event Integration", func(t *testing.T) {
		// Even if eventID has no direct link, it should output the Event header and title cleanly.
		if eventID != "" {
			out, err := execute("memory", "explain", "event", eventID)
			if err != nil {
				t.Fatalf("failed to explain event: %v", err)
			}
			if !strings.Contains(out, "Event:") || !strings.Contains(strings.ToLower(out), "initial repository commit") {
				t.Errorf("unexpected explain event output: %s", out)
			}
		}
	})

	t.Run("Explain Non-existent Error Integration", func(t *testing.T) {
		_, err := execute("memory", "explain", "decision", "non-existent-id")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), `decision with ID "non-existent-id" not found`) {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}
