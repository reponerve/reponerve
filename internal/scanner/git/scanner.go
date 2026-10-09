package git

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

	gitArgs = append(gitArgs, "--pretty=format:%x1e%H%x1f%an <%ae>%x1f%ad%x1f%B%x1f", "--date=iso-strict", "--name-only")

	runGit := func(args []string) ([]*models.Source, error) {
		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Dir = repo.Path
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		sources, err := s.StreamGitLog(repo.ID, stdout)
		if err != nil {
			_ = cmd.Wait()
			return nil, err
		}
		if err := cmd.Wait(); err != nil {
			return nil, err
		}
		return sources, nil
	}

	sources, err := runGit(gitArgs)
	if err != nil {
		if state != nil {
			fallbackArgs := []string{"log", "HEAD", "--pretty=format:%x1e%H%x1f%an <%ae>%x1f%ad%x1f%B%x1f", "--date=iso-strict", "--name-only"}
			sources, err = runGit(fallbackArgs)
			if err != nil {
				return nil, fmt.Errorf("failed to scan commits during fallback: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to scan commits: %w", err)
		}
	}

	return sources, nil
}

func splitCommitRecord(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	for i, b := range data {
		if b == 0x1e || b == 0x00 {
			return i + 1, data[0:i], nil
		}
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

// StreamGitLog streams raw git log outputs into structured models.Source pointers.
func (s *Scanner) StreamGitLog(repoID string, r io.Reader) ([]*models.Source, error) {
	var sources []*models.Source
	scanner := bufio.NewScanner(r)
	scanner.Split(splitCommitRecord)
	// Allocate 64KB initial buffer, up to 16MB max to handle large commit messages
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)

	for scanner.Scan() {
		chunk := strings.TrimSpace(scanner.Text())
		if chunk == "" {
			continue
		}
		var hash, author, dateStr, message string
		var files []string

		if strings.Contains(chunk, "\x1f") {
			parts := strings.Split(chunk, "\x1f")
			if len(parts) >= 3 {
				hash = strings.TrimSpace(parts[0])
				author = strings.TrimSpace(parts[1])
				dateStr = strings.TrimSpace(parts[2])
			}
			if len(parts) >= 4 {
				message = strings.TrimSpace(parts[3])
			}
			if len(parts) >= 5 {
				for _, line := range strings.Split(parts[4], "\n") {
					line = strings.TrimSpace(line)
					if line != "" {
						files = append(files, line)
					}
				}
			}
		} else {
			lines := strings.SplitN(chunk, "\n", 4)
			if len(lines) < 3 {
				continue
			}
			hash = strings.TrimSpace(lines[0])
			author = strings.TrimSpace(lines[1])
			dateStr = strings.TrimSpace(lines[2])
			if len(lines) == 4 {
				message = strings.TrimSpace(lines[3])
			}
		}

		t, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			t, err = time.Parse("2006-01-02 15:04:05 -0700", dateStr)
			if err != nil {
				t = time.Now()
			}
		}

		title := strings.TrimSpace(message)

		var metadataJSON string
		if len(files) > 0 {
			if metaBytes, err := json.Marshal(map[string]any{"files": files}); err == nil {
				metadataJSON = string(metaBytes)
			}
		}

		sources = append(sources, &models.Source{
			ID:           hash,
			RepositoryID: repoID,
			SourceType:   "commit",
			Reference:    hash,
			Title:        title,
			Author:       author,
			Timestamp:    t,
			MetadataJSON: metadataJSON,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan git log: %w", err)
	}
	return sources, nil
}

// ParseGitLog parses raw git log outputs into structured models.Source pointers.
func (s *Scanner) ParseGitLog(repoID string, output string) ([]*models.Source, error) {
	return s.StreamGitLog(repoID, strings.NewReader(output))
}
