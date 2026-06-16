---
name: reponerve
description: >-
  Load evidence-backed repository context from RepoNerve before explaining or
  editing code. Use for day-one onboarding, pasted task descriptions, architecture
  questions, symbols, decisions, planning, impact, and review. Enforces anti-
  hallucination and token discipline so weak and strong models work without
  context worry.
---

# RepoNerve — Zero Context Worry

**North star:** Day-one developers, assigned engineers, and AI (including weak models) understand and work on the repo **without guessing**. RepoNerve carries context; you narrate and edit from evidence.

**Contract:** `docs/architecture/agent-context-contract.md`  
**Product vision:** `docs/product/universal-understanding.md`

**Principles:** Understanding first. Evidence second. AI third.

---

## Pasted task description (start here)

Paste the full ticket text to `ask` or `plan` — `ask` auto-routes to `task_plan`:

```text
ask("<full pasted task>")   → answer_type: task_plan + plan + briefings
  OR
plan("<full pasted task>")
```

Then:

```text
1. Follow plan.suggested_steps
2. explain_file / explain_* on starting_points ONLY
3. analyze_topic_impact on risky areas
4. edit within scope
5. review("<task topic>")
```

Do **not** grep the repo first. Do **not** implement until `agent.completeness` is `full`.

---

## Day-one onboarding

One call:

```text
onboard()                          # orientation + key decisions
onboard("Add OAuth login")           # + assignment plan + briefings
```

Or stepwise: `ask` → `list_decisions` → `explain` → `plan`.

---

## MCP envelope (read in order)

1. `structured` — facts (`entity_briefings`, plan scope, links, evidence)
2. `agent` — completeness, edit gates, next tools, anti-hallucination guidance
3. `formatted` — human summary only

| `completeness` | You must |
| --- | --- |
| `full` | Answer/edit from structured; avoid bulk file reads |
| `partial` | Run `recommended_next_tools` before editing |
| `retrieval_only` | **Stop** — do not answer confidently or edit; use `ask`/`explain`/`plan` |

If `must_use_before_edit` is true, load and understand context before any code change.

---

## Anti-hallucination (mandatory)

- **Only cite** paths, types, ADRs, and relationships in `structured`
- **Missing fact** → say "RepoNerve has no evidence for X" and query more — never invent
- **Homonyms** → compare all `entity_briefings`; edit only the matching `defined_in`
- **No confident prose** from `search_summary` or hit counts
- **No architecture narrative** without `related_decisions` / evidence when available

Hallucination wastes tokens and causes wrong edits. When unsure, query RepoNerve again — do not explore blindly.

---

## Token discipline

```text
BAD:  grep → 20 files → guess → wrong fix → 10 more files
GOOD: plan → briefings → 2 scoped files → review
```

- One RepoNerve pass before bulk file reads
- Stop exploring when briefings already have `defined_in` and relationships
- Escalate via `recommended_next_tools`, not ad-hoc search

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

| Goal | Tool |
| --- | --- |
| Pasted task / assignment | `plan` |
| What is X? / How does X work? | `ask` |
| Explain topic or file | `explain`, `explain_file` |
| Explain symbol | `explain_struct`, `explain_function`, … (use `package_path` / `--package` for homonyms) |
| Refactor impact | `analyze_topic_impact` |
| Pre-merge check | `review` |
| ADRs | `list_decisions`, `trace_decision` |

---

## Prerequisites

```bash
reponerve init && reponerve scan
```

Restart MCP after `go install ./cmd/reponerve`.
