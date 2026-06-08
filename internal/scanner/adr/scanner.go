package adr

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/reponerve/reponerve/pkg/models"
)

// Scanner provides functionality to discover and parse Architecture Decision Records (ADRs).
type Scanner struct {
}

// NewScanner creates a new Scanner instance.
func NewScanner() *Scanner {
	return &Scanner{}
}

// ParseADR parses the raw markdown content to extract the main title and status.
func ParseADR(content string) (title string, status string) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			title = strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
			break
		}
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(trimmed), "status:") {
			status = strings.TrimSpace(trimmed[7:])
			break
		}
		if strings.ToLower(trimmed) == "## status" {
			for j := i + 1; j < len(lines); j++ {
				nextTrimmed := strings.TrimSpace(lines[j])
				if nextTrimmed != "" {
					if strings.HasPrefix(nextTrimmed, "#") {
						break
					}
					status = nextTrimmed
					break
				}
			}
			if status != "" {
				break
			}
		}
	}

	if status == "" {
		status = "Accepted"
	}
	return title, status
}

// Scan discovers ADR markdown files under the supported directories and parses them.
func (s *Scanner) Scan(ctx context.Context, repo *models.Repository) ([]*models.Source, error) {
	var sources []*models.Source

	dirs := []string{
		filepath.Join(repo.Path, "docs", "adr"),
		filepath.Join(repo.Path, "docs", "adrs"),
		filepath.Join(repo.Path, "adr"),
		filepath.Join(repo.Path, "adrs"),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
				return nil
			}

			// Read file content
			contentBytes, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}
			content := string(contentBytes)

			// Get relative path
			relPath, err := filepath.Rel(repo.Path, path)
			if err != nil {
				return fmt.Errorf("failed to get relative path for %s: %w", path, err)
			}

			// Parse content
			title, status := ParseADR(content)
			if title == "" {
				// Fallback to filename without extension
				base := filepath.Base(path)
				title = strings.TrimSuffix(base, filepath.Ext(base))
			}

			// Unique stable ID for the source
			hashInput := repo.ID + ":" + relPath
			h := sha256.Sum256([]byte(hashInput))
			id := fmt.Sprintf("adr_%s", hex.EncodeToString(h[:]))

			// Build ADR metadata
			metadata := map[string]interface{}{
				"content": content,
				"status":  status,
				"path":    relPath,
			}
			metadataBytes, err := json.Marshal(metadata)
			if err != nil {
				return fmt.Errorf("failed to marshal metadata for %s: %w", path, err)
			}

			src := &models.Source{
				ID:           id,
				RepositoryID: repo.ID,
				SourceType:   "adr",
				Reference:    relPath,
				Title:        title,
				Author:       "",
				Timestamp:    info.ModTime(),
				MetadataJSON: string(metadataBytes),
			}

			sources = append(sources, src)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return sources, nil
}
