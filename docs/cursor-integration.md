# Cursor Integration

RepoNerve gives AI **proper repository context** in **direct chat** — symbols, relationships, decisions, and scoped plans. Cursor is one of several supported hosts; the same MCP server works in VS Code, JetBrains, Windsurf, Continue, Claude, and more.

**Universal guide:** `docs/ai-chat-integration.md`  
**All clients:** `docs/mcp/compatibility-matrix.md`

Cursor connects through **MCP** (tools) and a project **skill** (workflow). The transport does not matter; the outcome does: evidence-backed understanding before synthesis and edits.

| Layer | Role |
| --- | --- |
| RepoNerve memory + code index | Source of truth (decisions, symbols, relationships) |
| MCP (`reponerve mcp`) | Delivers context into Agent chat |
| Skill (`.cursor/skills/reponerve/`) | Tells the agent when and how to load context before answering or changing code |

When MCP is connected, Cursor Agent calls tools directly. When MCP is off, the **same workflow** runs via `reponerve` CLI in the terminal — the skill defines both paths.

## Prerequisites

- Cursor with MCP support (Settings → Tools & MCP)
- `reponerve` installed and on your `PATH` (`go install ./cmd/reponerve` from this repo)
- RepoNerve initialized and scanned in the project:

```bash
reponerve init    # also installs Cursor skill + MCP configs
reponerve scan
```

`reponerve integrate` re-installs or merges IDE configs without re-initializing the database.

## Project setup (recommended)

This repository includes a project-scoped MCP config at `.cursor/mcp.json`:

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

`REPONERVE_WORKSPACE` points Cursor at the RepoNerve config directory (`.reponerve/config.yaml` and `memory.db`). Cursor resolves `${workspaceFolder}` to the project root.

### Enable in Cursor

1. Open this project in Cursor.
2. Go to **Cursor Settings → Tools & MCP** (or **Features → MCP**).
3. Confirm **reponerve** appears under project MCP servers with a green connected status.
4. If it does not connect, click refresh or restart Cursor.
5. Check **Output → MCP** for connection errors.

## Global setup (all projects)

To use RepoNerve in every Cursor workspace, add the same server block to `~/.cursor/mcp.json`:

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

Project-level `.cursor/mcp.json` overrides global config when both define the same server name.

## Agent skill (primary integration)

This repo ships a **Cursor Agent Skill** at `.cursor/skills/reponerve/`:

| File | Purpose |
| --- | --- |
| `SKILL.md` | Context-first workflow — pasted tasks, onboarding, anti-hallucination, token discipline |
| `reference.md` | MCP ↔ CLI command map and install instructions |

Cursor discovers skills by description. The skill is **not** MCP-only: it tells the agent to use MCP tools when connected, or equivalent `reponerve` CLI commands when not.

The skill encodes: understand → locate → constrain → scope → verify — using RepoNerve output (`structured.entity_briefings` for MCP; section headers for CLI) before reading or editing source.

### Install skill globally (other repositories)

```bash
mkdir -p ~/.cursor/skills/reponerve
cp -r /path/to/reponerve/.cursor/skills/reponerve/* ~/.cursor/skills/reponerve/
```

In each target repo: `reponerve init && reponerve scan`. Optionally add `.cursor/mcp.json` for MCP.

### Project rule

`.cursor/rules/reponerve.mdc` nudges agents to load the skill before explaining or editing this repository.

## Using RepoNerve in Agent chat

1. Open Cursor Agent chat (`Cmd+L` / `Ctrl+L`).
2. Ask naturally — the agent should follow `.cursor/skills/reponerve/SKILL.md` automatically.
3. When MCP is enabled, prefer RepoNerve MCP tools; otherwise the agent runs CLI equivalents from `reference.md`.

Example prompts:

- "Onboard me to this repo" → `onboard` / `reponerve onboard`
- Paste a full ticket → `ask` or `plan` / `reponerve plan "..."`
- "Why do we use SQLite?" → `ask` + `list_decisions`
- "Plan where to add a new MCP tool" → `plan` with scope and impacted files
- "What breaks if we change the MCP registry?" → `analyze_topic_impact` / `reponerve impact`

## Available MCP tools (38)

See `docs/copilot-chat-integration.md` for the full tool list. Highlights:

| Category | Tools |
| --- | --- |
| Memory | `list_decisions`, `get_decision`, `trace_decision`, `list_facts`, … |
| Graph | `trace_graph`, `analyze_impact`, `find_dependencies`, … |
| Intelligence | `discover_knowledge`, `recommend_reviewers`, `generate_learning_path` |
| Development Experience | `ask`, `explain`, `explain_file`, `plan`, `review`, `analyze_topic_impact`, `onboard` |

**Impact tools:** `analyze_impact` takes a graph entity ID (`node_id` + `node_type`). `analyze_topic_impact` takes a natural-language `subject` (same as `reponerve impact "subject"`).

## Troubleshooting

### Server not connecting

- Verify binary: `which reponerve`
- Verify workspace: `ls .reponerve/config.yaml .reponerve/memory.db`
- Test manually:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | REPONERVE_WORKSPACE="$(pwd)/.reponerve" reponerve mcp
```

You should receive JSON with 38 tools. Any non-JSON output on stdout breaks MCP.

### Empty or stale results

- Run `reponerve scan` after adding ADRs or new commits.
- Rebuild after code changes: `go install ./cmd/reponerve`
- Restart the MCP server in Cursor Settings.

### Wrong repository

- Ensure `reponerve init` was run in the project root Cursor opened.
- Confirm `REPONERVE_WORKSPACE` resolves to `<project>/.reponerve`.

## Further reading

- Universal understanding (north star): `docs/product/universal-understanding.md`
- Agent context contract: `docs/architecture/agent-context-contract.md`
- Agent skill: `.cursor/skills/reponerve/SKILL.md` (CLI map: `reference.md`)
- Project rule: `.cursor/rules/reponerve.mdc`
- [Cursor MCP documentation](https://cursor.com/docs/mcp)
- MCP troubleshooting: `docs/mcp/troubleshooting.md`
- Configuration examples: `docs/mcp/configuration-examples.md`
- Copilot Chat (VS Code): `docs/copilot-chat-integration.md`
