package contextcmd

import (
	stdcontext "context"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	memorystorage "github.com/reponerve/reponerve/internal/memory/storage"
	repository "github.com/reponerve/reponerve/internal/scanner/repository"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

func executeContextCommand(args ...string) (string, error) {
	cmd := NewCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestContextCommandRegistration(t *testing.T) {
	cmd := NewCommand()
	if cmd.Use != "context" {
		t.Errorf("expected command Use 'context', got %q", cmd.Use)
	}

	subCommands := cmd.Commands()
	if len(subCommands) != 2 {
		t.Errorf("expected 2 subcommands, got %d", len(subCommands))
	}
	names := map[string]bool{}
	for _, c := range subCommands {
		names[c.Name()] = true
	}
	if !names["generate"] || !names["export"] {
		t.Errorf("expected generate and export subcommands to be registered, got: %v", names)
	}
}

func TestContextGenerateCommand(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-context-cmd-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origWorkspace := os.Getenv("REPONERVE_WORKSPACE")
	defer func() {
		if origWorkspace != "" {
			os.Setenv("REPONERVE_WORKSPACE", origWorkspace)
		} else {
			os.Unsetenv("REPONERVE_WORKSPACE")
		}
	}()

	workspacePath := filepath.Join(tempDir, ".reponerve")
	os.Setenv("REPONERVE_WORKSPACE", workspacePath)

	t.Run("Missing workspace", func(t *testing.T) {
		// Ensure environment workspace dir does not exist
		os.RemoveAll(workspacePath)

		_, err := executeContextCommand("generate")
		if err == nil {
			t.Fatal("expected error on missing workspace, got nil")
		}
		if !strings.Contains(err.Error(), "workspace not initialized") {
			t.Errorf("expected error message to contain 'workspace not initialized', got: %v", err)
		}
	})

	t.Run("Empty repository (no context)", func(t *testing.T) {
		os.RemoveAll(tempDir)
		os.MkdirAll(tempDir, 0755)

		// Initialize Git repository so discovery succeeds
		gitInit := exec.Command("git", "init")
		gitInit.Dir = tempDir
		if err := gitInit.Run(); err != nil {
			t.Fatalf("failed to init git: %v", err)
		}

		if err := os.MkdirAll(workspacePath, 0755); err != nil {
			t.Fatalf("failed to create workspace dir: %v", err)
		}
		configYAML := "repository:\n  path: " + tempDir + "\nstorage:\n  sqlite_path: " + filepath.Join(workspacePath, "memory.db") + "\nai:\n  provider: none\n"
		if err := os.WriteFile(filepath.Join(workspacePath, "config.yaml"), []byte(configYAML), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		db, err := sqlite.Open(filepath.Join(workspacePath, "memory.db"))
		if err != nil {
			t.Fatalf("failed to open database: %v", err)
		}
		defer db.Close()

		if err := migrations.RunUp(db); err != nil {
			t.Fatalf("failed to run migrations: %v", err)
		}

		output, err := executeContextCommand("generate")
		if err != nil {
			t.Fatalf("unexpected error executing generate: %v", err)
		}

		expected := "No repository context available.\n"
		if output != expected {
			t.Errorf("expected output %q, got %q", expected, output)
		}
	})

	t.Run("Successful context generation", func(t *testing.T) {
		os.RemoveAll(tempDir)
		os.MkdirAll(tempDir, 0755)

		// Initialize Git repository so discovery succeeds
		gitInit := exec.Command("git", "init")
		gitInit.Dir = tempDir
		if err := gitInit.Run(); err != nil {
			t.Fatalf("failed to init git: %v", err)
		}

		if err := os.MkdirAll(workspacePath, 0755); err != nil {
			t.Fatalf("failed to create workspace dir: %v", err)
		}
		configYAML := "repository:\n  path: " + tempDir + "\nstorage:\n  sqlite_path: " + filepath.Join(workspacePath, "memory.db") + "\nai:\n  provider: none\n"
		if err := os.WriteFile(filepath.Join(workspacePath, "config.yaml"), []byte(configYAML), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		db, err := sqlite.Open(filepath.Join(workspacePath, "memory.db"))
		if err != nil {
			t.Fatalf("failed to open database: %v", err)
		}
		defer db.Close()

		if err := migrations.RunUp(db); err != nil {
			t.Fatalf("failed to run migrations: %v", err)
		}

		// Calculate repo ID using GitDiscovery logic
		absPath, err := filepath.Abs(tempDir)
		if err != nil {
			t.Fatalf("failed to get absolute path: %v", err)
		}

		// Let's perform git discovery directly to get the exact ID
		discovery := repository.NewGitDiscovery()
		repo, err := discovery.Discover(stdcontext.Background(), absPath)
		if err != nil {
			t.Fatalf("failed to run git discovery: %v", err)
		}
		repoID := repo.ID

		// Seed repository metadata
		_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repoID, "Test Repo", absPath, "main")
		if err != nil {
			t.Fatalf("failed to insert repository: %v", err)
		}

		// Seed source metadata
		_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_1", repoID, "adr", "docs/adr/0001.md", "Author", "2024-01-01T00:00:00Z")
		if err != nil {
			t.Fatalf("failed to insert source: %v", err)
		}

		// Seed Decision
		decisionStore := memorystorage.NewSQLiteDecisionStore(db)
		err = decisionStore.UpsertDecision(stdcontext.Background(), &memorymodels.Decision{
			ID:           "dec_1",
			RepositoryID: repoID,
			Title:        "Use Cache Engine",
			Status:       "Accepted",
			SourceID:     "src_1",
			CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		})
		if err != nil {
			t.Fatalf("failed to write decision: %v", err)
		}

		output, err := executeContextCommand("generate")
		if err != nil {
			t.Fatalf("unexpected error executing generate: %v", err)
		}

		// Verify output markdown structure
		expectedSubstrings := []string{
			"# Repository Context",
			"Repository: " + repoID,
			"## Key Decisions",
			"* Use Cache Engine",
		}

		for _, sub := range expectedSubstrings {
			if !strings.Contains(output, sub) {
				t.Errorf("expected output to contain %q, got:\n%s", sub, output)
			}
		}
	})
}

func TestContextExportCommand(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-context-export-cmd-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origWorkspace := os.Getenv("REPONERVE_WORKSPACE")
	defer func() {
		if origWorkspace != "" {
			os.Setenv("REPONERVE_WORKSPACE", origWorkspace)
		} else {
			os.Unsetenv("REPONERVE_WORKSPACE")
		}
	}()

	workspacePath := filepath.Join(tempDir, ".reponerve")
	os.Setenv("REPONERVE_WORKSPACE", workspacePath)

	t.Run("Command registration", func(t *testing.T) {
		cmd := NewCommand()
		exportCmd, _, err := cmd.Find([]string{"export"})
		if err != nil {
			t.Fatalf("failed to find export subcommand: %v", err)
		}
		if exportCmd.Name() != "export" {
			t.Errorf("expected subcommand name 'export', got %q", exportCmd.Name())
		}
	})

	t.Run("Empty repository (no context) returns error", func(t *testing.T) {
		os.RemoveAll(tempDir)
		os.MkdirAll(tempDir, 0755)

		// Initialize Git repository so discovery succeeds
		gitInit := exec.Command("git", "init")
		gitInit.Dir = tempDir
		if err := gitInit.Run(); err != nil {
			t.Fatalf("failed to init git: %v", err)
		}

		if err := os.MkdirAll(workspacePath, 0755); err != nil {
			t.Fatalf("failed to create workspace dir: %v", err)
		}
		configYAML := "repository:\n  path: " + tempDir + "\nstorage:\n  sqlite_path: " + filepath.Join(workspacePath, "memory.db") + "\nai:\n  provider: none\n"
		if err := os.WriteFile(filepath.Join(workspacePath, "config.yaml"), []byte(configYAML), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		db, err := sqlite.Open(filepath.Join(workspacePath, "memory.db"))
		if err != nil {
			t.Fatalf("failed to open database: %v", err)
		}
		defer db.Close()

		if err := migrations.RunUp(db); err != nil {
			t.Fatalf("failed to run migrations: %v", err)
		}

		outputPath := filepath.Join(tempDir, "export-empty.md")
		_, err = executeContextCommand("export", "--output", outputPath)
		if err == nil {
			t.Fatal("expected error on exporting empty context, got nil")
		}
		if !strings.Contains(err.Error(), "no repository context available to export") {
			t.Errorf("expected empty context error, got: %v", err)
		}
	})

	t.Run("Successful context export", func(t *testing.T) {
		os.RemoveAll(tempDir)
		os.MkdirAll(tempDir, 0755)

		// Initialize Git repository so discovery succeeds
		gitInit := exec.Command("git", "init")
		gitInit.Dir = tempDir
		if err := gitInit.Run(); err != nil {
			t.Fatalf("failed to init git: %v", err)
		}

		if err := os.MkdirAll(workspacePath, 0755); err != nil {
			t.Fatalf("failed to create workspace dir: %v", err)
		}
		configYAML := "repository:\n  path: " + tempDir + "\nstorage:\n  sqlite_path: " + filepath.Join(workspacePath, "memory.db") + "\nai:\n  provider: none\n"
		if err := os.WriteFile(filepath.Join(workspacePath, "config.yaml"), []byte(configYAML), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		db, err := sqlite.Open(filepath.Join(workspacePath, "memory.db"))
		if err != nil {
			t.Fatalf("failed to open database: %v", err)
		}
		defer db.Close()

		if err := migrations.RunUp(db); err != nil {
			t.Fatalf("failed to run migrations: %v", err)
		}

		// Calculate repo ID using GitDiscovery logic
		absPath, err := filepath.Abs(tempDir)
		if err != nil {
			t.Fatalf("failed to get absolute path: %v", err)
		}

		discovery := repository.NewGitDiscovery()
		repo, err := discovery.Discover(stdcontext.Background(), absPath)
		if err != nil {
			t.Fatalf("failed to run git discovery: %v", err)
		}
		repoID := repo.ID

		// Seed repository metadata
		_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repoID, "Test Repo", absPath, "main")
		if err != nil {
			t.Fatalf("failed to insert repository: %v", err)
		}

		// Seed source metadata
		_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_1", repoID, "adr", "docs/adr/0001.md", "Author", "2024-01-01T00:00:00Z")
		if err != nil {
			t.Fatalf("failed to insert source: %v", err)
		}

		// Seed Decision
		decisionStore := memorystorage.NewSQLiteDecisionStore(db)
		err = decisionStore.UpsertDecision(stdcontext.Background(), &memorymodels.Decision{
			ID:           "dec_1",
			RepositoryID: repoID,
			Title:        "Use Cache Engine",
			Status:       "Accepted",
			SourceID:     "src_1",
			CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		})
		if err != nil {
			t.Fatalf("failed to write decision: %v", err)
		}

		outputPath := filepath.Join(tempDir, "exported-context.md")
		output, err := executeContextCommand("export", "--output", outputPath)
		if err != nil {
			t.Fatalf("unexpected error executing export: %v", err)
		}

		// Verify CLI output
		expectedCLI := "✓ Repository context exported to " + outputPath + "\n"
		if output != expectedCLI {
			t.Errorf("expected CLI output %q, got %q", expectedCLI, output)
		}

		// Verify exported file content
		data, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("failed to read exported file: %v", err)
		}
		content := string(data)
		if !strings.Contains(content, "# Repository Context") || !strings.Contains(content, "Use Cache Engine") {
			t.Errorf("exported file content is incorrect:\n%s", content)
		}
	})

	t.Run("Unsupported format", func(t *testing.T) {
		_, err := executeContextCommand("export", "--format", "json")
		if err == nil {
			t.Fatal("expected error on unsupported format, got nil")
		}
		if !strings.Contains(err.Error(), "unsupported format") {
			t.Errorf("expected error message to contain 'unsupported format', got: %v", err)
		}
	})
}

