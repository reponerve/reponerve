# RFC-008: Scoped Monorepo Scan

Status: Accepted  
Date: 2026-06-24

Related:

* `internal/code/indexer/module.go`
* `docs/rfc/RFC-004-team-delivery-intelligence.md`

---

## Problem

`go.work` monorepos may contain many modules. Full code re-index on every `scan` is slow; developers often change one module. Existing incremental skip is repo-wide.

## Decision

Add scoped code indexing to `reponerve scan`:

```bash
reponerve scan --modules github.com/org/foo,github.com/org/bar
reponerve scan --changed   # modules touched by git working tree + last commit
```

- Git/ADR/ownership pipelines unchanged (full incremental git scan)
- Code indexer replaces entities only for requested `module_path` values
- Repository-code linker runs full pass after scoped index

---

## Non-goals

- Per-package scan for non-Go monorepos (npm/pnpm workspaces) in v1.4
- Skipping git ingestion when `--changed` has no code files

---

## Success criteria

| Verify |
| --- |
| `scan --modules` re-indexes one module in a `go.work` repo |
| `scan --changed` resolves modules from `git diff` paths |
| Other modules' entities remain in `code_entities` |

---

## v1.4.0 bundle

| Item | Surface |
| --- | --- |
| Scoped scan | `reponerve scan --modules`, `--changed` |
