# AI Chat Integration

RepoNerve is designed for **direct conversation in AI chat** — in any major IDE or assistant. You type natural language; the AI calls RepoNerve tools and answers from **repository evidence**, not from blind file search.

RepoNerve does **not** ship its own LLM. It is **model-agnostic**: whatever model your IDE uses (GPT, Claude, Gemini, Llama, Copilot, etc.) receives the same structured context through MCP or CLI.

---

## How direct chat works

```text
You (natural language in chat)
        ↓
IDE AI assistant (any LLM)
        ↓
RepoNerve MCP tools (38)  OR  reponerve CLI (terminal)
        ↓
Evidence-backed answer / plan / review
```

**You do not need to know CLI commands.** Ask normally:

- "Onboard me to this repo"
- Paste a full Jira ticket and ask "Where do I start?"
- "Why do we use SQLite?"
- "Explain `internal/mcp/server/server.go`"
- "What breaks if we change the MCP registry?"

The assistant should invoke `onboard`, `ask`, `plan`, `explain_file`, `analyze_topic_impact`, or related tools automatically.

### Prerequisites (once per repository)

```bash
go install github.com/reponerve/reponerve/cmd/reponerve@latest   # or go install ./cmd/reponerve
reponerve init    # creates workspace + installs Cursor skill + MCP configs automatically
reponerve scan
```

`reponerve init` writes project integration files (`.cursor/`, `.vscode/mcp.json`, `.continue/`) and installs the global Cursor skill to `~/.cursor/skills/reponerve/`. Re-run `reponerve integrate` to refresh, or `reponerve integrate --force` to overwrite skill files.

Optional: keep repository memory fresh after each commit:

```bash
reponerve hook install    # post-commit → reponerve scan
reponerve hook status
```

Works in any git repo (Cursor, Claude Code, VS Code, terminal agents). `reponerve hook uninstall` removes the RepoNerve block without deleting other hook content.

Install the binary: `go install ./cmd/reponerve` (from this repo) or your release artifact.

---

## Integration surfaces

| Surface | Works in | Role |
| --- | --- | --- |
| **MCP** (`reponerve mcp`) | Cursor, VS Code, JetBrains, Windsurf, Continue, Cline, Roo, Claude Desktop, Claude Code, … | Direct tool calls from AI chat |
| **Agent Skill** | Cursor (project + `~/.cursor/skills/`) | Context-first workflow when MCP is off |
| **CLI** | Any IDE terminal | Same semantics; agent runs commands for you |
| **Export / paste** | Web LLMs (ChatGPT, Gemini, Claude.ai) | `reponerve context export` or MCP `export_context` |

### Software Development Council

For multi-perspective review (architecture, security, shipping, product), Cursor loads `.cursor/rules/software-development-council.mdc`. Ask normally — the agent auto-routes to relevant council members. Full spec: `docs/council/software-development-council.md`.

---

## IDE setup (quick reference)

| IDE / client | Config location | Guide |
| --- | --- | --- |
| **All (auto)** | `reponerve init` | Installs `.cursor/`, `.vscode/mcp.json`, `.continue/` + global skill |
| **Cursor** | `.cursor/mcp.json` | `docs/cursor-integration.md` |
| **VS Code + Copilot** | `.vscode/mcp.json` | `docs/copilot-chat-integration.md` |
| **JetBrains** (IDEA, GoLand, WebStorm, …) | Settings → AI Assistant → MCP | `docs/mcp/configuration-examples.md` |
| **Windsurf** | `~/.codeium/windsurf/mcp.json` | `docs/mcp/configuration-examples.md` |
| **Continue** | `.continue/mcpServers/reponerve.json` | `docs/mcp/configuration-examples.md` |
| **Cline** | `cline_mcp_settings.json` | `docs/mcp/configuration-examples.md` |
| **Roo Code** | `roo_mcp_settings.json` | `docs/mcp/configuration-examples.md` |
| **Claude Desktop** | `claude_desktop_config.json` | `docs/mcp/configuration-examples.md` |
| **Claude Code** | Project or global MCP config | `docs/mcp/configuration-examples.md` |

Full compatibility matrix: `docs/mcp/compatibility-matrix.md`

### Standard MCP block

Most clients accept this shape (adjust top-level key: `servers` vs `mcpServers` per client):

```json
{
  "command": "reponerve",
  "args": ["mcp"],
  "env": {
    "REPONERVE_WORKSPACE": "${workspaceFolder}/.reponerve"
  }
}
```

Use **workspace-scoped** config so `${workspaceFolder}` resolves. If your client does not support variables, set `REPONERVE_WORKSPACE` to the absolute path of `<project>/.reponerve`.

---

## LLM compatibility

RepoNerve works with **any LLM** the host application provides:

| LLM family | Typical hosts | RepoNerve path |
| --- | --- | --- |
| OpenAI (GPT-4o, o-series, …) | Copilot, Cursor, Continue | MCP tools |
| Anthropic (Claude) | Cursor, Claude Desktop, Cline | MCP tools |
| Google (Gemini) | Cursor, Continue, web | MCP or paste export |
| Meta (Llama) | Ollama + Continue, local agents | MCP or CLI |
| Mistral, DeepSeek, etc. | Continue, custom agents | MCP or CLI |

RepoNerve never calls external LLM APIs. **No API keys** are required for RepoNerve itself. Token savings come from delivering pre-digested evidence instead of raw repo exploration.

### Chat without MCP (recommended default)

When MCP is off, the agent runs CLI in the terminal. You chat normally; the skill runs RepoNerve for you.

```bash
reponerve ask "Why do we use SQLite?" --json
reponerve ask "Why do we use SQLite?" --format compact --token-budget 1500
reponerve plan "Add OAuth login" --json
reponerve onboard --json
```

MCP equivalent: pass `"format": "compact"` and `"token_budget": 1500` on `ask`, `explain`, `plan`, and other DE tools.

`--json` returns the **same envelope as MCP**: `structured`, `agent`, `formatted`.

- **Cursor:** `/reponerve ask "..."` or ask naturally — skill auto-loads
- **Claude Code:** `CLAUDE.md` RepoNerve section (installed by `reponerve init`)
- **Any agent with bash:** follow `.cursor/skills/reponerve/SKILL.md`

### Web chat (no MCP, no terminal)

For browser-only assistants where the agent cannot run CLI:

```bash
reponerve context export -o /tmp/repo-context.md
cat /tmp/repo-context.md
```

Paste into the chat with: "Answer only from this RepoNerve evidence."

---

## Example chat prompts

| You type | RepoNerve should use |
| --- | --- |
| Paste full ticket | `plan` / `reponerve plan "..." --json` |
| "I'm new here" | `onboard` / `reponerve onboard --json` |
| "What is RepositoryContext?" | `ask` / `reponerve ask "..." --json` |
| "Explain this file: internal/foo/bar.go" | `explain_file` |
| "Is this fix correct?" / verify one symbol | `explain_function`, `explain_struct`, `explain_file` with `--package` |
| "Who owns the storage layer?" | `list_expertise`, `recommend_reviewers` |
| "What ADRs mention authentication?" | `list_decisions`, `ask` |
| "Review my OAuth change" | `review` |
| "Impact of renaming Config struct" | `analyze_topic_impact` |

---

## Agent contract (all hosts)

Whether the host is Cursor, Copilot, or Claude:

1. **Load RepoNerve context before** broad file reads or edits
2. **Cite only** paths, symbols, and decisions from RepoNerve output
3. **Say "no evidence"** when RepoNerve does not have a fact — do not invent
4. Read `structured` → `agent` → `formatted` (MCP or `reponerve ... --json`)

Cursor users: also follow `.cursor/skills/reponerve/SKILL.md`.

---

## Troubleshooting

| Problem | Fix |
| --- | --- |
| Tools not in chat | Use CLI: `reponerve ask "..." --json` (no MCP required) |
| Empty answers | Run `reponerve scan`; check `.reponerve/memory.db` exists |
| Wrong repo | Set `REPONERVE_WORKSPACE` to `<project>/.reponerve` |
| MCP JSON errors on stdout | Rebuild CLI; check Output → MCP for stderr pollution |

See `docs/mcp/troubleshooting.md`.

---

## Further reading

- Universal understanding: `docs/product/universal-understanding.md`
- Agent context contract: `docs/architecture/agent-context-contract.md`
- MCP configuration examples: `docs/mcp/configuration-examples.md`
- MCP compatibility matrix: `docs/mcp/compatibility-matrix.md`
