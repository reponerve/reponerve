# ISSUE-061 — Evidence Graph & Session Memory

Status: Planned

Milestone: v0.14.0-alpha

Depends On: ISSUE-057, ISSUE-060 (recommended)

Part of: v1.0.0 (single product release)

---

# Objective

Deliver graph discovery, exploration, and repository-scoped session memory — Graphify and ICM capabilities on RepoNerve's evidence model.

---

# Deliverables

## Evidence Graph Intelligence

| Capability | Package / surface |
| --- | --- |
| Community detection | `internal/graph/communities/` |
| Surprising connections, god nodes | `internal/graph/discovery/` |
| Token-budget traversal | `internal/graph/traversal/budget.go` |
| `reponerve explore` | CLI + HTML export |
| MCP tools | `discover_surprises`, `suggest_questions`, `query_graph` |

## Session Memory

| Capability | Surface |
| --- | --- |
| `reponerve remember` / `reponerve forget` | CLI + MCP |
| Session writeback | Q&A creates traceable Facts with provenance |
| Temporal relevance | Access-aware memory ranking |
| Agent handoff bundles | Deterministic context transfer between sessions |
| Workflow templates | Fixed presets: onboarding, review prep, change prep |

---

# Acceptance Criteria

* Communities detected on knowledge graph with deterministic ordering
* `reponerve explore` renders HTML graph
* `remember`/`forget` create/update memory with evidence
* Handoff bundle export/import round-trips
* Workflow templates invoke existing agent/workflow services

---

# Tag

Engineering checkpoint: `v0.14.0-alpha`
