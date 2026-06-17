package development

import (
	"strings"
	"unicode"
)

var cavemanHeaderReplacements = map[string]string{
	"ENTITY BRIEFINGS":     "BRIEF",
	"CODE CONTEXT":         "CODE",
	"REPOSITORY CONTEXT":   "REPO",
	"Evidence:":            "EV:",
	"Source Services:":     "SRC:",
	"Suggested Workflow:":  "FLOW:",
	"Suggested Steps:":     "STEPS:",
	"Related Entities:":    "REL:",
	"Answer Type:":         "TYPE:",
	"Question:":            "Q:",
	"Topic:":               "TOPIC:",
	"Task:":                "TASK:",
	"Summary:":             "SUM:",
}

// ToCaveman compresses prose DE output for token-efficient agent consumption.
func ToCaveman(text string) string {
	if text == "" {
		return ""
	}
	out := text
	for from, to := range cavemanHeaderReplacements {
		out = strings.ReplaceAll(out, from, to)
	}

	lines := strings.Split(out, "\n")
	compact := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		trimmed = collapseSpaces(trimmed)
		if strings.HasPrefix(trimmed, "- ") {
			trimmed = "· " + strings.TrimPrefix(trimmed, "- ")
		}
		trimmed = strings.ReplaceAll(trimmed, "[FILE]", "[F]")
		trimmed = strings.ReplaceAll(trimmed, "[FUNCTION]", "[fn]")
		trimmed = strings.ReplaceAll(trimmed, "[METHOD]", "[m]")
		trimmed = strings.ReplaceAll(trimmed, "[STRUCT]", "[S]")
		trimmed = strings.ReplaceAll(trimmed, "[DECISION]", "[D]")
		trimmed = strings.ReplaceAll(trimmed, "internal/", "i/")
		compact = append(compact, trimmed)
	}
	return strings.Join(compact, "\n")
}

// TruncateToTokenBudget limits text to an approximate token budget (4 chars ≈ 1 token).
func TruncateToTokenBudget(text string, budget int) string {
	if budget <= 0 || text == "" {
		return text
	}
	maxChars := budget * 4
	if len(text) <= maxChars {
		return text
	}
	if maxChars <= 3 {
		return text[:maxChars]
	}
	return text[:maxChars-3] + "..."
}

func collapseSpaces(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
			continue
		}
		b.WriteRune(r)
		prevSpace = false
	}
	return strings.TrimSpace(b.String())
}
