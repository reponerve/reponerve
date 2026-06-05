package scancmd

import (
	"github.com/spf13/cobra"
)

// NewCommand creates and returns the scan subcommand.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan the repository to build memory",
		Long:  `Scan repository artifacts (code, git history, PRs, ADRs) to build and update repository memory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Scanning repository...")
			return nil
		},
	}
}
