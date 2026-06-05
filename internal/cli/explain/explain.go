package explaincmd

import (
	"github.com/spf13/cobra"
)

// NewCommand creates and returns the explain subcommand.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "explain [component]",
		Short: "Explain a repository component",
		Long:  `Provide understanding of a component, including purpose, history, dependencies, and ownership.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			component := ""
			if len(args) > 0 {
				component = args[0]
			}
			if component != "" {
				cmd.Printf("Explaining component %q...\n", component)
			} else {
				cmd.Println("Explaining component...")
			}
			return nil
		},
	}
}
