package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/reponerve/reponerve/internal/storage"
	"github.com/reponerve/reponerve/pkg/models"
)

// Scanner provides functionality to discover Git commits.
type Scanner struct {
	scanStateStore storage.ScanStateStore
}

// NewScanner creates a new Scanner instance.
func NewScanner(scanStateStore storage.ScanStateStore) *Scanner {
	return &Scanner{scanStateStore: scanStateStore}
}

// Scan extracts new commits starting from the last scanned commit.
// It returns the list of new source records scanned.
func (s *Scanner) Scan(ctx context.Context, repo *models.Repository) ([]*models.Source, error) {
	headCmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	headCmd.Dir = repo.Path
	headOut, err := headCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD commit: %w", err)
	}
	headHash := strings.TrimSpace(string(headOut))

	var state *storage.ScanState
	if s.scanStateStore != nil {
		var err error
		state, err = s.scanStateStore.GetScanState(ctx, repo.ID)
		if err != nil {
			return nil, err
		}
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

	gitArgs = append(gitArgs, "--pretty=format:%H%n%an <%ae>%n%ad%n%B%x00", "--date=iso-strict")

	cmd := exec.CommandContext(ctx, "git", gitArgs...)
	cmd.Dir = repo.Path
	out, err := cmd.Output()
	if err != nil {
		if state != nil {
			cmd = exec.CommandContext(ctx, "git", "log", "HEAD", "--pretty=format:%H%n%an <%ae>%n%ad%n%B%x00", "--date=iso-strict")
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

	return sources, nil
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
