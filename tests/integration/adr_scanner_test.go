package integration

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"reponerve/internal/scanner/adr"
	"reponerve/internal/storage/migrations"
	"reponerve/internal/storage/sqlite"
	"reponerve/pkg/models"
)

func TestADRScannerIntegration(t *testing.T) {
	// Create temporary directory for mock repo
	tempDir, err := os.MkdirTemp("", "reponerve-adr-scanner-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create ADR folders
	docsAdrDir := filepath.Join(tempDir, "docs", "adr")
	if err := os.MkdirAll(docsAdrDir, 0755); err != nil {
		t.Fatalf("failed to create docs/adr directory: %v", err)
	}
	plainAdrDir := filepath.Join(tempDir, "adr")
	if err := os.MkdirAll(plainAdrDir, 0755); err != nil {
		t.Fatalf("failed to create adr directory: %v", err)
	}

	// Write mock ADR files
	adr1Path := filepath.Join(docsAdrDir, "0001-use-sqlite.md")
	adr1Content := `# 1. Use SQLite for Local Memory Store

## Status

Accepted

## Context
We need a lightweight database.`
	if err := os.WriteFile(adr1Path, []byte(adr1Content), 0644); err != nil {
		t.Fatalf("failed to write adr1 file: %v", err)
	}

	adr2Path := filepath.Join(plainAdrDir, "0002-viper-config.md")
	adr2Content := `# 2. Viper for Config Management
Status: Proposed

We need to load yaml files.`
	if err := os.WriteFile(adr2Path, []byte(adr2Content), 0644); err != nil {
		t.Fatalf("failed to write adr2 file: %v", err)
	}

	// Set up temporary SQLite database
	dbPath := filepath.Join(tempDir, "test_adr.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite connection: %v", err)
	}
	defer db.Close()

	err = migrations.RunUp(db)
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repoID := "repo_test_adr_scanner"
	repo := &models.Repository{
		ID:            repoID,
		Name:          "test-repo",
		Path:          tempDir,
		DefaultBranch: "main",
	}

	_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repo.ID, repo.Name, repo.Path, repo.DefaultBranch)
	if err != nil {
		t.Fatalf("failed to insert mock repository: %v", err)
	}

	// Initialize ADR scanner and scan
	sourceStore := sqlite.NewSourceStore(db)
	scanner := adr.NewScanner()
	ctx := context.Background()

	sources, err := scanner.Scan(ctx, repo)
	if err != nil {
		t.Fatalf("ADR scan failed: %v", err)
	}

	for _, src := range sources {
		err = sourceStore.UpsertSource(ctx, src)
		if err != nil {
			t.Fatalf("failed to store ADR source: %v", err)
		}
	}

	if len(sources) != 2 {
		t.Fatalf("expected 2 discovered ADR sources, got %d", len(sources))
	}

	// Verify sources table in database
	var sourceCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sources WHERE repository_id = ? AND source_type = 'adr'", repo.ID).Scan(&sourceCount)
	if err != nil {
		t.Fatalf("failed to query sources count: %v", err)
	}
	if sourceCount != 2 {
		t.Errorf("expected 2 source records of type 'adr' in database, got %d", sourceCount)
	}

	// Retrieve one source from database and verify metadata
	var dbID, dbTitle, dbReference, dbMetadataJSON string
	err = db.QueryRow("SELECT id, title, reference, metadata_json FROM sources WHERE reference = 'docs/adr/0001-use-sqlite.md'").Scan(&dbID, &dbTitle, &dbReference, &dbMetadataJSON)
	if err != nil {
		t.Fatalf("failed to query source record from db: %v", err)
	}

	if dbTitle != "1. Use SQLite for Local Memory Store" {
		t.Errorf("expected title to be '1. Use SQLite for Local Memory Store', got %q", dbTitle)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(dbMetadataJSON), &metadata); err != nil {
		t.Fatalf("failed to parse metadata_json: %v", err)
	}

	if metadata["status"] != "Accepted" {
		t.Errorf("expected status 'Accepted' in metadata, got %q", metadata["status"])
	}
	if metadata["content"] != adr1Content {
		t.Errorf("expected correct content in metadata")
	}

	// Verify the second one as well
	err = db.QueryRow("SELECT title, metadata_json FROM sources WHERE reference = 'adr/0002-viper-config.md'").Scan(&dbTitle, &dbMetadataJSON)
	if err != nil {
		t.Fatalf("failed to query adr2 from db: %v", err)
	}
	if dbTitle != "2. Viper for Config Management" {
		t.Errorf("expected title to be '2. Viper for Config Management', got %q", dbTitle)
	}

	var metadata2 map[string]interface{}
	if err := json.Unmarshal([]byte(dbMetadataJSON), &metadata2); err != nil {
		t.Fatalf("failed to parse metadata_json for adr2: %v", err)
	}
	if metadata2["status"] != "Proposed" {
		t.Errorf("expected status 'Proposed' in metadata, got %q", metadata2["status"])
	}

	// Verify Upsert Behavior: Modifying file and scanning again
	time.Sleep(10 * time.Millisecond) // ensure file modification time changes slightly
	newContent := `# 1. Use SQLite for Local Memory Store - Updated

## Status

Approved

## Context
We need a lightweight database with update.`
	if err := os.WriteFile(adr1Path, []byte(newContent), 0644); err != nil {
		t.Fatalf("failed to update adr1 file: %v", err)
	}

	rescanSources, err := scanner.Scan(ctx, repo)
	if err != nil {
		t.Fatalf("ADR scan fail on rescan: %v", err)
	}
	for _, src := range rescanSources {
		err = sourceStore.UpsertSource(ctx, src)
		if err != nil {
			t.Fatalf("failed to store ADR source on rescan: %v", err)
		}
	}
	if len(rescanSources) != 2 {
		t.Fatalf("expected still 2 sources on rescan, got %d", len(rescanSources))
	}

	// Verify total count in DB remains 2 (no duplicates)
	err = db.QueryRow("SELECT COUNT(*) FROM sources WHERE repository_id = ? AND source_type = 'adr'", repo.ID).Scan(&sourceCount)
	if err != nil {
		t.Fatalf("failed to query sources count on rescan: %v", err)
	}
	if sourceCount != 2 {
		t.Errorf("expected 2 source records in database after rescan, got %d", sourceCount)
	}

	// Verify that title and status are updated in DB
	err = db.QueryRow("SELECT title, metadata_json FROM sources WHERE reference = 'docs/adr/0001-use-sqlite.md'").Scan(&dbTitle, &dbMetadataJSON)
	if err != nil {
		t.Fatalf("failed to query updated source: %v", err)
	}
	if dbTitle != "1. Use SQLite for Local Memory Store - Updated" {
		t.Errorf("expected updated title, got %q", dbTitle)
	}

	var metadataUpdated map[string]interface{}
	if err := json.Unmarshal([]byte(dbMetadataJSON), &metadataUpdated); err != nil {
		t.Fatalf("failed to parse updated metadata_json: %v", err)
	}
	if metadataUpdated["status"] != "Approved" {
		t.Errorf("expected updated status 'Approved', got %q", metadataUpdated["status"])
	}
}
