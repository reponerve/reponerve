# RepoNerve Versioning Policy

Version: 1.0

Status: Current

Updated: 2026-06-24

Related:

* `docs/roadmap/v1.0-iteration-plan.md` — pre-1.0 engineering checkpoints
* `docs/roadmap/v1.x-backlog.md` — capabilities out of v1.0 scope
* `docs/releases/v1.0.0.md` — first product release

---

# Summary

RepoNerve uses **two eras** of version tags:

| Era | Tag pattern | Purpose |
| --- | --- | --- |
| Pre-1.0 | `v0.x.0-alpha` | Engineering checkpoints while building toward v1.0 |
| Post-1.0 | Semver `vMAJOR.MINOR.PATCH` | Product releases after v1.0.0 |

**Do not create new `v0.x-alpha` tags after `v1.0.0`.**

---

# Current release line

| Tag | Date | Notes |
| --- | --- | --- |
| `v1.0.0` | 2026-06-18 | First product release — complete v1.0 scope |
| `v1.0.1` | 2026-06-19 | Patch — scan reliability on real repositories |

Latest tagged release: **`v1.0.1`**.

---

# Pre-1.0 (`v0.x-alpha`)

Before `v1.0.0`, alpha tags marked **engineering milestones** for contributors — not partial product releases.

```text
v0.10.0-alpha  →  Foundation fixes (ISSUE-059)
v0.11.0-alpha  →  Code Intelligence core
v0.12.x-alpha  →  Development Experience + linking (ISSUE-057)
v0.13.0-alpha  →  Token Intelligence (ISSUE-060) — documented; shipped in v1.0.0
v0.14.0-alpha  →  Evidence Graph + Session Memory (ISSUE-061) — documented; shipped in v1.0.0
v0.15.0-alpha  →  Multi-language indexing (ISSUE-062) — documented; shipped in v1.0.0
v1.0.0         →  Product release (all v1.0 scope)
```

**Git note:** The last alpha tag before `v1.0.0` is `v0.12.2-alpha`. Milestones `v0.13`–`v0.15` are recorded in release notes and task files; their work is included in `v1.0.0` without separate git tags. That is acceptable — alphas were checkpoints, not consumer-facing releases.

---

# Post-1.0 (semver)

After `v1.0.0`, follow [Semantic Versioning](https://semver.org/):

| Bump | When | Examples |
| --- | --- | --- |
| **PATCH** `v1.0.x` | Bug fixes, docs-only corrections, internal hardening — no new capabilities, no breaking MCP/CLI contract | `v1.0.1` scan fix |
| **MINOR** `v1.1.0` | New backward-compatible capabilities — new MCP tools, CLI commands, default behavior that agents can opt out of | Bounded responses (RFC-001), feature intelligence v2 (RFC-002) |
| **MAJOR** `v2.0.0` | Breaking changes to MCP tool schemas, JSON envelope contract, CLI flags, or memory format requiring migration | Requires RFC + migration notes |

### Rules

1. **RFC required** for MINOR and MAJOR scope (see `docs/roadmap/v1.x-backlog.md`).
2. **Release notes** in `docs/releases/vX.Y.Z.md` for every tag.
3. **No `-alpha` suffix** on post-1.0 product tags unless explicitly running a long beta program (not the default).
4. **Tag from `main`** after `go test ./...` passes and release checklist items for that version are complete.

### Suggested next releases

| Work | Suggested tag |
| --- | --- |
| Bounded agent responses + Feature Intelligence v2 + MCP tool additions | `v1.1.0` |
| Reuse Protocol (`reuse-check`) + Ship Readiness (`ship-check`) | `v1.2.0` (RFC-003 B/C) |
| Docs-only / council / stale-doc fixes without behavior change | Include in nearest release or `v1.0.2` |
| Breaking envelope or storage migration | `v2.0.0` (RFC) |

---

# What not to do

| Anti-pattern | Why |
| --- | --- |
| `v0.16-alpha` after v1.0.0 | Confuses consumers; v0 era is closed |
| "Pending tag" docs after tag exists | Drift — update status when `git tag` is created |
| MINOR bump without release notes | Agents and users depend on MCP contract stability |
| Claiming "Complete" without acceptance evidence | Version number does not substitute for capability honesty |

---

# Checklist for a new tag

1. RFC approved (MINOR/MAJOR) or trivial fix justified (PATCH)
2. `go test ./...` passes
3. `docs/releases/vX.Y.Z.md` written
4. Stale "latest version" references updated (`README.md`, `AGENTS.md`, `implementation-status.md`)
5. `git tag vX.Y.Z` on the release commit
6. GitHub release published from release notes (optional but recommended)

---

# Document maintenance

When creating a git tag, update these files if they mention release status:

* `README.md`
* `AGENTS.md`
* `docs/product/implementation-status.md`
* `docs/releases/versioning.md` (current release line table)

Historical audit documents (e.g. `docs/audits/v1.0-release-review.md`) may keep their original approval date; add a one-line note that the tag was created rather than rewriting the audit verdict.
