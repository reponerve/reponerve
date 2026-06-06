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
	runGitCommand(t, tempDir, "commit", "-m", "feat: initial repository commit to optimize storage")
	runGitCommand(t, tempDir, "commit", "--allow-empty", "-m", "feat: adopt Go language in production")
	runGitCommand(t, tempDir, "branch", "-M", "main")

	// Create an ADR
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

	output := scanBuf.String()
	expectedLines := []string{
		"Scanning repository...",
		"✓ Repository discovered",
		"✓ 2 commits indexed",
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
	if commitCount != 2 {
		t.Errorf("expected 2 commits in db, got %d", commitCount)
	}

	var adrCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sources WHERE source_type = 'adr'").Scan(&adrCount)
	if err != nil {
		t.Fatalf("failed to query adr count: %v", err)
	}
	if adrCount != 1 {
		t.Errorf("expected 1 ADR in db, got %d", adrCount)
	}

	// Verify memory_events content
	var eventCount int
	err = db.QueryRow("SELECT COUNT(*) FROM memory_events").Scan(&eventCount)
	if err != nil {
		t.Fatalf("failed to query memory_events count: %v", err)
	}
	if eventCount != 2 {
		t.Errorf("expected 2 events in memory_events, got %d", eventCount)
	}

	// Verify memory_decisions content
	var decisionCount int
	err = db.QueryRow("SELECT COUNT(*) FROM memory_decisions").Scan(&decisionCount)
	if err != nil {
		t.Fatalf("failed to query memory_decisions count: %v", err)
	}
	if decisionCount != 1 {
		t.Errorf("expected 1 decision in memory_decisions, got %d", decisionCount)
	}

	var decTitle, decStatus string
	err = db.QueryRow("SELECT title, status FROM memory_decisions").Scan(&decTitle, &decStatus)
	if err != nil {
		t.Fatalf("failed to query memory_decision fields: %v", err)
	}
	if decTitle != "1. Use Go" {
		t.Errorf("expected title '1. Use Go', got %q", decTitle)
	}
	if decStatus != "Accepted" {
		t.Errorf("expected status 'Accepted', got %q", decStatus)
	}

	// Verify memory_intents content
	var intentCount int
	err = db.QueryRow("SELECT COUNT(*) FROM memory_intents").Scan(&intentCount)
	if err != nil {
		t.Fatalf("failed to query memory_intents count: %v", err)
	}
	if intentCount != 3 {
		t.Errorf("expected 3 intents in memory_intents, got %d", intentCount)
	}

	// Verify specific descriptions are present
	rows, err := db.Query("SELECT description FROM memory_intents ORDER BY description ASC")
	if err != nil {
		t.Fatalf("failed to query memory_intents: %v", err)
	}
	defer rows.Close()

	var descs []string
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			t.Fatalf("failed to scan description: %v", err)
		}
		descs = append(descs, d)
	}

	if len(descs) != 3 {
		t.Fatalf("expected 3 description rows, got %d", len(descs))
	}
	if descs[0] != "Optimize Deployment" {
		t.Errorf("expected 'Optimize Deployment', got %q", descs[0])
	}
	if descs[1] != "Optimize Storage" {
		t.Errorf("expected 'Optimize Storage', got %q", descs[1])
	}
	if descs[2] != "Simplify Configuration" {
		t.Errorf("expected 'Simplify Configuration', got %q", descs[2])
	}

	// Verify memory_facts content
	var factCount int
	err = db.QueryRow("SELECT COUNT(*) FROM memory_facts").Scan(&factCount)
	if err != nil {
		t.Fatalf("failed to query memory_facts count: %v", err)
	}
	if factCount != 2 {
		t.Errorf("expected 2 facts in memory_facts, got %d", factCount)
	}

	// Verify memory_relationships content
	var relCount int
	err = db.QueryRow("SELECT COUNT(*) FROM memory_relationships").Scan(&relCount)
	if err != nil {
		t.Fatalf("failed to query memory_relationships count: %v", err)
	}
	if relCount != 4 {
		t.Errorf("expected 4 relationships in memory_relationships, got %d", relCount)
	}

	// Verify specific relationship types are represented
	typeCounts := make(map[string]int)
	relRows, err := db.Query("SELECT relationship_type FROM memory_relationships")
	if err != nil {
		t.Fatalf("failed to query relationship types: %v", err)
	}
	defer relRows.Close()

	for relRows.Next() {
		var tName string
		if err := relRows.Scan(&tName); err != nil {
			t.Fatalf("failed to scan relationship type: %v", err)
		}
		typeCounts[tName]++
	}

	if typeCounts["INTENT_DRIVES_DECISION"] != 2 {
		t.Errorf("expected 2 INTENT_DRIVES_DECISION, got %d", typeCounts["INTENT_DRIVES_DECISION"])
	}
	if typeCounts["DECISION_RESULTS_IN_EVENT"] != 1 {
		t.Errorf("expected 1 DECISION_RESULTS_IN_EVENT, got %d", typeCounts["DECISION_RESULTS_IN_EVENT"])
	}
	if typeCounts["FACT_SUPPORTS_DECISION"] != 1 {
		t.Errorf("expected 1 FACT_SUPPORTS_DECISION, got %d", typeCounts["FACT_SUPPORTS_DECISION"])
	}
}
