package handoffcmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/sessionmemory"
	"github.com/reponerve/reponerve/internal/cli/devwire"
	"github.com/reponerve/reponerve/internal/config"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// NewCommand creates the handoff subcommand group.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "handoff",
		Short: "Export or import agent session handoff bundles",
	}
	cmd.AddCommand(newExportCommand(), newImportCommand())
	return cmd
}

func newExportCommand() *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export session memory to a handoff bundle",
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
			bundle, err := svc.ExportHandoff(cmd.Context(), handle.RepositoryID)
			if err != nil {
				return err
			}

			out := output
			if !filepath.IsAbs(out) {
				out = filepath.Join(cfg.Repository.Path, out)
			}
			if err := sessionmemory.ExportHandoffFile(bundle, out); err != nil {
				return err
			}
			cmd.Printf("✓ Exported handoff to %s (%d facts)\n", out, len(bundle.Facts))
			return nil
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "handoff.json", "Output bundle path")
	return cmd
}

func newImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import [bundle.json]",
		Short: "Import a session handoff bundle",
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

			bundle, err := sessionmemory.ImportHandoffFile(args[0])
			if err != nil {
				return err
			}

			svc := devwire.WireSessionMemoryService(db, workspaceDir)
			if err := svc.ImportHandoff(cmd.Context(), bundle); err != nil {
				return err
			}
			cmd.Printf("✓ Imported handoff (%d facts)\n", len(bundle.Facts))
			return nil
		},
	}
	return cmd
}
