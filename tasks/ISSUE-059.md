# ISSUE-059 — Foundation Fixes

Status: Planned

Milestone: v0.10.0-alpha

Blocks: None (can start before ISSUE-057 architecture approval)

Part of: v1.0.0 (single product release)

---

# Objective

Ship quick wins and technical debt fixes on the existing Repository Intelligence stack before and during ISSUE-057.

---

# Deliverables

| Item | Location |
| --- | --- |
| Wire expertise detection into scan | `internal/ingestion/coordinator.go` |
| Expose agent search via CLI | `internal/agent/search/` → CLI |
| Expose graph impact via CLI | `internal/graph/impact/` → CLI |
| Fix scan help text (git + adr only) | `internal/cli/scan/scan.go` |
| Wire or remove FTS5 `memory_search` | `internal/storage/migrations.go` + readers |
| Document unified impact authority | `docs/architecture/` — agent vs graph impact |

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
