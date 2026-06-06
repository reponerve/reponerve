package memorycmd

import (
	"github.com/spf13/cobra"
)

// NewCommand creates and returns the memory root command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Repository memory management and queries",
		Long:  `Parent command for querying and listing repository memories (events, decisions, intents, facts, relationships).`,
	}

	// Register subcommands
	cmd.AddCommand(newListCommand())

	return cmd
}
