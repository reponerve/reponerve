package versioncmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/version"
)

// NewCommand creates the version subcommand.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print RepoNerve version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), version.String())
		},
	}
}
