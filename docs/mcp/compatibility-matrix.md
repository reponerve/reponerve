# RepoNerve MCP Compatibility Matrix

RepoNerve exposes **38 MCP tools** over **STDIO** (`reponerve mcp`). Any MCP-capable client can use RepoNerve in **AI chat** with **any underlying LLM** the client provides.

**Transport:** STDIO JSON-RPC  
**Launch:** `reponerve mcp`  
**Workspace:** set `REPONERVE_WORKSPACE` to `<project>/.reponerve` when the client supports `${workspaceFolder}`

---

## Summary

| Client | IDE / platform | MCP client | Direct chat | Config |
| --- | --- | --- | --- | --- |
| **Cursor** | Cursor | Yes | Agent chat + MCP tools | `.cursor/mcp.json` |
| **GitHub Copilot** | VS Code | Yes | Copilot Chat (Agent) | `.vscode/mcp.json` |
| **JetBrains AI Assistant** | IDEA, GoLand, WebStorm, PyCharm, … | Yes | AI Assistant chat | Settings → MCP |
| **Windsurf** | Windsurf | Yes | Cascade / chat | `~/.codeium/windsurf/mcp.json` |
| **Continue** | VS Code, JetBrains | Yes | Continue Agent | `.continue/mcpServers/` |
| **Cline** | VS Code | Yes | Cline task/chat | `cline_mcp_settings.json` |
| **Roo Code** | VS Code | Yes | Roo chat | `roo_mcp_settings.json` |
| **Claude Desktop** | Desktop app | Yes | Claude chat | `claude_desktop_config.json` |
| **Claude Code** | Terminal / IDE | Yes | Claude Code sessions | Project MCP config |
| **Zed** | Zed | Varies by version | Agent panel | MCP settings (if enabled) |
| **Codex / custom agents** | Any | Yes | Agent loop | Custom `mcp.json` |

### LLM backends (model-agnostic)

RepoNerve does not depend on a specific model. These LLMs are commonly used **through** the clients above:

| Provider | Models (examples) | Notes |
| --- | --- | --- |
| OpenAI | GPT-4o, GPT-4.1, o-series | Copilot, Cursor, Continue |
| Anthropic | Claude Sonnet, Opus, Haiku | Cursor, Claude apps, Cline |
| Google | Gemini Pro, Flash | Cursor, Continue, some Copilot |
| Meta | Llama (via Ollama, etc.) | Continue, local MCP hosts |
| Mistral, DeepSeek, Qwen, … | Various | Continue, custom agents |

### Without MCP (any LLM)

| Method | Use when |
| --- | --- |
| `reponerve ask` / `plan` / `onboard` in terminal | MCP unavailable; paste output into chat |
| `reponerve context export` | Web LLMs (ChatGPT, Gemini web, Claude.ai) |
| MCP `export_context` | Agent can pull markdown into context |

---

## Tool surface (38)

All MCP clients receive the full registry from `internal/mcp/registry.go`:

| Category | Count | Examples |
| --- | ---: | --- |
| Repository memory | 14 | `list_decisions`, `trace_decision`, `generate_context` |
| Ownership & intelligence | 8 | `recommend_reviewers`, `discover_knowledge` |
| Knowledge graph | 5 | `trace_graph`, `analyze_impact` |
| Development Experience | 11 | `ask`, `explain`, `plan`, `review`, `onboard`, `analyze_topic_impact` |

CLI equivalents: `docs/mcp/configuration-examples.md` and `.cursor/skills/reponerve/reference.md`.

---

## Client notes

### Cursor

- Project: `.cursor/mcp.json` (included in this repo)
- Agent Skill: `.cursor/skills/reponerve/` for workflow when MCP is off
- Settings → Tools & MCP → confirm **reponerve** connected

### VS Code + GitHub Copilot

- Requires VS Code 1.99+ and Copilot extension
- Project: `.vscode/mcp.json` (included in this repo)
- Open `.vscode/mcp.json` → Start server → Copilot Chat → **Agent** mode
- `${workspaceFolder}` must be in **workspace** config, not user-global only

### JetBrains AI Assistant

- IDEA 2025.1+ with AI Assistant plugin
- Settings → Tools → AI Assistant → Model Context Protocol (MCP) → Add
- Paste JSON with `mcpServers.reponerve` block; set working directory to project root
- Can import from Claude Desktop config

### Windsurf

- Config: `~/.codeium/windsurf/mcp.json`
- Same `reponerve` / `mcp` command shape

### Continue

- Place JSON in `.continue/mcpServers/` (this repo includes `reponerve.json`)
- MCP active in **Agent** mode only
- Supports Claude-style `mcpServers` JSON natively

### Claude Desktop / Claude Code

- Global config without `${workspaceFolder}`: run MCP with `cwd` set to project root, or use absolute `REPONERVE_WORKSPACE`

### Limitations (all clients)

- `reponerve` binary must be on `PATH` (or use absolute `command` path)
- SQLite database at `.reponerve/memory.db` must be readable
- MCP stdout must be JSON-RPC only (RepoNerve logs errors to stderr)
- Run `reponerve scan` after repo changes for fresh memory

---

## Verification

```bash
which reponerve
reponerve init && reponerve scan
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | REPONERVE_WORKSPACE="$(pwd)/.reponerve" reponerve mcp
```

Expect JSON listing **38** tools.

---

## Further reading

- AI chat integration (start here): `docs/ai-chat-integration.md`
- Configuration templates: `docs/mcp/configuration-examples.md`
- Troubleshooting: `docs/mcp/troubleshooting.md`
