# v1.0 Implementation Plan

Status: In Progress

Updated: 2026-06-11

Architecture approval: **Approved** (`issue-057-architecture.md`, ARCH-001)

Authoritative iteration map: `docs/roadmap/v1.0-iteration-plan.md`

---

# Current Sprint

## Track A — ISSUE-059 (`v0.10.0-alpha`) — COMPLETE

| Task | Status |
| --- | --- |
| Wire expertise detection into scan | Done |
| `reponerve search` CLI | Done |
| `reponerve impact` CLI (graph impact) | Done |
| Fix scan help text | Done |
| Wire FTS5 `memory_search` into search + scan rebuild | Done |

Ready to tag `v0.10.0-alpha`.

## Track B — ISSUE-057 Step 1 (`v0.11.0-alpha`) — IN PROGRESS

| Task | Status |
| --- | --- |
| `internal/code/models` | Done |
| Migration v9 (code tables) | Done |
| Code storage stores (SQLite) | Done |
| Entity ID determinism tests | Done |
| Go parser / indexer | **Next** |
| Register code scanner in ingestion | Next |

---

# Sequence After This Sprint

1. ISSUE-057 steps 2–4: indexer, linker, Code Intelligence service → `v0.11.0-alpha`
2. ISSUE-057 steps 5–9: Development Experience + CLI → `v0.12.0-alpha`
3. ISSUE-060 → `v0.13.0-alpha`
4. ISSUE-061 → `v0.14.0-alpha`
5. ISSUE-062 → `v0.15.0-alpha`
6. Release audits → `v1.0.0`

---

# Authority Rules (unchanged)

* Graph `impact` CLI / MCP = canonical for graph traversal impact
* Agent `impact` service = memory-relationship impact (used by `ask` only)
* Code Intelligence = authoritative for symbols (after indexer ships)
* Repository Intelligence = authoritative for decisions, ownership, events
