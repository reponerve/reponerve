# ISSUE-060 — Token Intelligence Layer

Status: Planned

Milestone: v0.13.0-alpha

Depends On: ISSUE-057 (Development Experience core)

Part of: v1.0.0 (single product release)

---

# Objective

Deliver token-efficient understanding delivery so premium LLM models stay within context limits.

See `docs/product/token-economics.md`.

---

# Deliverables

| Capability | Description |
| --- | --- |
| Graph-aware compression | Upgrade `internal/agent/compression/` — relevance-ranked, token-budget |
| Output formats | `--format caveman\|prose\|json` on CLI and MCP |
| Agent hooks | `reponerve hook install` — post-commit scan, session context inject |
| Incremental scan | Re-index changed files on commit without full scan |
| RTK guidance | Document composition: RTK (shell) + RepoNerve (understanding) |

---

# Acceptance Criteria

* Context pack size configurable via token budget
* Caveman format reduces output size ≥50% vs prose for same evidence
* Hooks install for Cursor and Claude Code (documented)
* Incremental scan updates code + memory indices
* Tests for compression ranking and format rendering

---

# Tag

Engineering checkpoint: `v0.13.0-alpha`
