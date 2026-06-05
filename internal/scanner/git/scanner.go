package git

import (
	"context"
	"database/sql"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"reponerve/internal/storage/sqlite"
	"reponerve/pkg/models"
)

// ScanState holds the scan state for a repository.
type ScanState struct {
	RepositoryID   string
	LastScanCommit string
	UpdatedAt      time.Time
}

// Scanner provides functionality to discover and ingest Git commits.
type Scanner struct {
	db *sqlite.Database
}

// NewScanner creates a new Scanner instance.
func NewScanner(db *sqlite.Database) *Scanner {
	return &Scanner{db: db}
}

// GetScanState retrieves the scan state for a repository.
func (s *Scanner) GetScanState(ctx context.Context, repoID string) (*ScanState, error) {
	var state ScanState
	query := "SELECT repository_id, last_scan_commit, updated_at FROM scan_state WHERE repository_id = ?"
	err := s.db.QueryRowContext(ctx, query, repoID).Scan(&state.RepositoryID, &state.LastScanCommit, &state.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to query scan state: %w", err)
	}
	return &state, nil
}

// UpdateScanState stores or updates the scan state.
func (s *Scanner) UpdateScanState(ctx context.Context, repoID string, commitHash string) error {
	query := `
		INSERT INTO scan_state (repository_id, last_scan_commit, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(repository_id) DO UPDATE SET
			last_scan_commit = excluded.last_scan_commit,
			updated_at = excluded.updated_at
	`
	_, err := s.db.ExecContext(ctx, query, repoID, commitHash, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update scan state: %w", err)
	}
	return nil
}

// Scan extracts and stores new commits starting from the last scanned commit.
// It returns the list of new source records scanned.
func (s *Scanner) Scan(ctx context.Context, repo *models.Repository) ([]*models.Source, error) {
	headCmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	headCmd.Dir = repo.Path
	headOut, err := headCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD commit: %w", err)
	}
	headHash := strings.TrimSpace(string(headOut))

	state, err := s.GetScanState(ctx, repo.ID)
	if err != nil {
		return nil, err
	}

	var gitArgs []string
	gitArgs = append(gitArgs, "log")

	if state != nil && state.LastScanCommit != "" {
		if state.LastScanCommit == headHash {
			return nil, nil
		}
		gitArgs = append(gitArgs, fmt.Sprintf("%s..HEAD", state.LastScanCommit))
	} else {
		gitArgs = append(gitArgs, "HEAD")
	}

	gitArgs = append(gitArgs, "--pretty=format:%H%n%an%n%ad%n%B%x00", "--date=iso-strict")

	cmd := exec.CommandContext(ctx, "git", gitArgs...)
	cmd.Dir = repo.Path
	out, err := cmd.Output()
	if err != nil {
		if state != nil {
			cmd = exec.CommandContext(ctx, "git", "log", "HEAD", "--pretty=format:%H%n%an%n%ad%n%B%x00", "--date=iso-strict")
			cmd.Dir = repo.Path
			out, err = cmd.Output()
			if err != nil {
				return nil, fmt.Errorf("failed to scan commits during fallback: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to scan commits: %w", err)
		}
	}

	sources, err := s.ParseGitLog(repo.ID, string(out))
	if err != nil {
		return nil, fmt.Errorf("failed to parse git log: %w", err)
	}

	for _, src := range sources {
		err := s.storeSource(ctx, src)
		if err != nil {
			return nil, fmt.Errorf("failed to store source commit %s: %w", src.ID, err)
		}
	}

	err = s.UpdateScanState(ctx, repo.ID, headHash)
	if err != nil {
		return nil, err
	}

	return sources, nil
}

func (s *Scanner) storeSource(ctx context.Context, src *models.Source) error {
	query := `
		INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, metadata_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			author = excluded.author,
			timestamp = excluded.timestamp,
			metadata_json = excluded.metadata_json
	`
	now := time.Now()
	_, err := s.db.ExecContext(ctx, query, src.ID, src.RepositoryID, src.SourceType, src.Reference, src.Title, src.Author, src.Timestamp, nil, now)
	return err
}

// ParseGitLog parses raw git log outputs into structured models.Source pointers.
func (s *Scanner) ParseGitLog(repoID string, output string) ([]*models.Source, error) {
	var sources []*models.Source
	chunks := strings.Split(output, "\x00")
	for _, chunk := range chunks {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}
		lines := strings.SplitN(chunk, "\n", 4)
		if len(lines) < 3 {
			continue
		}
		hash := strings.TrimSpace(lines[0])
		author := strings.TrimSpace(lines[1])
		dateStr := strings.TrimSpace(lines[2])
		message := ""
		if len(lines) == 4 {
			message = strings.TrimSpace(lines[3])
		}

		t, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			t, err = time.Parse("2006-01-02 15:04:05 -0700", dateStr)
			if err != nil {
				t = time.Now()
			}
		}

		title := strings.TrimSpace(message)

		sources = append(sources, &models.Source{
			ID:           hash,
			RepositoryID: repoID,
			SourceType:   "commit",
			Reference:    hash,
			Title:        title,
			Author:       author,
			Timestamp:    t,
		})
	}
	return sources, nil
}
