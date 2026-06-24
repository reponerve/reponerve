package initcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/config"
	"github.com/reponerve/reponerve/internal/integration"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// NewCommand creates and returns the init subcommand.
func NewCommand() *cobra.Command {
	var skipIDE bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize RepoNerve inside a repository",
		Long: `Initialize a new RepoNerve workspace, config file, and database in the current repository.

Also installs IDE integration automatically: Cursor skill + MCP, Native Development Discipline
rules (coding guidelines + plan/review habits), VS Code Copilot MCP, and Continue MCP configuration.`,
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

			if !skipIDE {
				result, err := integration.Install(integration.Options{GlobalSkill: true})
				if err != nil {
					return fmt.Errorf("failed to install IDE integration: %w", err)
				}
				for _, line := range integration.FormatSummary(result) {
					cmd.Println(line)
				}
			}

			cmd.Println("✓ RepoNerve ready")
			cmd.Println("  → Run: reponerve scan")

			return nil
		},
	}

	cmd.Flags().BoolVar(&skipIDE, "skip-ide", false, "Skip automatic Cursor skill and MCP installation")

	return cmd
}
