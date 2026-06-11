# RepoNerve Implementation Status

Version: 1.0

Status: Draft

Updated: 2026-06-11

Related:

* `tasks/ISSUE-057.md`
* `docs/releases/v1.0.0-checklist.md`
* `docs/product/market-positioning.md`

---

# Summary

Repository Intelligence (Phases 0–6) is **substantially implemented and tested**. ISSUE-057 through ISSUE-062 are **not yet implemented** — v1.0.0 is blocked until all complete via `docs/roadmap/v1.0-iteration-plan.md`.

The documentation and strategic vision are ahead of the code in several areas. This document records an honest implementation snapshot.

---

# Shipped and Mature

| Layer | Status | Notes |
| --- | --- | --- |
| Init, config, storage, migrations | ✅ | SQLite v1–v8 |
| Git + ADR ingestion | ✅ | `scan` registers git + adr scanners |
| Memory extraction + linking | ✅ | Deterministic extractors, relationship linker |
| Query engine (readers) | ✅ | `internal/query/storage/` |
| Context engine | ✅ | generate, export |
| Ownership (contributors) | ✅ | Runs during scan |
| Graph intelligence | ✅ | Relationships, traversal, impact — MCP-primary |
| MCP server | ✅ | 27 tools, extensive tests |
| Agent QA (`ask`) | ⚠️ | Regex routing + git grep/blame fallbacks; partial |
| Agent onboarding, guidance, impact | ✅ | Used by `ask` |
| Agent search, workflow, session | ⚠️ | Implemented; not CLI-exposed |
| Repository intelligence services | ✅ | Discovery, learning, reviewers, change plan — MCP |

---

# v1.0 Blockers (ISSUE-057)

| Deliverable | Code state |
| --- | --- |
| Code Intelligence (`go/ast`, symbols, call graph) | Done — indexer, linker, service API |
| Repository-Code Linking | Not started |
| Feature Understanding | Not started |
| Development Experience CLI | `explain` is stub; `plan`, `impact`, `review`, symbol explain commands absent |
| Expertise detection in scan pipeline | Implemented but **not wired** in coordinator |

---

# Known Gaps and Debt

| Issue | Severity | Detail |
| --- | --- | --- |
| Code intelligence missing | Blocker | No AST parsing in codebase |
| `explain` CLI stub | Blocker | Prints message only |
| Expertise not in scan | High | `ask` ownership-by-expertise often empty |
| FTS5 `memory_search` | Done | Rebuilt on scan; queried by `reponerve search` |
| `agent/compression` naive | Medium | List truncation only; not graph-aware |
| Dual impact implementations | Medium | `agent/impact` vs `graph/impact` |
| `agent/compression` orphaned from prod | Low | No production imports outside tests |
| `AIConfig` unused | Low | Config field never consumed |
| Scan help text overstated | Low | Claims code/PRs; only git + adr registered |

---

# v1.0 Iteration Plan

All scope ships in v1.0.0 via v0.x alpha tags:

| Tag | Issue | Status |
| --- | --- | --- |
| v0.10.0-alpha | ISSUE-059 Foundation fixes | Complete |
| v0.11–v0.12.0-alpha | ISSUE-057 Code + DE | Not started |
| v0.13.0-alpha | ISSUE-060 Token Intelligence | Not started |
| v0.14.0-alpha | ISSUE-061 Graph + Session Memory | Not started |
| v0.15.0-alpha | ISSUE-062 Multi-language | Not started |
| v1.0.0 | Release | Blocked |

See `docs/roadmap/v1.0-iteration-plan.md`.

---

# Architecture Approval Gates

Implementation must not begin until:

* `docs/architecture/issue-057-architecture.md` — approved
* ARCH-001 (`docs/architecture/architecture-overview.md` v1.1) — approved

---

# Test Coverage Impression

* 45+ `_test.go` files under `internal/`
* 16 integration tests under `tests/integration/`
* MCP server: extensive coverage
* `internal/cli/ask/`: limited CLI-level tests
* `internal/memory/storage/`: no direct unit tests

---

# Documentation vs Code

| Documented | In code |
| --- | --- |
| 10 Development Experience CLI commands | 1 partial (`ask`), 1 stub (`explain`) |
| Code indexer in ingestion | Not registered |
| Repository-code linker | Not present |
| Token-budget compression | Truncation only |
| Caveman output format | Not present |

Keep this document updated as ISSUE-057 lands.
