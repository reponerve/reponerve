# RepoNerve Greenfield Guide

Version: 1.0

Status: Draft

Updated: 2026-06-11

Related:

* `docs/vision/project-charter.md`
* `docs/product/token-economics.md`
* `docs/product/use-cases.md`

---

# Purpose

Clarify how RepoNerve applies when building a repository from scratch — from idea through first commits and beyond.

RepoNerve is **not** an autonomous coding agent and does **not** turn an idea into a shipped product by itself. See `docs/vision/project-charter.md` (out of scope: autonomous coding agent).

RepoNerve **is** valuable on greenfield projects when used as **memory from day one** — so understanding never becomes archaeology.

---

# Lifecycle Phases

```text
Phase 0: Idea only           →  RepoNerve alone: minimal value (nothing to ingest)
Phase 1: First commit        →  Value begins (if decisions are captured)
Phase 2: Growing codebase    →  High value (plan, impact, explain compound)
Phase 3: Mature repository   →  Maximum value (full memory, linking, onboarding)
```

Brownfield maximizes value by **excavating** lost context. Greenfield maximizes value by **never losing** context.

---

# Recommended Workflow

```text
1. Idea           → Human/agent: scope, stack, sketch (any build tool)
2. reponerve init → Local memory workspace
3. ADR-0001..N    → Record decisions WITH or BEFORE first code
4. Scaffold       → Agent generates structure (Cursor, Claude Code, etc.)
5. reponerve scan → After each meaningful milestone
6. reponerve mcp  → Agent builds with persistent understanding
7. plan/impact/review → Per feature, not per incident
```

---

# Division of Labor

| Responsibility | Tool |
| --- | --- |
| Generate code from idea | LLM + IDE agent (Cursor, Claude Code, Copilot) |
| Preserve why, who, what breaks | RepoNerve |
| Compress shell noise | RTK (optional, adjacent) |
| Cross-session chat memory | Supermemory/ICM (optional, adjacent) |

RepoNerve is the **brain that makes greenfield projects stay intelligible** — not the layer that invents them.

---

# Greenfield vs Brownfield

| | Brownfield | Greenfield |
| --- | --- | --- |
| Main pain | Nobody remembers why | "We'll document later" |
| RepoNerve role | Excavate and preserve | Capture from birth |
| Onboarding | New devs catch up fast | Founder does not become the bottleneck |
| Agent sessions | Stop re-learning legacy | Never build amnesia into the project |
| Best practice | Scan after clone | `init` + ADRs + scan from commit 1 |

---

# Commands Most Useful Early

| Need | Command | LLM tokens |
| --- | --- | --- |
| Workspace setup | `reponerve init` | 0 |
| Capture history | `reponerve scan` | 0 |
| Repo overview | `reponerve context generate` | 0 |
| Feature guidance | `reponerve plan "Add OAuth"` | 0 for analysis (v1.0) |
| Agent integration | `reponerve mcp` | Per-query, bounded |

---

# What Not to Promise

* RepoNerve will not replace build-from-prompt tools for Phase 0 (idea only)
* RepoNerve will not autonomously implement from a napkin sketch
* Value requires something to preserve: commits, ADRs, or code

---

# v1.0 Enhancements

Delivered as part of v1.0 via `docs/roadmap/v1.0-iteration-plan.md`:

* `reponerve remember` / `reponerve forget` — ISSUE-061
* Incremental scan on commit via hooks — ISSUE-060
* `reponerve hook install` — ISSUE-060

---

# One Sentence

RepoNerve will not build your repository from an idea; it will ensure a repository built from an idea **stays understandable, cheap to extend with AI, and survivable after contributors move on**.
