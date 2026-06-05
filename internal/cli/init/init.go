package initcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"reponerve/internal/config"
	"reponerve/internal/storage/migrations"
	"reponerve/internal/storage/sqlite"
)

// NewCommand creates and returns the init subcommand.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize RepoNerve inside a repository",
		Long:  `Initialize a new RepoNerve workspace, config file, and database in the current repository.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceDir := config.GetWorkspaceDir()

			cfg, err := config.Initialize(workspaceDir)
			if err != nil {
				return fmt.Errorf("failed to initialize configuration: %w", err)
			}
			cmd.Println("✓ Workspace created")
			cmd.Println("✓ Configuration created")

			db, err := sqlite.Open(cfg.Storage.SQLitePath)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer db.Close()

			if err := migrations.RunUp(db); err != nil {
				return fmt.Errorf("failed to run database migrations: %w", err)
			}
			cmd.Println("✓ Database initialized")
			cmd.Println("✓ RepoNerve ready")

			return nil
		},
	}
}
