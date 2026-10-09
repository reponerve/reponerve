package botfilter

import "testing"

func TestIsBot(t *testing.T) {
	tests := []struct {
		name  string
		email string
		isBot bool
	}{
		{"dependabot[bot]", "dependabot[bot]@users.noreply.github.com", true},
		{"renovate[bot]", "renovate@whitesourcesoftware.com", true},
		{"github-actions[bot]", "41898282+github-actions[bot]@users.noreply.github.com", true},
		{"snyk-bot", "snyk-bot@snyk.io", true},
		{"Alice Smith", "alice@example.com", false},
		{"Bob Jones", "bob@gmail.com", false},
		{"Bot Maker", "human@corp.com", true}, // contains word "bot"
	}

	for _, tt := range tests {
		got := IsBot(tt.name, tt.email)
		if got != tt.isBot {
			t.Errorf("IsBot(%q, %q) = %v; want %v", tt.name, tt.email, got, tt.isBot)
		}
	}
}
