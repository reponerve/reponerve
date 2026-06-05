package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigInitializeAndLoad(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	workspaceDir := filepath.Join(tempDir, ".reponerve")

	cfg, err := Initialize(workspaceDir)
	if err != nil {
		t.Fatalf("failed to initialize config: %v", err)
	}

	if cfg.Repository.Path != "." {
		t.Errorf("expected repository path to be '.', got %q", cfg.Repository.Path)
	}
	expectedDBPath := filepath.Join(workspaceDir, "memory.db")
	if cfg.Storage.SQLitePath != expectedDBPath {
		t.Errorf("expected sqlite path to be %q, got %q", expectedDBPath, cfg.Storage.SQLitePath)
	}
	if cfg.AI.Provider != "none" {
		t.Errorf("expected AI provider to be 'none', got %q", cfg.AI.Provider)
	}

	configPath := filepath.Join(workspaceDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("config file was not created: %v", err)
	}

	loadedCfg, err := Load(workspaceDir)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loadedCfg.Repository.Path != "." {
		t.Errorf("expected loaded repository path to be '.', got %q", loadedCfg.Repository.Path)
	}
	if loadedCfg.Storage.SQLitePath != expectedDBPath {
		t.Errorf("expected loaded sqlite path to be %q, got %q", expectedDBPath, loadedCfg.Storage.SQLitePath)
	}
	if loadedCfg.AI.Provider != "none" {
		t.Errorf("expected loaded AI provider to be 'none', got %q", loadedCfg.AI.Provider)
	}
}
