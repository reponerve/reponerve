package remembercmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/sessionmemory"
	"github.com/reponerve/reponerve/internal/cli/devwire"
	"github.com/reponerve/reponerve/internal/config"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// NewCommand creates the remember subcommand.
func NewCommand() *cobra.Command {
	var subject string
	cmd := &cobra.Command{
		Use:   "remember [content]",
		Short: "Store session knowledge as an evidence-backed fact",
		Args:  cobra.MinimumNArgs(1),
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
			content := strings.Join(args, " ")
			if strings.TrimSpace(subject) == "" {
				subject = "session"
			}

			fact, err := svc.Remember(cmd.Context(), sessionmemory.RememberRequest{
				RepositoryID: handle.RepositoryID,
				Subject:      subject,
				Content:      content,
			})
			if err != nil {
				return err
			}
			cmd.Printf("✓ Remembered %s [%s]\n", fact.Subject, fact.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&subject, "subject", "", "Topic subject for the memory fact")
	return cmd
}
