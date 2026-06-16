package searchindex

import (
	"encoding/json"
	"strings"
)

func sourceDocumentText(metadataJSON, title string) string {
	text := strings.TrimSpace(title)
	if metadataJSON == "" {
		return text
	}
	var meta struct {
		Content string `json:"content"`
		Status  string `json:"status"`
	}
	if err := json.Unmarshal([]byte(metadataJSON), &meta); err != nil {
		return text
	}
	body := plainTextFromMarkdown(meta.Content)
	return joinNonEmpty(text, meta.Status, body)
}

func plainTextFromMarkdown(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}
	var lines []string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		lines = append(lines, trimmed)
	}
	text := strings.Join(lines, " ")
	return strings.Join(strings.Fields(text), " ")
}
