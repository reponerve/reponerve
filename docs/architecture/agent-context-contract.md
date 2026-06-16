# Agent Context Contract

Status: Active

Related:

* `docs/architecture/development-experience-contracts.md`
* `docs/architecture/agent-native-repository-intelligence.md`
* `.cursor/skills/reponerve/SKILL.md`

---

# Purpose

RepoNerve exists to give AI **proper repository context** before synthesis and edits.

The agent context contract defines:

1. What context RepoNerve must supply
2. How context is packaged for agents (MCP, CLI, future adapters)
3. What agents must do with that context

Transport does not matter. Outcome does: **evidence-backed understanding that enables precise fixes.**

**North star:** Anyone — day-one developer, assigned engineer, or AI (including weak models) — understands and works on the repository without context worry. See `docs/product/universal-understanding.md`.

---

# Principles

```text
Understanding First
        ↓
Evidence Second
        ↓
AI Third
```

RepoNerve must not emit free-form opinions, guessed ownership, or narrative Purpose/History fields.

Agents must not edit code from filename guesses when RepoNerve returned structured briefings, links, or change scope.

---

# Context Dimensions

Every agent-ready response should cover as many dimensions as the topic allows:

| Dimension | What the agent needs | Primary sources |
| --- | --- | --- |
| **Identity** | What the symbol/topic is, entity type, role | `entity_briefings`, `code_context` |
| **Location** | Package, file, line range | `defined_in`, `code_context.files` |
| **Shape** | Fields, signature, members | `entity_briefings.fields`, `signature` |
| **Relationships** | Callers, callees, dependencies | `producers`, `consumers`, `dependencies` |
| **Constraints** | ADRs, facts, events | `repository_context`, `related_decisions` |
| **Cross-authority links** | Memory ↔ code connections | `repository_code_links` |
| **Change scope** | Impacted areas, starting points | `plan`, `impact` outputs |
| **Review scope** | Reviewers, affected knowledge | `review` output |
| **Evidence** | Traceable upstream payloads | `evidence`, `source_services` |

Missing dimensions must be reflected in `agent.completeness` — never implied as complete when partial.

---

# Response Envelope (MCP)

Development Experience MCP tools return:

```json
{
  "formatted": "human-readable summary",
  "structured": { },
  "agent": {
    "kind": "concept_explanation",
    "completeness": "full",
    "must_use_before_edit": true,
    "guidance_for_agent": [
      "Anchor edits on entity_briefings[].defined_in",
      "Disambiguate homonyms before changing code"
    ],
    "recommended_next_tools": ["plan", "analyze_topic_impact"]
  }
}
```

| Field | Meaning |
| --- | --- |
| `formatted` | Display layer for humans |
| `structured` | Authoritative machine context — **read first** |
| `agent.kind` | Response classifier (`concept_explanation`, `unified_explanation`, `plan`, `impact`, `review`, …) |
| `agent.completeness` | `full` \| `partial` \| `retrieval_only` |
| `agent.must_use_before_edit` | Agent must load this context before modifying code |
| `agent.guidance_for_agent` | Deterministic instructions for synthesis and edits |
| `agent.recommended_next_tools` | Follow-up RepoNerve tools to strengthen context |

---

# Completeness Levels

| Level | When | Agent obligation |
| --- | --- | --- |
| `full` | Entity briefings and/or plan/impact scope present | Edit only within scoped files/symbols; cite evidence |
| `partial` | Code or repository context without briefings | Run `explain_*` or `ask` ("What is X?") before editing |
| `retrieval_only` | Search summary or thin Q&A hit list | Do not treat as understanding; escalate to `explain` or `plan` |

---

# Command Requirements

## ask

Must return `answer_type` and prefer `concept_explanation` for definition questions.

When code entities match the subject, must include `entity_briefings` with identity, location, shape, and relationships.

`search_summary` is a fallback — mark `completeness: retrieval_only`.

Triggers for concept routing include: *what is*, *what does*, *how does … work*, *tell me about*.

## explain / explain-*

Must return unified `DevelopmentExplanation` with:

* `entity_briefings` when symbol/topic resolves to code entities
* `code_context` when code hits exist
* `repository_context` when memory hits exist
* `repository_code_links` when cross-authority links exist

## plan

Must return impacted areas, relevant decisions, starting points, and links.

`must_use_before_edit: true` for implementation tasks.

## analyze_topic_impact (impact)

Must return dependent areas, code dependencies, impacted decisions, owners, links.

`must_use_before_edit: true` before refactors.

## review

Must return reviewers, affected areas, related knowledge, links.

Run after edits; before merge when topic is non-trivial.

---

# Agent Workflow (Mandatory)

Before **answering** a repository question:

1. Load RepoNerve context (`ask` or `explain`)
2. Read `structured` then `agent`
3. Synthesize from briefings and links — not from raw search hits

Before **changing** code:

1. **Understand** — `ask` / `explain_*` with `full` or `partial` context
2. **Constrain** — `list_decisions` / `trace_decision` when architecture applies
3. **Scope** — `plan` for new work; `analyze_topic_impact` for refactors
4. **Edit** — only files/symbols in scope
5. **Verify** — `review` on the topic

Skip file reads only when `completeness` is `full` and the question does not require line-level logic.

---

# Task Intake (Pasted Description)

When the user provides a task — ticket text, assignment, feature request — **start with `plan`**, not repository exploration.

```text
plan("<pasted task>")
  → explain / ask on unknown terms
  → analyze_topic_impact on risky areas
  → explain_* on starting_points only
  → edit within scope
  → review
```

Do not implement until `agent.completeness` is `full` for the symbols and files in scope.

---

# First-Day Onboarding

```text
ask / explain (orientation)
  → list_decisions / trace_decision (architecture)
  → explain (domains they will touch)
  → plan (first assignment)
```

---

# Anti-Hallucination Rules

RepoNerve and agents share responsibility for zero fabrication:

| Rule | RepoNerve | Agent |
| --- | --- | --- |
| No invented ADRs or file paths | Only emit indexed evidence | Only cite `structured` fields |
| No hidden ambiguity | Multiple briefings for homonyms | Compare before editing |
| No false completeness | Set `completeness` honestly | Do not edit on `retrieval_only` |
| No exploration tax | Pre-digest at scan | Query RepoNerve before bulk file reads |
| Missing evidence | Return partial + next tools | Say "no evidence" — do not guess |

When `structured` lacks a fact, the correct output is **narrow the question or run scan** — not a confident invention.

---

# Anti-Patterns

| Invalid | Valid |
| --- | --- |
| Edit `RepositoryContext` in the first package match | Use `defined_in` from the correct briefing |
| Implement from `search_summary` hit counts | Route to `concept_explanation` |
| Refactor without `analyze_topic_impact` | Load impact scope first |
| Ignore `related_decisions` | Treat ADRs as hard constraints |
| Narrate without citing `source_services` | Trace conclusions to evidence |

---

# Entity Briefing Minimum

Each `EntityBriefing` must include when available:

* `qualified_name`, `entity_type`, `layer`, `role`
* `defined_in` (file:line)
* `fields` or `signature` for types
* `producers`, `consumers` when relationships exist
* `related_decisions` when links exist

Homonyms must return multiple briefings — never fail silently on ambiguity.

---

# Acceptance

Agent context contract is satisfied when:

* All DE MCP tools return `formatted`, `structured`, and `agent` metadata
* Definition questions resolve to `concept_explanation` with briefings when indexed
* Plan and impact set `must_use_before_edit: true`
* `search_summary` is never the final context for code edits
* Skill and integration docs reference this contract

See `docs/examples/development-experience.md` and `docs/product/universal-understanding.md` for worked examples.
