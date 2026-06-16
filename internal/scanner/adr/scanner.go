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

type scanTarget struct {
	dir        string
	sourceType string
	idPrefix   string
}

// Scan discovers ADR and architecture markdown files under supported directories.
func (s *Scanner) Scan(ctx context.Context, repo *models.Repository) ([]*models.Source, error) {
	var sources []*models.Source

	targets := []scanTarget{
		{dir: filepath.Join(repo.Path, "docs", "adr"), sourceType: "adr", idPrefix: "adr_"},
		{dir: filepath.Join(repo.Path, "docs", "adrs"), sourceType: "adr", idPrefix: "adr_"},
		{dir: filepath.Join(repo.Path, "adr"), sourceType: "adr", idPrefix: "adr_"},
		{dir: filepath.Join(repo.Path, "adrs"), sourceType: "adr", idPrefix: "adr_"},
		{dir: filepath.Join(repo.Path, "docs", "architecture"), sourceType: "architecture_doc", idPrefix: "archdoc_"},
	}

	for _, target := range targets {
		batch, err := s.scanDirectory(ctx, repo, target)
		if err != nil {
			return nil, err
		}
		sources = append(sources, batch...)
	}

	return sources, nil
}

func (s *Scanner) scanDirectory(ctx context.Context, repo *models.Repository, target scanTarget) ([]*models.Source, error) {
	if _, err := os.Stat(target.dir); err != nil {
		return nil, nil
	}

	var sources []*models.Source
	err := filepath.Walk(target.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			return nil
		}

		contentBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}
		content := string(contentBytes)

		relPath, err := filepath.Rel(repo.Path, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}
		relPath = filepath.ToSlash(relPath)

		title, status := ParseADR(content)
		if title == "" {
			base := filepath.Base(path)
			title = strings.TrimSuffix(base, filepath.Ext(base))
		}

		hashInput := repo.ID + ":" + relPath
		h := sha256.Sum256([]byte(hashInput))
		id := fmt.Sprintf("%s%s", target.idPrefix, hex.EncodeToString(h[:]))

		metadata := map[string]interface{}{
			"content": content,
			"status":  status,
			"path":    relPath,
			"kind":    target.sourceType,
		}
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata for %s: %w", path, err)
		}

		sources = append(sources, &models.Source{
			ID:           id,
			RepositoryID: repo.ID,
			SourceType:   target.sourceType,
			Reference:    relPath,
			Title:        title,
			Author:       "",
			Timestamp:    info.ModTime(),
			MetadataJSON: string(metadataBytes),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return sources, nil
}
