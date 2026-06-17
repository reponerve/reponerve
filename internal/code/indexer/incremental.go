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
	changed, err := repositorySourceFilesChangedSince(repositoryPath, state.LastIndexedAt)
	if err != nil {
		return false, err
	}
	return !changed, nil
}

func repositorySourceFilesChangedSince(repositoryPath string, since time.Time) (bool, error) {
	files, err := listAllIndexableFiles(repositoryPath)
	if err != nil {
		return true, err
	}
	for _, marker := range []string{
		"go.mod", "go.work", "package.json", "pyproject.toml", "Cargo.toml",
		"pom.xml", "build.gradle", "build.gradle.kts", "Gemfile", "composer.json",
		"Package.swift", "CMakeLists.txt", "build.sbt", "pubspec.yaml", "mix.exs", "build.zig.zon",
	} {
		markerPath := filepath.Join(repositoryPath, marker)
		if info, err := os.Stat(markerPath); err == nil && info.ModTime().After(since) {
			return true, nil
		}
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
