package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// RepositoryConfig defines configuration for the repository settings.
type RepositoryConfig struct {
	Path string `mapstructure:"path"`
}

// StorageConfig defines configuration for SQLite storage path.
type StorageConfig struct {
	SQLitePath string `mapstructure:"sqlite_path"`
}

// AIConfig defines configuration for the AI settings.
type AIConfig struct {
	Provider string `mapstructure:"provider"`
}

// Config represents the complete application configuration structure.
type Config struct {
	Repository RepositoryConfig `mapstructure:"repository"`
	Storage    StorageConfig    `mapstructure:"storage"`
	AI         AIConfig         `mapstructure:"ai"`
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() *Config {
	return &Config{
		Repository: RepositoryConfig{Path: "."},
		Storage:    StorageConfig{SQLitePath: ".reponerve/memory.db"},
		AI:         AIConfig{Provider: "none"},
	}
}

// Initialize creates the workspace directory and default config.yaml if they do not exist.
func Initialize(workspaceDir string) (*Config, error) {
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspace directory: %w", err)
	}

	configPath := filepath.Join(workspaceDir, "config.yaml")

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		v.Set("repository.path", ".")
		v.Set("storage.sqlite_path", filepath.Join(workspaceDir, "memory.db"))
		v.Set("ai.provider", "none")

		if err := v.WriteConfigAs(configPath); err != nil {
			return nil, fmt.Errorf("failed to write default config: %w", err)
		}
	}

	return Load(workspaceDir)
}

// Load loads configuration from config.yaml within the workspace directory.
func Load(workspaceDir string) (*Config, error) {
	configPath := filepath.Join(workspaceDir, "config.yaml")

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// GetWorkspaceDir returns the workspace directory, defaulting to ".reponerve"
// unless overridden by the REPONERVE_WORKSPACE environment variable.
func GetWorkspaceDir() string {
	if envDir := os.Getenv("REPONERVE_WORKSPACE"); envDir != "" {
		return envDir
	}
	return ".reponerve"
}

