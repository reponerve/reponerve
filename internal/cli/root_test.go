package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	tempDir, err := os.MkdirTemp("", "reponerve-cli-test-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	origDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if err := os.Chdir(tempDir); err != nil {
		panic(err)
	}

	// Initialize git repo in tempDir so scanning works
	initCmd := exec.Command("git", "init")
	initCmd.Dir = tempDir
	if err := initCmd.Run(); err != nil {
		panic(err)
	}
	// Setup user details
	exec.Command("git", "config", "user.name", "Test").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	// Add initial commit
	os.WriteFile(filepath.Join(tempDir, "dummy.txt"), []byte("dummy"), 0644)
	exec.Command("git", "add", "dummy.txt").Run()
	exec.Command("git", "commit", "-m", "initial commit").Run()

	os.Setenv("REPONERVE_WORKSPACE", filepath.Join(tempDir, ".reponerve"))

	code := m.Run()

	os.Chdir(origDir)
	os.Exit(code)
}

func executeCommand(args ...string) (string, error) {
	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestRootCommandHelp(t *testing.T) {
	output, err := executeCommand("--help")
	if err != nil {
		t.Fatalf("unexpected error executing help: %v", err)
	}

	expectedSubstrings := []string{
		"reponerve",
		"init",
		"integrate",
		"scan",
		"hook",
		"ask",
		"search",
		"explain",
		"explain-file",
		"explain-function",
		"explain-struct",
		"explain-interface",
		"explain-type",
		"plan",
		"review",
		"impact",
		"context",
		"mcp",
	}

	for _, sub := range expectedSubstrings {
		if !strings.Contains(output, sub) {
			t.Errorf("expected help output to contain %q", sub)
		}
	}
}

func TestInitCommand(t *testing.T) {
	output, err := executeCommand("init")
	if err != nil {
		t.Fatalf("unexpected error executing init: %v", err)
	}

	expectedSubstrings := []string{
		"✓ Workspace created",
		"✓ Configuration created",
		"✓ Database initialized",
		"✓ IDE integration installed",
		"✓ RepoNerve ready",
		"reponerve scan",
	}
	for _, expected := range expectedSubstrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got %q", expected, output)
		}
	}
}

func TestScanCommand(t *testing.T) {
	output, err := executeCommand("scan")
	if err != nil {
		t.Fatalf("unexpected error executing scan: %v", err)
	}

	expected := "Scanning repository..."
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got %q", expected, output)
	}
}

func TestAskCommand(t *testing.T) {
	t.Run("no arguments", func(t *testing.T) {
		output, err := executeCommand("ask")
		if err == nil {
			t.Fatalf("expected error executing ask with no arguments")
		}

		expected := "accepts 1 arg(s), received 0"
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got %q", expected, output)
		}
	})

	t.Run("with question", func(t *testing.T) {
		output, err := executeCommand("ask", "Why was Redis introduced?")
		if err != nil {
			t.Fatalf("unexpected error executing ask: %v", err)
		}

		expected := `Querying repository memory for: "Why was Redis introduced?"...`
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got %q", expected, output)
		}

		if !strings.Contains(output, "Answer Type:") {
			t.Errorf("expected structured answer output, got %q", output)
		}
		if !strings.Contains(output, "No deterministic answer pattern matched") &&
			!strings.Contains(output, "Search found") &&
			!strings.Contains(output, "No decision evidence") {
			t.Errorf("expected fallback or search summary in output, got %q", output)
		}
	})
}

func TestPlanCommand(t *testing.T) {
	t.Run("no arguments", func(t *testing.T) {
		_, err := executeCommand("plan")
		if err == nil {
			t.Fatal("expected plan without task to fail")
		}
	})

	t.Run("with task", func(t *testing.T) {
		output, err := executeCommand("plan", "Add OAuth login")
		if err != nil {
			t.Fatalf("unexpected error executing plan: %v", err)
		}
		if !strings.Contains(output, "Task: Add OAuth login") {
			t.Errorf("expected task in output, got %q", output)
		}
		if !strings.Contains(output, "Suggested Workflow: change_preparation") {
			t.Errorf("expected workflow in output, got %q", output)
		}
	})
}

func TestReviewCommand(t *testing.T) {
	t.Run("no arguments", func(t *testing.T) {
		_, err := executeCommand("review")
		if err == nil {
			t.Fatal("expected review without topic to fail")
		}
	})

	t.Run("with topic", func(t *testing.T) {
		output, err := executeCommand("review", "metadata panel")
		if err != nil {
			t.Fatalf("unexpected error executing review: %v", err)
		}
		if !strings.Contains(output, "Topic: metadata panel") {
			t.Errorf("expected topic in output, got %q", output)
		}
		if !strings.Contains(output, "Suggested Workflow: review_preparation") {
			t.Errorf("expected workflow in output, got %q", output)
		}
	})
}

func TestImpactCommand(t *testing.T) {
	t.Run("no arguments", func(t *testing.T) {
		_, err := executeCommand("impact")
		if err == nil {
			t.Fatal("expected impact without subject to fail")
		}
	})

	t.Run("with subject", func(t *testing.T) {
		output, err := executeCommand("impact", "user-service")
		if err != nil {
			t.Fatalf("unexpected error executing impact: %v", err)
		}
		if !strings.Contains(output, "Subject: user-service") {
			t.Errorf("expected subject in output, got %q", output)
		}
	})
}

func TestExplainFileCommand(t *testing.T) {
	t.Run("no arguments", func(t *testing.T) {
		_, err := executeCommand("explain-file")
		if err == nil {
			t.Fatal("expected explain-file without path to fail")
		}
		if !strings.Contains(err.Error(), "accepts 1 arg") {
			t.Fatalf("expected arg validation error, got %v", err)
		}
	})
}

func TestExplainCommand(t *testing.T) {
	t.Run("no arguments", func(t *testing.T) {
		_, err := executeCommand("explain")
		if err == nil {
			t.Fatal("expected explain without topic to fail")
		}
		if !strings.Contains(err.Error(), "accepts 1 arg") {
			t.Fatalf("expected arg validation error, got %v", err)
		}
	})

	t.Run("with topic", func(t *testing.T) {
		output, err := executeCommand("explain", "services/auth")
		if err != nil {
			t.Fatalf("unexpected error executing explain: %v", err)
		}

		expected := "Topic: services/auth"
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got %q", expected, output)
		}
		if !strings.Contains(output, "REPOSITORY CONTEXT") {
			t.Errorf("expected repository context section, got %q", output)
		}
	})
}
