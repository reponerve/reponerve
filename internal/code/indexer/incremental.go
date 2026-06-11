package indexer

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/reponerve/reponerve/internal/storage"
)

func shouldSkipIndexing(ctx context.Context, stateStore storage.CodeIndexStateStore, repositoryID, repositoryPath string) (bool, error) {
	if stateStore == nil {
		return false, nil
	}
	state, err := stateStore.GetByRepository(ctx, repositoryID)
	if err != nil || state == nil || state.LastIndexedAt.IsZero() {
		return false, err
	}
	changed, err := repositoryGoFilesChangedSince(repositoryPath, state.LastIndexedAt)
	if err != nil {
		return false, err
	}
	return !changed, nil
}

func repositoryGoFilesChangedSince(repositoryPath string, since time.Time) (bool, error) {
	files, err := listGoFiles(repositoryPath)
	if err != nil {
		return true, err
	}
	goModPath := filepath.Join(repositoryPath, "go.mod")
	if info, err := os.Stat(goModPath); err == nil && info.ModTime().After(since) {
		return true, nil
	}
	workPath := filepath.Join(repositoryPath, "go.work")
	if info, err := os.Stat(workPath); err == nil && info.ModTime().After(since) {
		return true, nil
	}
	for _, rel := range files {
		abs := filepath.Join(repositoryPath, filepath.FromSlash(rel))
		info, err := os.Stat(abs)
		if err != nil {
			return true, err
		}
		if info.ModTime().After(since) {
			return true, nil
		}
	}
	return false, nil
}
