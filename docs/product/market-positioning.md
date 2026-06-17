# RepoNerve Market Positioning

Version: 1.0

Status: Draft

Updated: 2026-06-11

Related:

* `docs/product/token-economics.md`
* `docs/product/implementation-status.md`
* `docs/architecture/agent-native-repository-intelligence.md`

---

# Category

RepoNerve owns **Software Understanding Infrastructure** — not generic agent memory, not code search alone, not autonomous coding.

Positioning statement:

> **RepoNerve is local-first software understanding — the only system that tells agents and developers not just where code is, but why it exists, who owns it, and what breaks if you change it — with evidence on every answer.**

Users do not buy intelligence. They buy understanding, development speed, confidence, and reduced exploration.

---

# Competitive Landscape (2026)

## Tier 1: Code Graph + MCP (closest competitors)

| Tool | Strength | What they have that RepoNerve lacks (today) |
| --- | --- | --- |
| [GitNexus](https://github.com/abhigyanpatwari/GitNexus) | Local code KG, web UI, 16 MCP tools | Multi-language Tree-sitter, visual explorer, large adoption |
| [Cortex](https://github.com/DanielBlomma/cortex/) | Code + ADRs + rules, one-command bootstrap | Git hooks, optional vectors, rule enforcement at retrieval |
| [Code-Nexus](https://github.com/snagrecha/code-nexus) | KG + git temporal overlay | Time-travel graph, 3D viz, context pruning, session persistence |
| [Codebase-Memory](https://arxiv.org/html/2603.27277v1) | 66 languages, published token benchmarks | Community detection, incremental index at scale |
| CodeGraphContext | Graph DB + MCP, MIT | Permissive license, high PyPI adoption |

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
                    Code structure   Why it exists   Who owns it   Evidence    Local-first
GitNexus/Cortex     ████████████     ██░░░░░░░░░░    █░░░░░░░░░░░  █░░░░░░░░░  ████████████
Sourcegraph Cody    ████████████     ████░░░░░░░░    ██░░░░░░░░░░  ████░░░░░░  ░░░░░░░░░░░░
Supermemory/ICM     ████░░░░░░░░     ██████░░░░░░    ░░░░░░░░░░░░  ██░░░░░░░░  ████████░░░░
KG discovery tools   ██████░░░░░░     ████████░░░░    ░░░░░░░░░░░░  ████████░░  ████████████
RepoNerve (today)   ██░░░░░░░░░░     ████████████    ████████░░░░  ████████████ ████████████
RepoNerve (v1.0)    ████████████     ████████████    ████████████  ████████████ ████████████
```

---

# What RepoNerve Must Not Claim at Launch

* Multi-language support beyond Go (v1.0 initial scope)
* Semantic or embedding search (out of v1.0 — see `docs/roadmap/v1.x-backlog.md`)
* Replacement for Cursor, Copilot, or autonomous build-from-idea tools
* Enterprise polyrepo scale comparable to Sourcegraph

---

# Audience Pitches

## Individual Developer + AI Agent

**Pitch:** Stop re-exploring your repo every session. Point your agent at `reponerve mcp`.

**Entry:** MCP install, 60-second demo on an OSS repo.

## Team / Tech Lead

**Pitch:** Decisions do not die when people leave. ADRs link to code. PRs get impact and ownership context.

**Entry:** `reponerve explain`, `reponerve review`, scan on merge.

## Enterprise / Regulated

**Pitch:** Evidence-mandatory AI reasoning. Every conclusion traceable to commit, ADR, or symbol.

**Entry:** Audit trail, deterministic outputs, local-first deployment.

---

# Go-to-Market Sequence

1. **MCP-first** — developers discover via MCP registries and agent configs
2. **OSS demo** — `reponerve scan` on Gin/Hugo/Kubernetes; show `explain` with ADR + code + owner
3. **Compose with RTK** — shell noise + understanding compression together
4. **GitHub Action** — scan on merge; PR comments with impact and linked decisions
5. **Manifesto** — *Evidence-Free Conclusions Are Invalid* for quality-conscious teams

---

# Moat

RepoNerve wins when the question is:

* Why does this exist?
* Who owns it?
* What decision shaped it?
* What breaks if I change it?
* Can I trust this answer?

Code-graph tools win when the question is only *where is this symbol defined?*

RepoNerve must ship Code Intelligence (ISSUE-057) to compete on structure **and** own the why layer.

---

# Growth Risks

| Risk | Mitigation |
| --- | --- |
| Crowded MCP code-graph space | Own evidence + repository memory niche |
| IDEs add native code intelligence | Moat = repository memory + linking, not grep |
| v1.0 slips while competitors ship | Ship ISSUE-057; document honest implementation status |
| Limited visual UI | `reponerve explore` HTML in ISSUE-061 (v1.0) |

See `docs/roadmap/v1.0-iteration-plan.md` for delivery path.
