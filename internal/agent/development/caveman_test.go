package development

import (
	"strings"
	"testing"
)

func TestToCavemanShortensHeaders(t *testing.T) {
	in := "ENTITY BRIEFINGS\n  foo [bar]\n\nREPOSITORY CONTEXT\n  decision"
	out := ToCaveman(in)
	if strings.Contains(out, "ENTITY BRIEFINGS") {
		t.Fatalf("expected header shortened: %q", out)
	}
	if !strings.Contains(out, "BRIEF") || !strings.Contains(out, "REPO") {
		t.Fatalf("expected caveman headers: %q", out)
	}
}

func TestToCavemanReducesSize(t *testing.T) {
	in := strings.Repeat("ENTITY BRIEFINGS\n  internal/foo/bar.go [FILE]\n  Related decisions: ADR-1\n\n", 20)
	proseLen := len(in)
	caveLen := len(ToCaveman(in))
	if caveLen >= proseLen {
		t.Fatalf("caveman should shrink prose: prose=%d cave=%d", proseLen, caveLen)
	}
	if float64(caveLen)/float64(proseLen) > 0.70 {
		t.Fatalf("expected meaningful reduction, got %d/%d (%.0f%%)", caveLen, proseLen, 100*float64(caveLen)/float64(proseLen))
	}
}

func TestTruncateToTokenBudget(t *testing.T) {
	text := strings.Repeat("word ", 100)
	out := TruncateToTokenBudget(text, 10)
	if len(out) > 10*4+3 {
		t.Fatalf("truncate exceeded budget: len=%d", len(out))
	}
}
