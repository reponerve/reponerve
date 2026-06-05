package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"reponerve/internal/scanner/git"
	"reponerve/internal/storage/migrations"
	"reponerve/internal/storage/sqlite"
	"reponerve/pkg/models"
)

func runGit(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to execute git command %v: %v", args, err)
	}
}

func createTestGitRepository(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "reponerve-git-scanner-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}

	runGit(t, tempDir, "init")
	runGit(t, tempDir, "config", "user.name", "Test User")
	runGit(t, tempDir, "config", "user.email", "test@reponerve.com")

	if err := os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("commit 1"), 0644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}
	runGit(t, tempDir, "add", "file1.txt")
	runGit(t, tempDir, "commit", "-m", "first commit\n\nwith description")

	if err := os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("commit 2"), 0644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}
	runGit(t, tempDir, "add", "file2.txt")
	runGit(t, tempDir, "commit", "-m", "second commit")

	runGit(t, tempDir, "branch", "-M", "main")

	return tempDir
}

func TestGitScannerIntegration(t *testing.T) {
	repoPath := createTestGitRepository(t)
	defer os.RemoveAll(repoPath)

	dbPath := filepath.Join(repoPath, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite connection: %v", err)
	}
	defer db.Close()

	err = migrations.RunUp(db)
	if err != nil {
		t.Fatalf("failed to run database migrations: %v", err)
	}

	ctx := context.Background()
	repoID := "repo_test_git_scanner"

	repo := models.Repository{
		ID:            repoID,
		Name:          "test-repo",
		Path:          repoPath,
		DefaultBranch: "main",
	}

	_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repo.ID, repo.Name, repo.Path, repo.DefaultBranch)
	if err != nil {
		t.Fatalf("failed to insert mock repository: %v", err)
	}

	sourceStore := sqlite.NewSourceStore(db)
	scanStateStore := sqlite.NewScanStateStore(db)
	scanner := git.NewScanner(scanStateStore)

	commits, err := scanner.Scan(ctx, &repo)
	if err != nil {
		t.Fatalf("first scan failed: %v", err)
	}
	if len(commits) != 2 {
		t.Errorf("expected 2 commits from initial scan, got %d", len(commits))
	}

	for _, commit := range commits {
		err = sourceStore.UpsertSource(ctx, commit)
		if err != nil {
			t.Fatalf("failed to store commit: %v", err)
		}
	}
	err = scanStateStore.UpdateScanState(ctx, repo.ID, commits[0].Reference)
	if err != nil {
		t.Fatalf("failed to update scan state: %v", err)
	}

	var sourceCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sources WHERE repository_id = ? AND source_type = 'commit'", repo.ID).Scan(&sourceCount)
	if err != nil {
		t.Fatalf("failed to query sources count: %v", err)
	}
	if sourceCount != 2 {
		t.Errorf("expected 2 source commits in DB, got %d", sourceCount)
	}

	state, err := scanStateStore.GetScanState(ctx, repo.ID)
	if err != nil {
		t.Fatalf("failed to get scan state: %v", err)
	}
	if state == nil || state.LastScanCommit == "" {
		t.Fatalf("expected last scan commit to be set in scan_state")
	}

	if err := os.WriteFile(filepath.Join(repoPath, "file3.txt"), []byte("commit 3"), 0644); err != nil {
		t.Fatalf("failed to write file3: %v", err)
	}
	runGit(t, repoPath, "add", "file3.txt")
	runGit(t, repoPath, "commit", "-m", "third commit")

	commitsInc, err := scanner.Scan(ctx, &repo)
	if err != nil {
		t.Fatalf("incremental scan failed: %v", err)
	}
	if len(commitsInc) != 1 {
		t.Errorf("expected 1 commit in incremental scan, got %d", len(commitsInc))
	}
	if commitsInc[0].Title != "third commit" {
		t.Errorf("expected incremental commit message 'third commit', got %q", commitsInc[0].Title)
	}

	for _, commit := range commitsInc {
		err = sourceStore.UpsertSource(ctx, commit)
		if err != nil {
			t.Fatalf("failed to store commit: %v", err)
		}
	}
	err = scanStateStore.UpdateScanState(ctx, repo.ID, commitsInc[0].Reference)
	if err != nil {
		t.Fatalf("failed to update scan state: %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM sources WHERE repository_id = ? AND source_type = 'commit'", repo.ID).Scan(&sourceCount)
	if err != nil {
		t.Fatalf("failed to query sources count after incremental scan: %v", err)
	}
	if sourceCount != 3 {
		t.Errorf("expected 3 source commits in DB, got %d", sourceCount)
	}

	err = scanStateStore.UpdateScanState(ctx, repo.ID, "fake_commit_hash_non_existent")
	if err != nil {
		t.Fatalf("failed to update scan state with fake hash: %v", err)
	}

	commitsFallback, err := scanner.Scan(ctx, &repo)
	if err != nil {
		t.Fatalf("fallback scan failed: %v", err)
	}
	if len(commitsFallback) != 3 {
		t.Errorf("expected 3 commits from fallback full scan, got %d", len(commitsFallback))
	}

	for _, commit := range commitsFallback {
		err = sourceStore.UpsertSource(ctx, commit)
		if err != nil {
			t.Fatalf("failed to store commit: %v", err)
		}
	}
	err = scanStateStore.UpdateScanState(ctx, repo.ID, commitsFallback[0].Reference)
	if err != nil {
		t.Fatalf("failed to update scan state: %v", err)
	}
}
