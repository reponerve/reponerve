# RFC-004: Team Delivery Intelligence

Status: Accepted (Phase A–C)  
Date: 2026-06-24

Related:

* `docs/rfc/RFC-003-native-development-discipline.md`
* `docs/releases/v1.3.0.md`
* `docs/product/market-positioning.md` — PR / CI GTM wedge

---

## Problem

Individual developers get value from RepoNerve MCP in the IDE. **Teams** adopt through pull requests and CI — where competitors post graph context but not **why**, **who owns it**, or **ship blockers** tied to ADRs and discipline policy.

RFC-003 shipped Reuse Protocol, Ship Readiness, and repo-adaptive `discipline-policy.json`, but:

- `review` does not surface discipline packs (policy hints, ADR expectations, next tools).
- There is no PR-scoped command to assemble review + ship evidence for changed files.
- GTM lists a GitHub Action; nothing is bundled for consumers.

## Decision

Ship **Team Delivery Intelligence** in **v1.3.0** — evidence-native PR workflow without cloud SaaS.

### Naming (RepoNerve-native)

| Concept | Name | Surface |
| --- | --- | --- |
| Umbrella | **Team Delivery Intelligence** | RFC-004 |
| Structured critique with policy | **Evidence Review** | Enhanced `review` output (Phase A) |
| PR-scoped evidence pack | **PR Context** | `pr-context` CLI/MCP (Phase B) |
| CI install path | **PR workflow template** | Bundled on `integrate` (Phase C) |

---

## Phase A — Evidence Review (shipped)

Enhance `review` / `reponerve review` structured output:

| Field | Source |
| --- | --- |
| `discipline_checks` | `.reponerve/discipline-policy.json` — ship hints, ADR directory, layer conventions |
| `recommended_next_tools` | `ship_check`, `reuse_check`, `analyze_topic_impact` |

Agent envelope (`kind: review`):

- Guidance to apply `discipline_checks` before merge
- `recommended_next_tools` when review scope is incomplete

Completes RFC-003 **Evidence Review** row without LLM roleplay.

---

## Phase B — PR Context (shipped)

New command and MCP tool: `pr-context` / `reponerve pr-context`

**Input:**

- `changed_files` — paths from `git diff --name-only` (CLI: positional args or `--file`)
- optional `topic` — overrides file-derived topic

**Output (`PRContextResult`):**

- `topic`, `changed_files`
- nested `review` (Evidence Review guide)
- nested `ship_check` (blockers + advisories)
- `pr_comment_markdown` — bounded markdown for GitHub PR comments
- standard `evidence`, `source_services`, agent envelope (`kind: pr_context`)

Topic derivation is deterministic: dominant path segment under `internal/` or top-level package directory.

---

## Phase C — GitHub Action template (shipped)

`reponerve integrate` installs:

| Bundle | Target |
| --- | --- |
| `github-workflow-reponerve-pr.yml` | `.github/workflows/reponerve-pr.yml.example` |

Users copy or rename to enable. Workflow:

1. Checkout with history
2. Install RepoNerve (`go install` or pinned release)
3. `reponerve init` + `reponerve scan` (if `.reponerve/` missing)
4. `reponerve pr-context --json` on changed files vs base branch
5. Post `pr_comment_markdown` via `actions/github-script`

No hosted RepoNerve service. All evidence stays in the runner workspace.

---

## Non-goals

- Hosted PR bot SaaS
- Semantic diff summarization
- Auto-approve / auto-merge
- `reponerve doctor` (freshness guard) — candidate for RFC-005
- Scoped monorepo scan — candidate for RFC-005

---

## Success criteria

| Phase | Verify |
| --- | --- |
| A | `review "auth"` returns `discipline_checks` when policy exists; agent kind `review` includes guidance |
| B | `pr-context internal/agent/foo.go --json` returns review + ship_check + `pr_comment_markdown` |
| C | `reponerve integrate` writes `reponerve-pr.yml.example`; doc describes enable steps |

---

## v1.3.0 bundle

| RFC | Capability |
| --- | --- |
| RFC-003 Phase D | `discipline-policy.json` on scan |
| RFC-004 Phase A | Evidence Review |
| RFC-004 Phase B | PR Context |
| RFC-004 Phase C | GitHub workflow template |
