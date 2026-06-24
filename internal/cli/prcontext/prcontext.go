package prcontextcmd

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/cli/devwire"
)

// NewCommand creates the pr-context subcommand.
func NewCommand() *cobra.Command {
	var files []string
	var topic string

	cmd := &cobra.Command{
		Use:   "pr-context [topic]",
		Short: "Assemble PR evidence from changed files (review + ship readiness)",
		Long:  `Team Delivery Intelligence — structured PR context with review, ship_check, and pr_comment_markdown for CI.`,
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			changed := append([]string{}, files...)
			for _, arg := range args {
				if topic == "" && len(changed) == 0 && !looksLikePath(arg) {
					topic = arg
					continue
				}
				changed = append(changed, arg)
			}
			if len(changed) == 0 {
				return cmd.Help()
			}

			return devwire.RunPRContext(cmd, topic, changed, func(ctx context.Context, session *devwire.Handle, topic string, files []string) (*development.PRContextResult, error) {
				return session.Service.PreparePRContext(ctx, development.PRContextRequest{
					RepositoryID: session.RepositoryID,
					Topic:        topic,
					ChangedFiles: files,
				})
			})
		},
	}
	cmd.Flags().StringArrayVarP(&files, "file", "f", nil, "Changed file path (repeatable)")
	return devwire.BindDECmd(cmd)
}

func looksLikePath(s string) bool {
	return strings.Contains(s, "/") || strings.HasSuffix(s, ".go") || strings.HasSuffix(s, ".md")
}
