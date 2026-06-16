# Cursor Integration

RepoNerve gives AI **proper repository context** — what symbols mean, where they live, how they connect, and which decisions constrain changes. With that context, agents understand the codebase accurately and fix things precisely instead of guessing from filenames or search hits.

Cursor connects through **MCP** (tools) and a project **skill** (how to use the context). The transport does not matter; the outcome does: evidence-backed understanding before synthesis and edits.

| Layer | Role |
| --- | --- |
| RepoNerve memory + code index | Source of truth (decisions, symbols, relationships) |
| MCP (`reponerve mcp`) | Delivers context into Agent chat |
| Skill (`.cursor/skills/reponerve/`) | Tells the agent when and how to load context before answering or changing code |

When connected, Cursor Agent can query repository memory, traverse the knowledge graph, and use Development Experience tools (`ask`, `explain`, `plan`, `review`, and more) without shelling out to the CLI.

## Prerequisites

- Cursor with MCP support (Settings → Tools & MCP)
- `reponerve` installed and on your `PATH` (`go install ./cmd/reponerve` from this repo)
- RepoNerve initialized and scanned in the project:

```bash
reponerve init
reponerve scan
```

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

## Agent skill

This repo ships `.cursor/skills/reponerve/SKILL.md`. Cursor loads it automatically when you ask about architecture, symbols, decisions, or changes in this project.

The skill encodes the **context-first workflow**: understand → locate → constrain → scope → verify — using RepoNerve output (especially `structured.entity_briefings`) before reading or editing source.

Copy the skill to `~/.cursor/skills/reponerve/` if you want the same behavior in other repositories (with RepoNerve initialized there).

## Using RepoNerve in Agent chat

1. Open Cursor Agent chat (`Cmd+L` / `Ctrl+L`).
2. Enable MCP tools for the conversation (tools picker in the chat UI).
3. Ask natural-language questions; the agent should load RepoNerve context first, then answer or edit.

Example prompts:

- "Use RepoNerve to understand RepositoryContext before explaining it."
- "Why do we use SQLite? Check repository memory and ADRs."
- "Plan where to add a new MCP tool — use RepoNerve for scope and impacted files."
- "What breaks if we change the MCP registry? Use impact analysis, then propose a precise fix."

## Available tools (37)

See `docs/copilot-chat-integration.md` for the full tool list. Highlights:

| Category | Tools |
| --- | --- |
| Memory | `list_decisions`, `get_decision`, `trace_decision`, `list_facts`, … |
| Graph | `trace_graph`, `analyze_impact`, `find_dependencies`, … |
| Intelligence | `discover_knowledge`, `recommend_reviewers`, `generate_learning_path` |
| Development Experience | `ask`, `explain`, `explain_file`, `plan`, `review`, `analyze_topic_impact` |

**Impact tools:** `analyze_impact` takes a graph entity ID (`node_id` + `node_type`). `analyze_topic_impact` takes a natural-language `subject` (same as `reponerve impact "subject"`).

## Troubleshooting

### Server not connecting

- Verify binary: `which reponerve`
- Verify workspace: `ls .reponerve/config.yaml .reponerve/memory.db`
- Test manually:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | REPONERVE_WORKSPACE="$(pwd)/.reponerve" reponerve mcp
```

You should receive JSON with 37 tools. Any non-JSON output on stdout breaks MCP.

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
- Agent skill: `.cursor/skills/reponerve/SKILL.md`
- [Cursor MCP documentation](https://cursor.com/docs/mcp)
- MCP troubleshooting: `docs/mcp/troubleshooting.md`
- Configuration examples: `docs/mcp/configuration-examples.md`
- Copilot Chat (VS Code): `docs/copilot-chat-integration.md`
