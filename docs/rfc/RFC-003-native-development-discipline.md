# RFC-003: Native Development Discipline

Status: Accepted (Phase A–C baseline)  
Date: 2026-06-24

Related:

* `docs/releases/versioning.md`
* `docs/architecture/agent-context-contract.md`
* RFC-001 bounded responses, RFC-002 feature intelligence

---

## Problem

Teams install RepoNerve for repository understanding, then add separate agent rulesets for:

- Reuse-before-write discipline
- Pre-ship review and risk surfacing
- Surgical change habits

That duplicates setup, splits authority (prompt vs evidence), and leaves generic rules that ignore the scanned repository.

## Decision

RepoNerve ships **Native Development Discipline** by default on `reponerve init` — evidence-backed development habits bundled with the skill and MCP, no third-party rules required.

### Naming (RepoNerve-native)

| Concept | Name | Surface |
| --- | --- | --- |
| Umbrella | **Development Discipline** | Bundled Cursor rules + agent envelope guidance |
| Reuse existing code first | **Reuse Protocol** | `reuse_check` CLI/MCP (Phase B) |
| Pre-merge / pre-ship validation | **Ship Readiness** | `ship_check` CLI/MCP (Phase C) |
| Structured critique of a change | **Evidence Review** | Enhanced `review` output + discipline packs |

Do not reference external discipline products in docs, rules, or code comments.

---

## Phase A — Bundled on init (this RFC)

`reponerve init` installs:

| File | Purpose |
| --- | --- |
| `.cursor/rules/reponerve.mdc` | Context-first: load RepoNerve before grep |
| `.cursor/rules/coding-guidelines.mdc` | Surgical changes, simplicity, verifiable goals |
| `.cursor/rules/development-discipline.mdc` | When to run `plan`, `review`, `impact`; skip discipline on explain/ask |

Rules are **lite** (~80 lines total discipline). Full multi-perspective review remains optional team documentation, not required per repo.

### Discipline workflow (agent)

```text
New feature / epic     → reponerve plan "<task>" --json
Before writing code    → check structured.reuse_candidates (Phase B) or plan starting_points
Before merge / ship    → reponerve review "<topic>" --json  (+ ship_check Phase C)
Ambiguous symbol       → explain_function / explain_file with --package
```

Skip discipline framing for informational `ask`, `explain-*`, and narrow verification — same as RepoNerve skill.

---

## Phase B — Reuse Protocol (shipped)

New command and MCP tool: `reuse_check` / `reponerve reuse-check`

Deterministic output:

- Existing symbols, files, and ADRs relevant to a stated intent
- Ranked by graph proximity and repository search evidence
- `recommended_next_tools`: `explain_function`, `explain_file`, `plan`

Replaces generic "search the codebase first" prompts with structured reuse candidates.

---

## Phase C — Ship Readiness (shipped)

New command and MCP tool: `ship_check` / `reponerve ship-check`

Input: optional topic or inferred diff scope.

Output:

- Impacted entities and ADRs
- Test / migration / rollback reminders from repository evidence
- Reviewer recommendations from ownership
- `ship_blockers` vs `advisories` (structured, not narrative council roleplay)

---

## Phase D — Repo-adaptive policy (future)

After `scan`, generate `.reponerve/discipline-policy.json`:

- Detected ADR directory → offer ADR on architecture changes
- CI workflow files → link ship checks to pipeline
- Dominant language / layout → layer conventions in agent envelope

---

## Non-goals

- LLM roleplay of fixed persona panels on every query
- Replacing the IDE agent runtime
- Hard enforcement (agents may ignore rules; RepoNerve supplies evidence)

---

## Success criteria

| Phase | Verify |
| --- | --- |
| A | Fresh `reponerve init` writes three rules; skill references discipline workflow |
| B | `reuse_check "add OAuth"` returns symbol candidates with `defined_in` and evidence |
| C | `ship_check` returns structured blockers for a scoped change topic |
