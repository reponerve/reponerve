package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const chatHookMarker = "## RepoNerve"

const chatHookSection = `## RepoNerve

RepoNerve provides evidence-backed repository context in **AI chat without MCP**.

Before answering questions about this codebase or making edits:

1. Ensure memory exists:
   ` + "`test -f .reponerve/memory.db || (reponerve init && reponerve scan)`" + `
2. Load context (pick one):
   - ` + "`reponerve ask \"<question>\" --json`" + `
   - ` + "`reponerve plan \"<task>\" --json`" + ` (pasted tickets)
   - ` + "`reponerve onboard --json`" + ` (day one)
   - ` + "`reponerve explain-function \"<name>\" --package <pkg> --json`" + ` (verify a fix / one symbol)
   - ` + "`reponerve explain-file \"<path>\" --json`" + ` (verify a file)
3. Read the JSON envelope in order: ` + "`structured`" + ` → ` + "`agent`" + ` → ` + "`formatted`" + `
4. Answer and edit **only** from RepoNerve evidence. Do not grep the repo first.

For verification ("is this fix correct?"), prefer ` + "`explain-function`" + ` / ` + "`explain-struct`" + ` / ` + "`explain-file`" + ` over broad ` + "`ask`" + ` or full ` + "`plan`" + ` JSON.

Chat triggers: ` + "`/reponerve ask \"...\"`" + ` | Skill: ` + "`.cursor/skills/reponerve/SKILL.md`" + `
`

func installChatHooks(projectRoot string, force bool) (string, error) {
	for _, name := range []string{"CLAUDE.md", "AGENTS.md"} {
		path := filepath.Join(projectRoot, name)
		action, err := mergeMarkdownSection(path, chatHookMarker, chatHookSection, force)
		if err != nil {
			return "", err
		}
		if action == actionInstalled || action == actionUpdated {
			return filepath.Join(".", name), nil
		}
	}
	return "", nil
}

func mergeMarkdownSection(path, marker, section string, force bool) (writeAction, error) {
	section = strings.TrimRight(section, "\n") + "\n"
	content := []byte(section)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if name := filepath.Base(path); name == "AGENTS.md" {
			return actionSkipped, nil
		}
		return writeProjectFile(path, content, force, nil)
	} else if err != nil {
		return actionSkipped, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return actionSkipped, err
	}
	text := string(data)
	if strings.Contains(text, marker) {
		if !force {
			return actionSkipped, nil
		}
		start := strings.Index(text, marker)
		rest := text[start+len(marker):]
		end := len(text)
		if idx := strings.Index(rest, "\n## "); idx >= 0 {
			end = start + len(marker) + idx
		}
		text = text[:start] + section + strings.TrimLeft(text[end:], "\n")
		if !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			return actionSkipped, fmt.Errorf("write %s: %w", path, err)
		}
		return actionUpdated, nil
	}

	merged := strings.TrimRight(text, "\n") + "\n\n" + section
	if err := os.WriteFile(path, []byte(merged), 0o644); err != nil {
		return actionSkipped, fmt.Errorf("write %s: %w", path, err)
	}
	return actionInstalled, nil
}
