package devwire

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
)

// BindPackageFlag registers --package for symbol explain commands.
func BindPackageFlag(cmd *cobra.Command) *string {
	var packagePath string
	cmd.Flags().StringVar(&packagePath, "package", "", "Go package path to disambiguate short symbol names (e.g. internal/context)")
	return &packagePath
}

// RunSymbolExplanation executes a symbol explain workflow with optional package disambiguation.
func RunSymbolExplanation(
	cmd *cobra.Command,
	symbol string,
	packagePath *string,
	run func(context.Context, *Handle, string, string) (*development.DevelopmentExplanation, error),
) error {
	pkg := ""
	if packagePath != nil {
		pkg = *packagePath
	}
	return RunExplanation(cmd, symbol, func(ctx context.Context, session *Handle, _ string) (*development.DevelopmentExplanation, error) {
		return run(ctx, session, symbol, pkg)
	})
}
