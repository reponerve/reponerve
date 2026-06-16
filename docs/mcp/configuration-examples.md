# RepoNerve MCP Configuration Examples & Workflows

Templates for connecting RepoNerve to **AI chat** in any MCP-capable IDE. After setup, talk to your assistant in natural language — it calls RepoNerve tools directly.

**Start here:** `docs/ai-chat-integration.md`  
**Compatibility:** `docs/mcp/compatibility-matrix.md`

---

## Standard server block

```json
{
  "command": "reponerve",
  "args": ["mcp"],
  "env": {
    "REPONERVE_WORKSPACE": "${workspaceFolder}/.reponerve"
  }
}
```

Prerequisites per repo: `reponerve init && reponerve scan` (`init` installs these configs automatically).

---

## Configuration by client

### 1. VS Code + GitHub Copilot

**File:** `.vscode/mcp.json` (included in this repository)

```json
{
  "servers": {
    "reponerve": {
      "type": "stdio",
      "command": "reponerve",
      "args": ["mcp"],
      "env": {
        "REPONERVE_WORKSPACE": "${workspaceFolder}/.reponerve"
      }
    }
  }
}
```

Open the file → **Start** → Copilot Chat → **Agent** mode. See `docs/copilot-chat-integration.md`.

### 2. Cursor

**Project:** `.cursor/mcp.json` (included)

```json
{
  "mcpServers": {
    "reponerve": {
      "command": "reponerve",
      "args": ["mcp"],
      "env": {
        "REPONERVE_WORKSPACE": "${workspaceFolder}/.reponerve"
      }
    }
  }
}
```

**Global:** same block in `~/.cursor/mcp.json`.  
**Skill:** `.cursor/skills/reponerve/` — see `docs/cursor-integration.md`.

### 3. JetBrains AI Assistant

Settings → Tools → AI Assistant → Model Context Protocol (MCP) → **Add** → As JSON:

```json
{
  "mcpServers": {
    "reponerve": {
      "command": "reponerve",
      "args": ["mcp"],
      "env": {
        "REPONERVE_WORKSPACE": "/absolute/path/to/project/.reponerve"
      }
    }
  }
}
```

Set **Working directory** to the project root. JetBrains may not resolve `${workspaceFolder}` in all versions — use an absolute path for `REPONERVE_WORKSPACE` when needed. You can **Import from Claude** config as a shortcut.

### 4. Continue

**File:** `.continue/mcpServers/reponerve.json` (included)

```json
{
  "mcpServers": {
    "reponerve": {
      "command": "reponerve",
      "args": ["mcp"],
      "env": {
        "REPONERVE_WORKSPACE": "${workspaceFolder}/.reponerve"
      }
    }
  }
}
```

Enable **Agent** mode in Continue. JSON configs in `.continue/mcpServers/` are auto-discovered.

### 5. Windsurf

**File:** `~/.codeium/windsurf/mcp.json`

```json
{
  "mcpServers": {
    "reponerve": {
      "command": "reponerve",
      "args": ["mcp"],
      "enabled": true,
      "env": {
        "REPONERVE_WORKSPACE": "${workspaceFolder}/.reponerve"
      }
    }
  }
}
```

### 6. Cline

**File:** `~/Library/Application Support/Code/User/globalStorage/saoudrizwan.claude-dev/settings/cline_mcp_settings.json` (macOS; paths vary on Linux/Windows)

```json
{
  "mcpServers": {
    "reponerve": {
      "command": "reponerve",
      "args": ["mcp"],
      "disabled": false,
      "env": {
        "REPONERVE_WORKSPACE": "/absolute/path/to/project/.reponerve"
      }
    }
  }
}
```

### 7. Roo Code

**File:** `~/Library/Application Support/Code/User/globalStorage/roocode.roo-cline/settings/roo_mcp_settings.json`

```json
{
  "mcpServers": {
    "reponerve": {
      "command": "reponerve",
      "args": ["mcp"],
      "disabled": false,
      "env": {
        "REPONERVE_WORKSPACE": "/absolute/path/to/project/.reponerve"
      }
    }
  }
}
```

### 8. Claude Desktop

**File:** `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS)

```json
{
  "mcpServers": {
    "reponerve": {
      "command": "reponerve",
      "args": ["mcp"],
      "env": {
        "REPONERVE_WORKSPACE": "/absolute/path/to/project/.reponerve"
      }
    }
  }
}
```

Claude Desktop has no `${workspaceFolder}` — use absolute paths or one config per project.

### 9. Claude Code

Add the same `mcpServers.reponerve` block to your Claude Code MCP settings (project or user scope).

---

## Direct chat workflows

After MCP is connected, use natural language in the IDE chat panel.

### Onboarding

- **Chat:** "Onboard me to this repository"
- **MCP:** `onboard`
- **CLI:** `reponerve onboard`

### Pasted task

- **Chat:** paste ticket → "Where should I start?"
- **MCP:** `ask` or `plan`
- **CLI:** `reponerve plan "<task>"`

### Architecture question

- **Chat:** "Why do we use SQLite?"
- **MCP:** `ask` with `question`
- **CLI:** `reponerve ask "Why do we use SQLite?"`

### File / symbol explain

- **MCP:** `explain_file`, `explain_function`, …
- **CLI:** `reponerve explain-file`, `explain-function`, …

### Impact before refactor

- **MCP:** `analyze_topic_impact` with `subject`
- **CLI:** `reponerve impact "subject"`

### Export for web LLM (no MCP)

```bash
reponerve context export -o repo-context.md
```

Or MCP `export_context` — paste markdown into ChatGPT, Gemini, or Claude.ai.

---

## JSON-RPC examples (advanced)

### List decisions

```json
{ "name": "list_decisions", "arguments": {} }
```

### Trace decision

```json
{ "name": "trace_decision", "arguments": { "decision_id": "dec_1" } }
```

### Ask

```json
{ "name": "ask", "arguments": { "question": "Why do we use SQLite?" } }
```

### Plan

```json
{ "name": "plan", "arguments": { "task": "Add OAuth login" } }
```

---

## Further reading

- AI chat integration: `docs/ai-chat-integration.md`
- Compatibility matrix: `docs/mcp/compatibility-matrix.md`
- Troubleshooting: `docs/mcp/troubleshooting.md`
