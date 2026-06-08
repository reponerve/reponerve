package mcpcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

func executeMcpCommand(ctx context.Context, stdin *bytes.Buffer, args ...string) (string, error) {
	cmd := NewCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetIn(stdin)
	cmd.SetArgs(args)

	err := cmd.ExecuteContext(ctx)
	return buf.String(), err
}

func TestMcpCommandRegistration(t *testing.T) {
	cmd := NewCommand()
	if cmd.Use != "mcp" {
		t.Errorf("expected command Use 'mcp', got %q", cmd.Use)
	}
	if !strings.Contains(cmd.Short, "Start the RepoNerve MCP Server") {
		t.Errorf("expected command short description to contain MCP Server, got %q", cmd.Short)
	}
}

func TestMcpCommandMissingWorkspace(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-mcp-cmd-test-*")
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

	stdin := new(bytes.Buffer)
	_, err = executeMcpCommand(context.Background(), stdin)
	if err == nil {
		t.Fatal("expected error on missing workspace, got nil")
	}
	if !strings.Contains(err.Error(), "workspace not initialized") {
		t.Errorf("expected error message to contain 'workspace not initialized', got: %v", err)
	}
}

func TestMcpCommandHandshakeSuccess(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-mcp-cmd-test-*")
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

	// Prepare standard initialize request JSON-RPC
	initReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0",
			},
		},
	}
	reqBytes, err := json.Marshal(initReq)
	if err != nil {
		t.Fatalf("failed to marshal init request: %v", err)
	}

	stdin := bytes.NewBuffer(append(reqBytes, '\n'))

	output, err := executeMcpCommand(context.Background(), stdin)
	if err != nil {
		t.Fatalf("unexpected error executing mcp command: %v", err)
	}

	if !strings.Contains(output, `"jsonrpc":"2.0"`) {
		t.Errorf("expected output to contain jsonrpc 2.0 response, got %q", output)
	}
	if !strings.Contains(output, `"reponerve"`) {
		t.Errorf("expected output to contain server name 'reponerve', got %q", output)
	}
}
