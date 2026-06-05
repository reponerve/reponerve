package scancmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"reponerve/internal/config"
	"reponerve/internal/scanner/adr"
	"reponerve/internal/scanner/git"
	"reponerve/internal/scanner/repository"
	"reponerve/internal/storage/sqlite"
)

// NewCommand creates and returns the scan subcommand.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan the repository to build memory",
		Long:  `Scan repository artifacts (code, git history, PRs, ADRs) to build and update repository memory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceDir := config.GetWorkspaceDir()
			cfg, err := config.Load(workspaceDir)
			if err != nil {
				return fmt.Errorf("workspace not initialized; run 'reponerve init' first")
			}

			db, err := sqlite.Open(cfg.Storage.SQLitePath)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer db.Close()

			cmd.Println("Scanning repository...")

			// 1. Discover repository
			discovery := repository.NewGitDiscovery(db)
			repo, err := discovery.Discover(cmd.Context(), cfg.Repository.Path)
			if err != nil {
				return fmt.Errorf("failed to discover repository: %w", err)
			}

			// 2. Store repository metadata
			err = discovery.Store(cmd.Context(), repo)
			if err != nil {
				return fmt.Errorf("failed to store repository metadata: %w", err)
			}
			cmd.Println("✓ Repository discovered")

			// 3. Execute Git Scanner
			gitScanner := git.NewScanner(db)
			commits, err := gitScanner.Scan(cmd.Context(), repo)
			if err != nil {
				return fmt.Errorf("failed to scan commits: %w", err)
			}
			cmd.Printf("✓ %d commits indexed\n", len(commits))

			// 4. Execute ADR Scanner
			adrScanner := adr.NewScanner(db)
			adrs, err := adrScanner.Scan(cmd.Context(), repo)
			if err != nil {
				return fmt.Errorf("failed to scan ADRs: %w", err)
			}
			cmd.Printf("✓ %d ADRs indexed\n", len(adrs))

			cmd.Println("Scan completed.")
			return nil
		},
	}
}
