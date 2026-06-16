# RepoNerve skill reference

Use **MCP tools** when connected; otherwise use **CLI** (same semantics).

## Development Experience

| Goal | MCP tool | CLI |
| --- | --- | --- |
| Pasted task / plan | `plan` | `reponerve plan "..."` |
| Question / what-is | `ask` | `reponerve ask "..."` |
| Topic explain | `explain` | `reponerve explain "..."` |
| File | `explain_file` | `reponerve explain-file "path"` |
| Function | `explain_function` | `reponerve explain-function "Name" --package pkg` |
| Struct | `explain_struct` | `reponerve explain-struct "Name" --package pkg` |
| Interface | `explain_interface` | `reponerve explain-interface "Name"` |
| Type alias | `explain_type` | `reponerve explain-type "Name"` |
| Topic impact | `analyze_topic_impact` | `reponerve impact "subject"` |
| Review prep | `review` | `reponerve review "topic"` |
| Day-one orientation | `onboard` | `reponerve onboard` |

## Repository memory (high value)

| MCP | CLI |
| --- | --- |
| `list_decisions` | `reponerve memory list-decisions` |
| `get_decision` | `reponerve memory get-decision <id>` |
| `trace_decision` | `reponerve memory trace-decision <id>` |
| `list_facts` | `reponerve memory list-facts` |
| `generate_context` | `reponerve context generate` |

## Install this skill in other repos

`reponerve init` installs project and global skill files automatically. To refresh:

```bash
reponerve integrate          # merge MCP configs, skip existing skill files
reponerve integrate --force  # overwrite skill files
```

Manual copy (only if needed):

```bash
mkdir -p ~/.cursor/skills/reponerve
cp -r /path/to/reponerve/.cursor/skills/reponerve/* ~/.cursor/skills/reponerve/
```

Then in each target repo: `reponerve init && reponerve scan`, and add `.cursor/mcp.json` (see `docs/cursor-integration.md`).

## Response contract

MCP returns `structured`, `agent`, and `formatted`. CLI prints human text — still follow anti-hallucination rules: only cite evidence RepoNerve returned.

See `docs/architecture/agent-context-contract.md`.
