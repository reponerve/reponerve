package integration

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Result summarizes files written or updated during IDE integration.
type Result struct {
	Installed []string
	Updated   []string
	Skipped   []string
}

// Options controls IDE integration installation.
type Options struct {
	ProjectRoot string
	Force       bool
	GlobalSkill bool
}

// Install configures MCP, Cursor skill, and related IDE files for the project.
func Install(opts Options) (Result, error) {
	if opts.ProjectRoot == "" {
		wd, err := os.Getwd()
		if err != nil {
			return Result{}, fmt.Errorf("resolve project root: %w", err)
		}
		opts.ProjectRoot = wd
	}

	var result Result

	projectFiles := []struct {
		bundleName string
		relPath    string
		merge      func(string, []byte) ([]byte, error)
	}{
		{"cursor-mcp.json", ".cursor/mcp.json", mergeCursorMCP},
		{"vscode-mcp.json", ".vscode/mcp.json", mergeVSCodeMCP},
		{"continue-mcp.json", ".continue/mcpServers/reponerve.json", nil},
		{"skill-SKILL.md", ".cursor/skills/reponerve/SKILL.md", nil},
		{"skill-reference.md", ".cursor/skills/reponerve/reference.md", nil},
		{"rule-reponerve.mdc", ".cursor/rules/reponerve.mdc", nil},
	}

	for _, file := range projectFiles {
		content, err := fs.ReadFile(bundleFS, filepath.Join("bundle", file.bundleName))
		if err != nil {
			return result, fmt.Errorf("read bundle %s: %w", file.bundleName, err)
		}

		target := filepath.Join(opts.ProjectRoot, file.relPath)
		action, err := writeProjectFile(target, content, opts.Force, file.merge)
		if err != nil {
			return result, err
		}
		switch action {
		case actionInstalled:
			result.Installed = append(result.Installed, file.relPath)
		case actionUpdated:
			result.Updated = append(result.Updated, file.relPath)
		case actionSkipped:
			result.Skipped = append(result.Skipped, file.relPath)
		}
	}

	if opts.GlobalSkill {
		globalResult, err := installGlobalSkill(opts.Force)
		if err != nil {
			return result, err
		}
		result.Installed = append(result.Installed, globalResult.Installed...)
		result.Updated = append(result.Updated, globalResult.Updated...)
		result.Skipped = append(result.Skipped, globalResult.Skipped...)
	}

	return result, nil
}

type writeAction int

const (
	actionSkipped writeAction = iota
	actionInstalled
	actionUpdated
)

func writeProjectFile(target string, content []byte, force bool, merge func(string, []byte) ([]byte, error)) (writeAction, error) {
	if merge != nil {
		if _, err := os.Stat(target); err == nil {
			merged, err := merge(target, content)
			if err != nil {
				return actionSkipped, err
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return actionSkipped, fmt.Errorf("create directory for %s: %w", target, err)
			}
			if err := os.WriteFile(target, merged, 0o644); err != nil {
				return actionSkipped, fmt.Errorf("write %s: %w", target, err)
			}
			return actionUpdated, nil
		}
	}

	if _, err := os.Stat(target); err == nil && !force {
		return actionSkipped, nil
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return actionSkipped, fmt.Errorf("create directory for %s: %w", target, err)
	}
	if err := os.WriteFile(target, content, 0o644); err != nil {
		return actionSkipped, fmt.Errorf("write %s: %w", target, err)
	}

	if force {
		return actionUpdated, nil
	}
	return actionInstalled, nil
}

func mergeCursorMCP(path string, bundle []byte) ([]byte, error) {
	return mergeMCPConfig(path, bundle, "mcpServers")
}

func mergeVSCodeMCP(path string, bundle []byte) ([]byte, error) {
	return mergeMCPConfig(path, bundle, "servers")
}

func mergeMCPConfig(path string, bundle []byte, topKey string) ([]byte, error) {
	existing := map[string]json.RawMessage{}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &existing); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	incoming := map[string]json.RawMessage{}
	if err := json.Unmarshal(bundle, &incoming); err != nil {
		return nil, fmt.Errorf("parse bundle for %s: %w", topKey, err)
	}

	servers := map[string]any{}
	if raw, ok := existing[topKey]; ok {
		if err := json.Unmarshal(raw, &servers); err != nil {
			return nil, fmt.Errorf("parse %s in %s: %w", topKey, path, err)
		}
	}

	incomingServers := map[string]any{}
	if raw, ok := incoming[topKey]; ok {
		if err := json.Unmarshal(raw, &incomingServers); err != nil {
			return nil, fmt.Errorf("parse incoming %s: %w", topKey, err)
		}
	}
	for name, cfg := range incomingServers {
		servers[name] = cfg
	}

	encodedServers, err := json.Marshal(servers)
	if err != nil {
		return nil, err
	}
	existing[topKey] = encodedServers

	out, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

func installGlobalSkill(force bool) (Result, error) {
	var result Result

	home, err := os.UserHomeDir()
	if err != nil {
		return result, fmt.Errorf("resolve home directory: %w", err)
	}

	globalFiles := []struct {
		bundleName string
		relPath    string
	}{
		{"skill-SKILL.md", "SKILL.md"},
		{"skill-reference.md", "reference.md"},
	}

	for _, file := range globalFiles {
		content, err := fs.ReadFile(bundleFS, filepath.Join("bundle", file.bundleName))
		if err != nil {
			return result, fmt.Errorf("read bundle %s: %w", file.bundleName, err)
		}

		target := filepath.Join(home, ".cursor", "skills", "reponerve", file.relPath)
		display := filepath.ToSlash(filepath.Join("~/.cursor/skills/reponerve", file.relPath))
		action, err := writeProjectFile(target, content, force, nil)
		if err != nil {
			return result, err
		}
		switch action {
		case actionInstalled:
			result.Installed = append(result.Installed, display)
		case actionUpdated:
			result.Updated = append(result.Updated, display)
		case actionSkipped:
			result.Skipped = append(result.Skipped, display)
		}
	}

	return result, nil
}

// FormatSummary returns human-readable install lines for CLI output.
func FormatSummary(result Result) []string {
	lines := make([]string, 0, len(result.Installed)+len(result.Updated)+1)
	if len(result.Installed)+len(result.Updated) > 0 {
		lines = append(lines, "✓ IDE integration installed (Cursor skill + MCP for Cursor, VS Code, Continue)")
	}
	for _, path := range result.Installed {
		lines = append(lines, fmt.Sprintf("  + %s", path))
	}
	for _, path := range result.Updated {
		lines = append(lines, fmt.Sprintf("  ~ %s", path))
	}
	if len(result.Installed)+len(result.Updated) > 0 {
		lines = append(lines, "  → Restart MCP in your IDE, then use Agent chat with RepoNerve")
	}
	return lines
}
