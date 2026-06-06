package contextcmd

import (
	"github.com/spf13/cobra"
)

// NewCommand creates and returns the context root command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Manage and generate repository context briefings",
		Long:  `Parent command for managing and generating structured repository context briefings for users and AI coding agents.`,
	}

	// Register subcommands
	cmd.AddCommand(newGenerateCommand())
	cmd.AddCommand(newExportCommand())

	return cmd
}
