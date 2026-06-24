package development

import "testing"

func TestDeriveTopicFromChangedFiles(t *testing.T) {
	topic := deriveTopicFromChangedFiles([]string{
		"internal/agent/development/review.go",
		"internal/agent/development/pr_context.go",
		"internal/cli/prcontext/prcontext.go",
	})
	if topic != "agent" {
		t.Fatalf("expected topic agent, got %q", topic)
	}
}

func TestNormalizeChangedFiles_DedupesAndSorts(t *testing.T) {
	got := normalizeChangedFiles([]string{"b.go", "a.go", "b.go", "  ", ""})
	if len(got) != 2 || got[0] != "a.go" || got[1] != "b.go" {
		t.Fatalf("unexpected normalized files: %#v", got)
	}
}

func TestFormatPRCommentMarkdown_IncludesTopic(t *testing.T) {
	md := FormatPRCommentMarkdown(&PRContextResult{
		Topic:        "agent",
		ChangedFiles: []string{"internal/agent/foo.go"},
		Review: &DevelopmentReviewGuide{
			DisciplineChecks: []DisciplineCheck{{
				Category: "adr",
				Severity: "required",
				Message:  "Record ADR",
			}},
		},
	})
	if !contains(md, "RepoNerve PR Context") || !contains(md, "agent") {
		t.Fatalf("unexpected markdown: %s", md)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
