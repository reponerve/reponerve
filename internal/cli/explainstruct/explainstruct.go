package explainstructcmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/cli/devwire"
)

// NewCommand creates the explain-struct subcommand.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "explain-struct [symbol]",
		Short: "Explain a struct symbol",
		Long:  `Resolve a struct through Code Intelligence and attach related repository context.`,
		Args:  cobra.ExactArgs(1),
	}
	pkgFlag := devwire.BindPackageFlag(cmd)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return devwire.RunSymbolExplanation(cmd, args[0], pkgFlag, func(ctx context.Context, session *devwire.Handle, symbol, packagePath string) (*development.DevelopmentExplanation, error) {
			return session.Service.ExplainStruct(ctx, session.RepositoryID, symbol, packagePath)
		})
	}
	return devwire.BindDECmd(cmd)
}
