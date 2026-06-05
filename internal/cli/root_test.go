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
		"scan",
		"ask",
		"explain",
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
		"✓ RepoNerve ready",
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
		if err != nil {
			t.Fatalf("unexpected error executing ask: %v", err)
		}

		expected := "Querying repository memory..."
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
	})
}

func TestExplainCommand(t *testing.T) {
	t.Run("no arguments", func(t *testing.T) {
		output, err := executeCommand("explain")
		if err != nil {
			t.Fatalf("unexpected error executing explain: %v", err)
		}

		expected := "Explaining component..."
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got %q", expected, output)
		}
	})

	t.Run("with component name", func(t *testing.T) {
		output, err := executeCommand("explain", "services/auth")
		if err != nil {
			t.Fatalf("unexpected error executing explain: %v", err)
		}

		expected := `Explaining component "services/auth"...`
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got %q", expected, output)
		}
	})
}
