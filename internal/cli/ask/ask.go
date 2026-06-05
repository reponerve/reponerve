package askcmd

import (
	"github.com/spf13/cobra"
)

// NewCommand creates and returns the ask subcommand.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ask [question]",
		Short: "Ask a question about the repository",
		Long:  `Retrieve repository memory and explain historical decisions based on developer queries.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			question := ""
			if len(args) > 0 {
				question = args[0]
			}
			if question != "" {
				cmd.Printf("Querying repository memory for: %q...\n", question)
			} else {
				cmd.Println("Querying repository memory...")
			}
			return nil
		},
	}
}
