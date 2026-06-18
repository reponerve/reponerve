package development

import (
	"strings"

	"github.com/spf13/cobra"
)

const (
	OutputFormatProse   = "prose"
	OutputFormatJSON    = "json"
	OutputFormatCompact = "compact"
)

// OutputOptions controls formatted DE output for CLI and MCP.
type OutputOptions struct {
	Format      string
	TokenBudget int
}

// NormalizeOutputFormat returns a supported format name.
func NormalizeOutputFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case OutputFormatCompact:
		return OutputFormatCompact
	case OutputFormatJSON:
		return OutputFormatJSON
	default:
		return OutputFormatProse
	}
}

// ApplyOutputFormat renders prose with optional compact compression and token budget.
func ApplyOutputFormat(formatted string, opts OutputOptions) string {
	out := formatted
	switch NormalizeOutputFormat(opts.Format) {
	case OutputFormatCompact:
		out = ToCompact(out)
	}
	return TruncateToTokenBudget(out, opts.TokenBudget)
}

// NewMCPResultWithFormat builds the MCP envelope with formatted output options applied.
func NewMCPResultWithFormat(formatted string, structured any, opts OutputOptions) MCPResult {
	display := ApplyOutputFormat(formatted, opts)
	return MCPResult{
		Formatted:  display,
		Structured: structured,
		Agent:      BuildAgentContextMeta(structured),
	}
}

// OutputOptionsFromFlags reads DE output options from command flags.
func OutputOptionsFromFlags(cmd *cobra.Command) (OutputOptions, error) {
	format, err := ResolveFormat(cmd)
	if err != nil {
		return OutputOptions{}, err
	}
	budget, err := cmd.Flags().GetInt("token-budget")
	if err != nil {
		return OutputOptions{}, err
	}
	return OutputOptions{Format: format, TokenBudget: budget}, nil
}

// ResolveFormat returns the effective output format from CLI flags.
func ResolveFormat(cmd *cobra.Command) (string, error) {
	useJSON, err := cmd.Flags().GetBool("json")
	if err != nil {
		return "", err
	}
	if useJSON {
		return OutputFormatJSON, nil
	}
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return "", err
	}
	format = strings.ToLower(strings.TrimSpace(format))
	switch format {
	case OutputFormatProse, OutputFormatJSON, OutputFormatCompact:
		return format, nil
	default:
		return "", &formatError{format: format}
	}
}

type formatError struct{ format string }

func (e *formatError) Error() string {
	return "unsupported format " + e.format + " (use prose, json, or compact)"
}
