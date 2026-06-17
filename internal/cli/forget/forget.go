package forgetcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/cli/devwire"
	"github.com/reponerve/reponerve/internal/config"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// NewCommand creates the forget subcommand.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forget [fact_id]",
		Short: "Remove a session memory fact",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceDir := config.GetWorkspaceDir()
			cfg, err := config.Load(workspaceDir)
			if err != nil {
				return fmt.Errorf("%s", config.FormatLoadError(workspaceDir, err))
			}
			db, err := sqlite.Open(cfg.Storage.SQLitePath)
			if err != nil {
				return err
			}
			defer db.Close()

			handle, err := devwire.Open(cmd.Context(), workspaceDir)
			if err != nil {
				return err
			}
			defer handle.Close()

			svc := devwire.WireSessionMemoryService(db, workspaceDir)
			if err := svc.Forget(cmd.Context(), handle.RepositoryID, args[0]); err != nil {
				return err
			}
			cmd.Printf("✓ Forgot %s\n", args[0])
			return nil
		},
	}
	return cmd
}
