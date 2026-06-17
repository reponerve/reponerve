package devwire

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
)

const (
	FormatProse   = "prose"
	FormatJSON    = "json"
	FormatCaveman = "caveman"
)

// BindDECmd registers --format, --json, and --token-budget on a DE command.
func BindDECmd(cmd *cobra.Command) *cobra.Command {
	BindOutputFlags(cmd)
	return cmd
}

// BindOutputFlags registers output format flags.
func BindOutputFlags(cmd *cobra.Command) {
	cmd.Flags().String("format", FormatProse, "Output format: prose, json, caveman")
	cmd.Flags().Bool("json", false, "Emit MCP-compatible JSON (same as --format json)")
	cmd.Flags().Int("token-budget", 0, "Approximate max tokens for prose/caveman output (0 = unlimited)")
}

// ResolveFormat returns the effective output format from flags.
func ResolveFormat(cmd *cobra.Command) (string, error) {
	useJSON, err := cmd.Flags().GetBool("json")
	if err != nil {
		return "", err
	}
	if useJSON {
		return FormatJSON, nil
	}
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return "", err
	}
	format = strings.ToLower(strings.TrimSpace(format))
	switch format {
	case FormatProse, FormatJSON, FormatCaveman:
		return format, nil
	default:
		return "", fmt.Errorf("unsupported format %q (use prose, json, or caveman)", format)
	}
}

// WriteDEResult prints or encodes a Development Experience result.
func WriteDEResult(cmd *cobra.Command, formatted string, structured any) error {
	format, err := ResolveFormat(cmd)
	if err != nil {
		return err
	}

	budget, err := cmd.Flags().GetInt("token-budget")
	if err != nil {
		return err
	}

	display := formatted
	switch format {
	case FormatCaveman:
		display = development.ToCaveman(formatted)
		display = development.TruncateToTokenBudget(display, budget)
	case FormatProse:
		display = development.TruncateToTokenBudget(formatted, budget)
	case FormatJSON:
		// formatted in JSON envelope stays prose unless caveman requested via format only
	}

	if format == FormatJSON {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(development.NewMCPResult(formatted, structured))
	}

	cmd.Print(display)
	if !strings.HasSuffix(display, "\n") {
		cmd.Print("\n")
	}
	return nil
}
