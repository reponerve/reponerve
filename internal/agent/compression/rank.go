package compression

import (
	"strings"
	"unicode"
)

func topicTokens(topic string) []string {
	topic = strings.ToLower(strings.TrimSpace(topic))
	if topic == "" {
		return nil
	}
	var tokens []string
	var b strings.Builder
	flush := func() {
		if b.Len() == 0 {
			return
		}
		tokens = append(tokens, b.String())
		b.Reset()
	}
	for _, r := range topic {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	return tokens
}

func scoreText(tokens []string, parts ...string) int {
	if len(tokens) == 0 {
		return 0
	}
	text := strings.ToLower(strings.Join(parts, " "))
	score := 0
	for _, tok := range tokens {
		if tok == "" {
			continue
		}
		if strings.Contains(text, tok) {
			score += 10
		}
	}
	return score
}

func estimateTokens(text string) int {
	if text == "" {
		return 0
	}
	return (len(text) + 3) / 4
}

func applyGraphBoost(scores map[string]int) {
	if len(scores) == 0 {
		return
	}
	// Propagate half of a node's score to graph neighbors (one hop).
	boosted := make(map[string]int, len(scores))
	for id, score := range scores {
		if score <= 0 {
			continue
		}
		boosted[id] += score / 2
	}
	for id, delta := range boosted {
		if delta > 0 {
			scores[id] += delta
		}
	}
}

func applyRelationshipBoost(scores map[string]int, edges [][2]string) {
	if len(scores) == 0 || len(edges) == 0 {
		return
	}
	for _, edge := range edges {
		from, to := edge[0], edge[1]
		if scores[from] > 0 {
			scores[to] += scores[from] / 2
		}
		if scores[to] > 0 {
			scores[from] += scores[to] / 2
		}
	}
}
