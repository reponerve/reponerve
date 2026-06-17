package development

import (
	"strings"
	"testing"
)

func TestApplyOutputFormat_CavemanAndBudget(t *testing.T) {
	in := strings.Repeat("Question: Why SQLite?\n\nRelated Entities:\n  - internal/storage/sqlite/sqlite.go [FILE]\n", 8)
	out := ApplyOutputFormat(in, OutputOptions{Format: OutputFormatCaveman, TokenBudget: 50})
	if strings.Contains(out, "Question:") {
		t.Fatalf("expected caveman header compression: %q", out[:min(80, len(out))])
	}
	if len(out) > 50*4+3 {
		t.Fatalf("budget not applied: len=%d", len(out))
	}
}

func TestNewMCPResultWithFormat(t *testing.T) {
	result := NewMCPResultWithFormat("Question: test\n", map[string]string{"k": "v"}, OutputOptions{
		Format: OutputFormatCaveman,
	})
	if !strings.HasPrefix(result.Formatted, "Q:") {
		t.Fatalf("expected caveman formatted output, got %q", result.Formatted)
	}
}
