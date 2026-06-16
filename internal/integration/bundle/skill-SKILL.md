---
name: reponerve
description: >-
  Load evidence-backed repository context from RepoNerve before explaining or
  editing code. Use for day-one onboarding, pasted task descriptions, architecture
  questions, symbols, decisions, planning, impact, and review — via MCP tools OR
  reponerve CLI. Enforces anti-hallucination and token discipline so weak and
  strong models work without context worry. Install in any repo after init+scan.
---

# RepoNerve — Zero Context Worry

**North star:** Day-one developers, assigned engineers, and AI (including weak models) understand and work on the repo **without guessing**. RepoNerve carries context; you narrate and edit from evidence.

**Contract:** `docs/architecture/agent-context-contract.md`  
**Product vision:** `docs/product/universal-understanding.md`  
**CLI/MCP map:** `.cursor/skills/reponerve/reference.md`  
**All IDEs / LLMs:** `docs/ai-chat-integration.md`

**Principles:** Understanding first. Evidence second. AI third.

---

## How to use RepoNerve (any IDE chat)

RepoNerve is a **Cursor Agent Skill** and an **MCP server** for all major IDEs. Use it in every conversation about this codebase.

| Mode | When | How |
| --- | --- | --- |
| **MCP** (preferred) | reponerve server connected (Cursor, Copilot, JetBrains, …) | Call MCP tools (`ask`, `explain`, `plan`, …) from chat |
| **CLI** | MCP off or unavailable | Run `reponerve <command>` in the terminal |

**Direct chat:** type natural language in the IDE assistant — you do not need CLI syntax. See `docs/ai-chat-integration.md` for VS Code, JetBrains, Windsurf, Continue, Claude, and web-LLM fallbacks.

Same workflow either way. **Never skip RepoNerve** and fall back to blind grep.

```bash
# One-time per repo — init installs skill + MCP automatically
reponerve init && reponerve scan
```

`reponerve integrate` refreshes IDE configs without touching the database.

---

## Pasted task description (start here)

Paste the full ticket text to `ask` or `plan` — `ask` auto-routes to `task_plan`:

```text
MCP:  ask({ topic: "<full pasted task>" })
CLI:  reponerve ask "<full pasted task>"
  OR
MCP:  plan({ topic: "..." })
CLI:  reponerve plan "<full pasted task>"
```

Then:

```text
1. Follow plan.suggested_steps
2. explain_file / explain_* on starting_points ONLY
3. analyze_topic_impact on risky areas
4. edit within scope
5. review("<task topic>")
```

Do **not** grep the repo first. Do **not** implement until context is complete (`agent.completeness` = `full` for MCP, or plan/briefings present for CLI).

---

## Day-one onboarding

```text
MCP:  onboard()  or  onboard({ topic: "Add OAuth login" })
CLI:  reponerve onboard
      reponerve onboard "Add OAuth login"
```

Or stepwise: `ask` → `list_decisions` → `explain` → `plan`.

---

## MCP envelope (when using MCP)

Read in order:

1. `structured` — facts (`entity_briefings`, plan scope, links, evidence)
2. `agent` — completeness, edit gates, next tools, anti-hallucination guidance
3. `formatted` — human summary only

| `completeness` | You must |
| --- | --- |
| `full` | Answer/edit from structured; avoid bulk file reads |
| `partial` | Run `recommended_next_tools` before editing |
| `retrieval_only` | **Stop** — do not answer confidently or edit; use `ask`/`explain`/`plan` |

If `must_use_before_edit` is true, load context before any code change.

---

## CLI output (when using terminal)

CLI prints sections: `ENTITY BRIEFINGS`, `CODE CONTEXT`, `REPOSITORY CONTEXT`, `Evidence:`. Treat those as authoritative. If a fact is missing, say so — do not invent.

---

## Anti-hallucination (mandatory)

- **Only cite** paths, types, ADRs, and relationships in RepoNerve output
- **Missing fact** → say "RepoNerve has no evidence for X" and query more — never invent
- **Homonyms** → compare all `entity_briefings`; use `--package` / `package_path`; edit only matching `defined_in`
- **No confident prose** from `search_summary` or hit counts alone
- **No architecture narrative** without decisions / evidence when available

---

## Token discipline

```text
BAD:  grep → 20 files → guess → wrong fix → 10 more files
GOOD: plan → briefings → 2 scoped files → review
```

- One RepoNerve pass before bulk file reads
- Stop exploring when briefings already have `defined_in` and relationships
- Escalate via `recommended_next_tools` or the next CLI command in `reference.md`

---

## Context checklist

- [ ] Identity — `qualified_name`, `role`
- [ ] Location — `defined_in`
- [ ] Shape — `fields`, `signature`
- [ ] Relationships — `producers`, `consumers`
- [ ] Constraints — `related_decisions`
- [ ] Scope — `starting_points`, `impacted_areas` (from plan)

---

## Query guide

| Goal | MCP | CLI |
| --- | --- | --- |
| Pasted task | `plan` | `reponerve plan "..."` |
| What is X? | `ask` | `reponerve ask "..."` |
| Explain topic/file | `explain`, `explain_file` | `reponerve explain`, `explain-file` |
| Explain symbol | `explain_*` + `package_path` | `explain-*` + `--package` |
| Refactor impact | `analyze_topic_impact` | `reponerve impact` |
| Pre-merge | `review` | `reponerve review` |
| ADRs | `list_decisions` | `reponerve memory list-decisions` |

Full table: `reference.md`

---

## Prerequisites

```bash
go install ./cmd/reponerve   # or release binary
reponerve init && reponerve scan   # init installs skill + MCP configs
```

Restart MCP in your IDE after reinstalling the binary.
