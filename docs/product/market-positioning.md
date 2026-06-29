# RepoNerve Market Positioning

Version: 1.1

Status: Current

Updated: 2026-06-25

Related:

* `docs/product/token-economics.md`
* `docs/product/implementation-status.md`
* `docs/architecture/agent-native-repository-intelligence.md`
* `docs/product/competitive-landscape-2026-06-29.md`

---

# Category

RepoNerve owns **Software Understanding Infrastructure** — not generic agent memory, not code search alone, not autonomous coding.

Positioning statement:

> **RepoNerve is local-first software understanding — the only system that tells agents and developers not just where code is, but why it exists, who owns it, and what breaks if you change it — with evidence on every answer.**

Users do not buy intelligence. They buy understanding, development speed, confidence, and reduced exploration.

---

# Competitive Landscape (2026)

Deep-dive issue backlog and source notes: `docs/product/competitive-landscape-2026-06-29.md`.

## Tier 1: Code Graph + MCP (closest competitors)

| Tool | Strength | RepoNerve differentiation (v1.5.1) |
| --- | --- | --- |
| [GitNexus](https://github.com/abhigyanpatwari/GitNexus) | Local code KG, web UI, 16 MCP tools | RepoNerve: repository memory (why), ownership, ADR→code linking, 49 MCP tools, native discipline on init |
| [Cortex](https://github.com/DanielBlomma/cortex/) | Code + ADRs + rules, one-command bootstrap | RepoNerve: evidence-mandatory conclusions, repo-adaptive `discipline-policy.json`, no optional vectors required |
| [Code-Nexus](https://github.com/snagrecha/code-nexus) | KG + git temporal overlay | RepoNerve: deterministic graph + session memory tied to repo entities, not chat logs |
| [Codebase-Memory](https://arxiv.org/html/2603.27277v1) | 66 languages, published token benchmarks | RepoNerve: smaller language set (20) but deeper why/who/impact layer and ship readiness |
| CodeGraphContext | Graph DB + MCP, MIT | RepoNerve: SQLite local-first, ownership + decisions, team PR workflow (`pr-context`) |

**Their model:** `Code → Graph → MCP → Agent`

**Their gap:** Little repository memory (why), ownership intelligence, or evidence-mandatory linking from decisions to code.

## Tier 2: Enterprise Platform

| Tool | Strength | RepoNerve differentiation |
| --- | --- | --- |
| [Sourcegraph Cody](https://sourcegraph.com/docs/cody) | SCIP code graph, cross-repo at enterprise scale | Local-first, evidence on conclusions, repository memory depth |
| Greptile, DeepWiki | Fast repo-wide AI Q&A (SaaS) | Deterministic, offline, no cloud dependency |

## Tier 3: Agent Memory (adjacent)

| Tool | Strength | RepoNerve differentiation |
| --- | --- | --- |
| Supermemory, ICM, Mem0 | Cross-session conversation memory, user profiles | RepoNerve remembers the **repository**, not chats |
| RTK | Shell output compression (60–90% token savings) | Composes with RepoNerve: RTK compresses shell; RepoNerve compresses understanding |
| Knowledge-graph tools (generic) | Docs/code → living KG, communities, audit trail | RepoNerve is deterministic-first; optional inference only with evidence tags |

## Tier 4: Context Packing (weak competition)

Repomix, code2prompt — dump repo into prompts. No structure, no memory, no graph. Not direct competition.

---

# Comparison Matrix

```text
                    Code structure   Why it exists   Who owns it   Evidence    Local-first   Dev discipline
GitNexus/Cortex     ████████████     ██░░░░░░░░░░    █░░░░░░░░░░░  █░░░░░░░░░  ████████████  ██░░░░░░░░░░
Sourcegraph Cody    ████████████     ████░░░░░░░░    ██░░░░░░░░░░  ████░░░░░░  ░░░░░░░░░░░░  ░░░░░░░░░░░░
Supermemory/ICM     ████░░░░░░░░     ██████░░░░░░    ░░░░░░░░░░░░  ██░░░░░░░░  ████████░░░░  ████░░░░░░░░
KG discovery tools   ██████░░░░░░     ████████░░░░    ░░░░░░░░░░░░  ████████░░  ████████████  ██░░░░░░░░░░
RepoNerve (v1.5.1)  ████████████     ████████████    ████████████  ████████████ ████████████  ████████████
```

**Dev discipline** = evidence-backed reuse, ship readiness, and review habits bundled on `reponerve init` (RFC-003/004) — not generic agent prompt packs.

---

# What RepoNerve Must Not Claim

* Semantic or embedding search as primary authority (see `docs/roadmap/v1.x-backlog.md`)
* Replacement for Cursor, Copilot, or autonomous build-from-idea tools
* Enterprise polyrepo federation at Sourcegraph scale
* Full graph-explorer product beyond the capped local `reponerve explore` UI (v1.5.0)

---

# Audience Pitches

## Individual Developer + AI Agent

**Pitch:** Stop re-exploring your repo every session. Point your agent at `reponerve mcp`.

**Entry:** MCP install, 60-second demo on an OSS repo.

## Team / Tech Lead

**Pitch:** Decisions do not die when people leave. ADRs link to code. PRs get impact and ownership context.

**Entry:** `reponerve review`, `reponerve ship-check`, `reponerve pr-context`, scan on merge.

## Enterprise / Regulated

**Pitch:** Evidence-mandatory AI reasoning. Every conclusion traceable to commit, ADR, or symbol.

**Entry:** Audit trail, deterministic outputs, local-first deployment.

---

# Go-to-Market Sequence

1. **MCP-first** — developers discover via MCP registries and agent configs
2. **OSS demo** — `reponerve scan` on Gin/Hugo/Kubernetes; show `explain` with ADR + code + owner
3. **Compose with RTK** — shell noise + understanding compression together
4. **PR workflow template** — `reponerve integrate` ships `.github/workflows/reponerve-pr.yml.example`; teams wire `pr-context` into CI
5. **Manifesto** — *Evidence-Free Conclusions Are Invalid* for quality-conscious teams
6. **npm + Homebrew** — frictionless install (`v1.3.2`+, `v1.4.0`+)

---

# Moat

RepoNerve wins when the question is:

* Why does this exist?
* Who owns it?
* What decision shaped it?
* What breaks if I change it?
* Can I trust this answer?

Code-graph tools win when the question is only *where is this symbol defined?*

RepoNerve competes on structure **and** the why layer: 20 languages (Tree-sitter), repository memory, ownership, and native development discipline on every `init`.

---

# Native Development Discipline (differentiator)

External agent discipline packs (reuse-first ladders, pre-ship review personas) duplicate setup and ignore scanned repository evidence. RepoNerve ships **Native Development Discipline** by default:

| Habit | RepoNerve surface |
| --- | --- |
| Reuse before write | `reuse-check`, `plan` starting points |
| Pre-ship validation | `ship-check` → `ship_blockers` / `advisories` |
| Evidence review | `review` + `discipline_checks` from repo policy |
| Team PR workflow | `pr-context`, bundled workflow template |

Spec: `docs/rfc/RFC-003-native-development-discipline.md`. Optional narrative multi-persona review (`docs/council/`) is **not** bundled on init — structured evidence replaces LLM roleplay.

---

# Growth Risks

| Risk | Mitigation |
| --- | --- |
| Crowded MCP code-graph space | Own evidence + repository memory + discipline-on-init niche |
| IDEs add native code intelligence | Moat = repository memory + linking + ship readiness, not grep |
| Agents ignore rules | Supply evidence via MCP/CLI; discipline rules are lite (~80 lines) |
| Visual UI expectations | `reponerve explore` (v1.5.0) for local browse; full explorer out of scope |

See `docs/product/implementation-status.md` and `docs/releases/versioning.md` for current release line (**v1.5.1**).
