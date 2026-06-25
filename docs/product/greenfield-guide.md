# RepoNerve Greenfield Guide

Version: 1.1

Status: Current

Updated: 2026-06-25

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
2. reponerve init → Local memory workspace + skill + MCP + discipline rules
3. ADR-0001..N    → Record decisions WITH or BEFORE first code
4. Scaffold       → Agent generates structure (Cursor, Claude Code, etc.)
5. reponerve scan → After each meaningful milestone
6. reponerve mcp  → Agent builds with persistent understanding
7. plan → reuse-check → implement → ship-check / review / pr-context → per feature
```

---

# Division of Labor

| Responsibility | Tool |
| --- | --- |
| Generate code from idea | LLM + IDE agent (Cursor, Claude Code, Copilot) |
| Preserve why, who, what breaks | RepoNerve |
| Reuse-before-write + pre-ship habits | RepoNerve Native Development Discipline (on `init`) |
| Compress shell noise | RTK (optional, adjacent) |
| Cross-session chat memory | Supermemory/ICM (optional, adjacent) |

No separate Ponytail-style discipline skills or council persona packs are required — `reponerve init` bundles evidence-backed habits. Optional narrative council: `docs/council/software-development-council.md`.

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
| Feature guidance | `reponerve plan "Add OAuth"` | 0 for analysis |
| Reuse existing code | `reponerve reuse-check "add OAuth"` | 0 |
| Pre-ship validation | `reponerve ship-check "OAuth"` | 0 |
| Agent integration | `reponerve mcp` | Per-query, bounded |

---

# What Not to Promise

* RepoNerve will not replace build-from-prompt tools for Phase 0 (idea only)
* RepoNerve will not autonomously implement from a napkin sketch
* Value requires something to preserve: commits, ADRs, or code

---

# Shipped Enhancements

Delivered in v1.0 and post-1.0 releases (see `docs/product/implementation-status.md`):

* `reponerve remember` / `reponerve forget` — session memory
* Incremental scan on commit via hooks — `reponerve hook install`
* Native Development Discipline on `init` — reuse, ship readiness, review habits (RFC-003)
* Team PR workflow — `reponerve pr-context`, CI template (RFC-004)

---

# One Sentence

RepoNerve will not build your repository from an idea; it will ensure a repository built from an idea **stays understandable, cheap to extend with AI, and survivable after contributors move on**.
