package server

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/reponerve/reponerve/internal/agent/development"
)

func parseDevelopmentOutputOptions(getArg func(string, bool) (string, error)) development.OutputOptions {
	opts := development.OutputOptions{Format: development.OutputFormatProse}
	format, err := getArg("format", false)
	if err == nil && strings.TrimSpace(format) != "" {
		opts.Format = development.NormalizeOutputFormat(format)
	}
	budgetStr, err := getArg("token_budget", false)
	if err != nil || strings.TrimSpace(budgetStr) == "" {
		budgetStr, _ = getArg("token-budget", false)
	}
	if budget, err := strconv.Atoi(strings.TrimSpace(budgetStr)); err == nil && budget > 0 {
		opts.TokenBudget = budget
	}
	return opts.WithDefaultBudget()
}

func addDevelopmentOutputSchema(schema *InputSchema) {
	schema.Properties["format"] = map[string]interface{}{
		"type":        "string",
		"description": "Formatted output style: prose, json, or compact (default: prose)",
	}
	schema.Properties["token_budget"] = map[string]interface{}{
		"type":        "integer",
		"description": fmt.Sprintf("Approximate max tokens for formatted output (default %d)", development.DefaultTokenBudget),
	}
}
