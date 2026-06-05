package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"reponerve/internal/storage/sqlite"
	"reponerve/pkg/models"
)

// GitDiscovery implements the Discovery interface for Git repositories.
type GitDiscovery struct {
	db *sqlite.Database
}

// NewGitDiscovery creates a new GitDiscovery service.
func NewGitDiscovery(db *sqlite.Database) *GitDiscovery {
	return &GitDiscovery{db: db}
}

// Discover extracts repository metadata for the given path.
func (g *GitDiscovery) Discover(ctx context.Context, path string) (*models.Repository, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Validate git repository existence
	gitDir := filepath.Join(absPath, ".git")
	if stat, err := os.Stat(gitDir); err != nil || !stat.IsDir() {
		return nil, fmt.Errorf("not a valid git repository: %s", absPath)
	}

	name := g.getRepositoryName(absPath)
	defaultBranch := g.getDefaultBranch(absPath)

	hash := sha256.Sum256([]byte(absPath))
	id := fmt.Sprintf("repo_%s", hex.EncodeToString(hash[:6]))

	now := time.Now()
	repo := &models.Repository{
		ID:            id,
		Name:          name,
		Path:          absPath,
		DefaultBranch: defaultBranch,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	return repo, nil
}

// Store persists or updates the repository metadata in the database.
func (g *GitDiscovery) Store(ctx context.Context, repo *models.Repository) error {
	query := `
		INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			path = excluded.path,
			default_branch = excluded.default_branch,
			updated_at = excluded.updated_at
	`
	_, err := g.db.ExecContext(ctx, query, repo.ID, repo.Name, repo.Path, repo.DefaultBranch, repo.CreatedAt, repo.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to store repository metadata: %w", err)
	}
	return nil
}

func (g *GitDiscovery) getRepositoryName(dir string) string {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err == nil {
		url := strings.TrimSpace(string(out))
		if url != "" {
			parts := strings.Split(url, "/")
			if len(parts) > 0 {
				last := parts[len(parts)-1]
				last = strings.TrimSuffix(last, ".git")
				if idx := strings.Index(last, ":"); idx != -1 {
					last = last[idx+1:]
				}
				if last != "" {
					return last
				}
			}
		}
	}
	return filepath.Base(dir)
}

func (g *GitDiscovery) getDefaultBranch(dir string) string {
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err == nil {
		ref := strings.TrimSpace(string(out))
		if strings.HasPrefix(ref, "refs/remotes/origin/") {
			return strings.TrimPrefix(ref, "refs/remotes/origin/")
		}
	}

	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "origin/HEAD")
	cmd.Dir = dir
	out, err = cmd.Output()
	if err == nil {
		ref := strings.TrimSpace(string(out))
		if ref != "" && ref != "origin/HEAD" {
			if strings.HasPrefix(ref, "origin/") {
				return strings.TrimPrefix(ref, "origin/")
			}
			return ref
		}
	}

	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	out, err = cmd.Output()
	if err == nil {
		ref := strings.TrimSpace(string(out))
		if ref != "" && ref != "HEAD" {
			return ref
		}
	}

	cmd = exec.Command("git", "symbolic-ref", "--short", "HEAD")
	cmd.Dir = dir
	out, err = cmd.Output()
	if err == nil {
		ref := strings.TrimSpace(string(out))
		if ref != "" {
			return ref
		}
	}

	return "main"
}
