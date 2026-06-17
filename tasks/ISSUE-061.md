# ISSUE-061 — Evidence Graph & Session Memory

Status: Complete — v0.14.0-alpha

Milestone: v0.14.0-alpha

Depends On: ISSUE-057, ISSUE-060 (recommended)

Part of: v1.0.0 (single product release)

---

# Objective

Deliver graph discovery, exploration, and repository-scoped session memory on RepoNerve's evidence model.

---

# Deliverables

## Evidence Graph Intelligence

| Capability | Package / surface | Status |
| --- | --- | --- |
| Community detection | `internal/graph/communities/` | Done |
| Surprising connections, god nodes | `internal/graph/discovery/` | Done |
| Token-budget traversal | `internal/graph/traversal/budget.go` | Done |
| `reponerve explore` | CLI + HTML export | Done |
| MCP tools | `discover_surprises`, `suggest_questions`, `query_graph` | Done |

## Session Memory

| Capability | Surface | Status |
| --- | --- | --- |
| `reponerve remember` / `reponerve forget` | CLI + MCP | Done |
| Session writeback | Q&A creates traceable Facts with provenance | Done |
| Temporal relevance | Access-aware memory ranking | Done |
| Agent handoff bundles | Deterministic context transfer between sessions | Done |
| Workflow templates | Fixed presets: onboarding, review prep, change prep | Done |

---

# Acceptance Criteria

* Communities detected on knowledge graph with deterministic ordering — Done
* `reponerve explore` renders HTML graph — Done
* `remember`/`forget` create/update memory with evidence — Done
* Handoff bundle export/import round-trips — Done
* Workflow templates invoke existing agent/workflow services — Done

---

# Tag

Engineering checkpoint: `v0.14.0-alpha`
