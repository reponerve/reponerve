# RFC-007: Freshness Doctor

Status: Accepted  
Date: 2026-06-24

Related:

* `docs/rfc/RFC-004-team-delivery-intelligence.md`
* `docs/install.md`

---

## Problem

Agents and developers use stale repository memory when `scan` was never run, git moved ahead of `scan_state`, or the code index is behind working tree changes. Failures surface late as wrong answers, not as a clear health signal.

## Decision

Ship **`reponerve doctor`** — deterministic freshness guard with CLI + MCP `doctor` tool.

### Checks

| Check | Status levels |
| --- | --- |
| Workspace initialized (`.reponerve/`, `memory.db`) | fail / ok |
| Scan state present | warn / ok |
| Git HEAD vs `last_scan_commit` | warn if stale |
| Code index vs source mtimes | warn if stale |
| Discipline policy present | warn if missing |
| Post-commit hook | advisory only |

Output: structured `DoctorResult` with `ok`, `checks[]`, `recommendations[]`.

---

## Non-goals

- Auto-running `scan`
- Cloud telemetry
- LLM-generated health narrative

---

## Success criteria

| Verify |
| --- |
| `reponerve doctor --json` after init+scan returns `ok: true` |
| After new commit without scan, git check warns |
| MCP `doctor` returns same structured payload |

---

## v1.4.0 bundle

| Item | Surface |
| --- | --- |
| Doctor | `reponerve doctor`, MCP `doctor` |
