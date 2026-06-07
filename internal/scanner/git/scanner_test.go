package git

import (
	"testing"
	"time"
)

func TestParseGitLog(t *testing.T) {
	mockOutput := "a1b2c3d4e5f6\nJohn Doe <john@example.com>\n2026-06-05T21:15:26+05:30\nfeat(foundations): initial schema implementation\n- add sqlite DB\n- add migrations\x00f6e5d4c3b2a1\nJane Smith <jane@example.com>\n2026-06-05T21:20:00+05:30\nfix: issue with DB connection\x00"

	scanner := NewScanner(nil)
	commits, err := scanner.ParseGitLog("test-repo", mockOutput)
	if err != nil {
		t.Fatalf("failed to parse mock git log output: %v", err)
	}

	if len(commits) != 2 {
		t.Fatalf("expected 2 parsed commits, got %d", len(commits))
	}

	c1 := commits[0]
	if c1.Reference != "a1b2c3d4e5f6" {
		t.Errorf("expected Reference to be 'a1b2c3d4e5f6', got %q", c1.Reference)
	}
	if c1.Author != "John Doe <john@example.com>" {
		t.Errorf("expected author to be 'John Doe <john@example.com>', got %q", c1.Author)
	}
	expectedTime, _ := time.Parse(time.RFC3339, "2026-06-05T21:15:26+05:30")
	if !c1.Timestamp.Equal(expectedTime) {
		t.Errorf("expected timestamp %v, got %v", expectedTime, c1.Timestamp)
	}
	expectedMessage := "feat(foundations): initial schema implementation\n- add sqlite DB\n- add migrations"
	if c1.Title != expectedMessage {
		t.Errorf("expected title to be %q, got %q", expectedMessage, c1.Title)
	}

	c2 := commits[1]
	if c2.Reference != "f6e5d4c3b2a1" {
		t.Errorf("expected Reference to be 'f6e5d4c3b2a1', got %q", c2.Reference)
	}
	if c2.Author != "Jane Smith <jane@example.com>" {
		t.Errorf("expected author to be 'Jane Smith <jane@example.com>', got %q", c2.Author)
	}
	if c2.Title != "fix: issue with DB connection" {
		t.Errorf("expected title to be 'fix: issue with DB connection', got %q", c2.Title)
	}
}
