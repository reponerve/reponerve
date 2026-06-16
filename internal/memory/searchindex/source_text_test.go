package searchindex

import (
	"strings"
	"testing"
)

func TestSourceDocumentText(t *testing.T) {
	meta := `{"content":"# Redis\n\nWe adopted Redis for caching to reduce database load.\n","status":"Accepted"}`
	text := sourceDocumentText(meta, "Use Redis Cache")
	if text == "" {
		t.Fatal("expected text")
	}
	for _, part := range []string{"Redis", "caching", "Accepted"} {
		if !strings.Contains(text, part) {
			t.Fatalf("expected %q in text %q", part, text)
		}
	}
}
