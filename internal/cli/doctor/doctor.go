package doctorcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/health"
	"github.com/reponerve/reponerve/internal/cli/devwire"
	"github.com/reponerve/reponerve/internal/config"
	"github.com/reponerve/reponerve/internal/scanner/repository"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// NewCommand creates the doctor subcommand.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check repository memory freshness and workspace health",
		Long:  `Run deterministic freshness checks on workspace, scan state, git HEAD, and code index.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceDir := config.GetWorkspaceDir()
			cfg, err := config.Load(workspaceDir)
			if err != nil {
				return fmt.Errorf("%s", config.FormatLoadError(workspaceDir, err))
			}

			db, err := sqlite.Open(cfg.Storage.SQLitePath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			if err := migrations.RunUp(db); err != nil {
				return fmt.Errorf("run migrations: %w", err)
			}

			checker := health.NewChecker(
				sqlite.NewScanStateStore(db),
				sqlite.NewSQLiteCodeIndexStateStore(db),
				repository.NewGitDiscovery(),
			)
			result, err := checker.Check(cmd.Context(), health.CheckInput{
				WorkspaceDir:   workspaceDir,
				RepositoryPath: cfg.Repository.Path,
			})
			if err != nil {
				return err
			}
			return devwire.WriteDEResult(cmd, health.FormatDoctor(result), result)
		},
	}
	return devwire.BindDECmd(cmd)
}
