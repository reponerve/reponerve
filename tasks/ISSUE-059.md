# ISSUE-059 — Foundation Fixes

Status: Complete

Milestone: v0.10.0-alpha

Blocks: None (can start before ISSUE-057 architecture approval)

Part of: v1.0.0 (single product release)

---

# Objective

Ship quick wins and technical debt fixes on the existing Repository Intelligence stack before and during ISSUE-057.

---

# Deliverables

| Item | Status | Location |
| --- | --- | --- |
| Wire expertise detection into scan | Done | `internal/ingestion/coordinator.go` |
| Expose agent search via CLI | Done | `internal/cli/search/` |
| Expose graph impact via CLI | Done | `internal/cli/impactcmd/` |
| Fix scan help text (git + adr only) | Done | `internal/cli/scan/scan.go` |
| Wire FTS5 `memory_search` into search + scan rebuild | Done | `internal/storage/sqlite/memory_search_store.go`, `internal/memory/searchindex/`, `internal/agent/search/service.go` |
| Document unified impact authority | Done | `internal/cli/impactcmd/impact.go` long help |

---

# Acceptance Criteria

* `reponerve scan` populates expertise store
* `reponerve ask "who owns X"` returns expertise after scan
* New CLI commands documented in `docs/architecture/cli-reference-v1.md`
* `go test ./...` passes

---

# Tag

Engineering checkpoint: `v0.10.0-alpha`

Not a product release. v1.0.0 remains gated on full scope in `docs/roadmap/v1.0-iteration-plan.md`.
