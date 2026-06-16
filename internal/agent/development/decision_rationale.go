package development

import (
	"encoding/json"
	"strings"
)

func adrRationaleSnippet(metadataJSON string) string {
	metadataJSON = strings.TrimSpace(metadataJSON)
	if metadataJSON == "" {
		return ""
	}
	var meta struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(metadataJSON), &meta); err != nil {
		return ""
	}
	content := strings.TrimSpace(meta.Content)
	if content == "" {
		return ""
	}
	return extractADRContext(content, 280)
}

func extractADRContext(content string, maxLen int) string {
	lines := strings.Split(content, "\n")
	var body []string
	capture := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "## context") || strings.HasPrefix(lower, "## rationale") || strings.HasPrefix(lower, "## decision") {
			capture = true
			continue
		}
		if capture && strings.HasPrefix(trimmed, "## ") {
			break
		}
		if capture && trimmed != "" {
			body = append(body, trimmed)
		}
	}
	if len(body) == 0 {
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			body = append(body, trimmed)
			if len(body) >= 3 {
				break
			}
		}
	}
	text := strings.Join(body, " ")
	text = strings.Join(strings.Fields(text), " ")
	if len(text) > maxLen {
		return text[:maxLen] + "..."
	}
	return text
}
