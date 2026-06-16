package development

import (
	"strings"
	"testing"
)

func TestAdrRationaleSnippet(t *testing.T) {
	meta := `{"content":"# Use Redis\n\n## Context\nWe need a fast cache for sessions.\n\n## Decision\nUse Redis.\n"}`
	snippet := adrRationaleSnippet(meta)
	if snippet == "" {
		t.Fatal("expected snippet")
	}
	if !strings.Contains(snippet, "fast cache") {
		t.Fatalf("expected context body, got %q", snippet)
	}
}

func TestNormalizeTaskTopic(t *testing.T) {
	got := NormalizeTaskTopic("PROJ-99: Add audit logging")
	if got != "Add audit logging" {
		t.Fatalf("got %q", got)
	}
}

func TestLooksLikeTaskDescription(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"Add OAuth login", true},
		{"PROJ-482: Support SAML SSO", true},
		{"As a user I want login\nSo that I can access the app", true},
		{"What is Redis?", false},
	}
	for _, tc := range cases {
		if got := LooksLikeTaskDescription(tc.in); got != tc.want {
			t.Fatalf("LooksLikeTaskDescription(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
