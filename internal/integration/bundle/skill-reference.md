# RepoNerve skill reference

Use **MCP tools** when connected; otherwise use **CLI** (same semantics).

## Development Experience

| Goal | MCP tool | CLI (no MCP) |
| --- | --- | --- |
| Pasted task / plan | `plan` | `reponerve plan "..." --json` |
| Question / what-is | `ask` | `reponerve ask "..." --json` |
| Topic explain | `explain` | `reponerve explain "..." --json` |
| File | `explain_file` | `reponerve explain-file "path" --json` |
| Function | `explain_function` | `reponerve explain-function "Name" --package pkg --json` |
| Struct | `explain_struct` | `reponerve explain-struct "Name" --package pkg --json` |
| Interface | `explain_interface` | `reponerve explain-interface "Name" --json` |
| Type alias | `explain_type` | `reponerve explain-type "Name" --json` |
| Topic impact | `analyze_topic_impact` | `reponerve impact "subject" --json` |
| Review prep | `review` | `reponerve review "topic" --json` |
| PR / CI context | `pr_context` | `reponerve pr-context path/to/file.go --json` |
| Discipline policy | — | `reponerve discipline-policy --json` |
| Day-one orientation | `onboard` | `reponerve onboard --json` |

## Native Development Discipline

Bundled on `reponerve init` — no separate discipline skills required.

| Intent | MCP / CLI |
| --- | --- |
| Feature / ticket | `plan` |
| Ship / merge / PR | `ship_check`, `review`, `pr_context` |
| Reuse before new code | `reuse_check` |
| Repo discipline policy | `discipline-policy` (after `scan`) |

Rules: `.cursor/rules/development-discipline.mdc`, `coding-guidelines.mdc`

## Token discipline

| Task | Prefer | Avoid |
| --- | --- | --- |
| Verify fix / "is this correct?" | `explain_function` / `explain-file` with `--package` | broad `ask`, full `plan`/`review` JSON |
| Architecture / why | `ask` + `format: compact`, `token_budget: 1500` | grep → bulk file reads |
| Pasted ticket | `plan` | ad-hoc exploration |

## Repository memory (high value)

| MCP | CLI |
| --- | --- |
| `list_decisions` | `reponerve memory list-decisions` |
| `get_decision` | `reponerve memory get-decision <id>` |
| `trace_decision` | `reponerve memory trace-decision <id>` |
| `list_facts` | `reponerve memory list-facts` |
| `generate_context` | `reponerve context generate` |

## Install this skill in other repos

`reponerve init` installs project and global skill files automatically. Install the CLI first — see **`docs/install.md`** (no Go required).

```bash
curl -fsSL https://raw.githubusercontent.com/reponerve/reponerve/main/scripts/install.sh | bash
reponerve init && reponerve scan
```

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
