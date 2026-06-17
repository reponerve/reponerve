# ISSUE-060 — Token Intelligence Layer

Status: Complete

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
| Graph-aware compression | `internal/agent/compression/` — topic relevance, relationship boost, token-budget packing | Done |
| Output formats | `--format prose\|json\|caveman` + `--token-budget` on DE CLI | Done |
| MCP format parity | `format` + `token_budget` on all DE MCP tools; `generate_context` topic/budget | Done |
| Agent hooks | `reponerve hook install` — post-commit scan; `uninstall`, `status` | Done |
| Incremental scan | Code indexer + git scan state; hook triggers `reponerve scan` | Done |
| RTK guidance | `docs/product/token-economics.md` Layer 6 composability | Done |

---

# Acceptance Criteria

* Context pack size configurable via token budget — Done (`generate_context`, DE CLI/MCP)
* Caveman format reduces output size ≥50% vs prose for same evidence — Done (unit test)
* Hooks install for Cursor and Claude Code (documented) — Done
* Incremental scan updates code + memory indices — Done (via scan on commit)
* Tests for compression ranking and format rendering — Done

---

# Tag

Engineering checkpoint: `v0.13.0-alpha`
