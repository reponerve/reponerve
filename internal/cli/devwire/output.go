package devwire

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
)

const (
	FormatProse   = development.OutputFormatProse
	FormatJSON    = development.OutputFormatJSON
	FormatCompact = development.OutputFormatCompact
)

// BindDECmd registers --format, --json, and --token-budget on a DE command.
func BindDECmd(cmd *cobra.Command) *cobra.Command {
	BindOutputFlags(cmd)
	return cmd
}

// BindOutputFlags registers output format flags.
func BindOutputFlags(cmd *cobra.Command) {
	cmd.Flags().String("format", FormatProse, "Output format: prose, json, compact")
	cmd.Flags().Bool("json", false, "Emit MCP-compatible JSON (same as --format json)")
	cmd.Flags().Int("token-budget", 0, fmt.Sprintf("Approximate max tokens for prose/compact output (0 = default %d)", development.DefaultTokenBudget))
}

// ResolveFormat returns the effective output format from flags.
func ResolveFormat(cmd *cobra.Command) (string, error) {
	return development.ResolveFormat(cmd)
}

// WriteDEResult prints or encodes a Development Experience result.
func WriteDEResult(cmd *cobra.Command, formatted string, structured any) error {
	opts, err := development.OutputOptionsFromFlags(cmd)
	if err != nil {
		return err
	}

	if opts.Format == FormatJSON {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(development.NewMCPResult(formatted, structured))
	}

	display := development.ApplyOutputFormat(formatted, opts)
	cmd.Print(display)
	if !strings.HasSuffix(display, "\n") {
		cmd.Print("\n")
	}
	return nil
}

// FormatErrorf wraps unsupported format errors for CLI callers.
func FormatErrorf(format string) error {
	return fmt.Errorf("unsupported format %q (use prose, json, or compact)", format)
}
