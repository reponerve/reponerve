package botfilter

import (
	"regexp"
	"strings"
)

var botPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\[bot\]`),
	regexp.MustCompile(`(?i)\bbot\b`),
	regexp.MustCompile(`(?i)github-actions`),
	regexp.MustCompile(`(?i)dependabot`),
	regexp.MustCompile(`(?i)renovate`),
	regexp.MustCompile(`(?i)greenkeeper`),
	regexp.MustCompile(`(?i)snyk-bot`),
}

// IsBot checks if a contributor name or email indicates an automated bot.
func IsBot(name, email string) bool {
	combined := strings.ToLower(name + " " + email)
	for _, p := range botPatterns {
		if p.MatchString(combined) {
			return true
		}
	}
	if strings.HasSuffix(strings.ToLower(email), "-bot@users.noreply.github.com") {
		return true
	}
	return false
}
