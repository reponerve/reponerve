# RepoNerve Vision

Version: 1.3

Status: Draft

Updated: 2026-06-11

Related: `tasks/ARCH-001.md`, `docs/architecture/architecture-overview.md`

---

## Vision

RepoNerve reduces the cost of software understanding.

RepoNerve is a software understanding platform built around knowledge preservation. Its purpose is to ensure that software understanding survives beyond individual contributors and remains accessible to both humans and AI systems.

The primary goal is to minimize the time, effort, and token consumption required for humans and AI systems to understand and evolve software.

RepoNerve is the intelligence layer for software understanding — not merely a repository intelligence platform.

Users do not buy intelligence. They buy understanding, development speed, confidence, and reduced exploration.

---

## Product Mission

RepoNerve preserves, organizes, and transfers software knowledge so that understanding survives beyond individual contributors and remains accessible to both humans and AI systems.

RepoNerve helps humans and AI systems understand:

- What a repository does
- What features exist
- How the code works
- Why the code exists
- Who built it
- Who owns it
- What depends on it
- What changes are required
- How the software should evolve

without repeated repository exploration.

---

## Core Problem

Software remembers code.

Software forgets context.

Teams lose knowledge through:

- Employee turnover
- Project evolution
- Documentation drift
- Architectural changes
- Team growth
- AI context limitations

Developers and AI agents repeatedly spend time and tokens rediscovering information that already existed somewhere in the repository or in people's heads.

RepoNerve exists to prevent this knowledge loss.

---

## Software Understanding Model

Software Understanding is the primary product outcome.

```text
Code Understanding
    +
Repository Understanding
    +
Ownership Understanding
    +
Architectural Understanding
    +
Change Understanding
    +
Historical Understanding
    ═══════════════════════
    Software Understanding
```

RepoNerve delivers Software Understanding through its intelligence layers and Development Experience.

---

## Knowledge Preservation

Software knowledge must survive:

- Team changes
- Contributor turnover
- Architectural evolution
- Long-lived repositories

RepoNerve acts as a **software memory system** — the organizational outcome of knowledge preservation.

```text
Knowledge Preservation
        ↓
Software Memory
        ↓
Software Understanding
```

Repository Intelligence captures decisions, facts, events, ownership, and relationships from repository artifacts.

Code Intelligence captures structure, symbols, and dependencies from source code.

The goal is that understanding remains available even when original authors are no longer present.

---

## Knowledge Transfer

RepoNerve helps knowledge move without requiring direct access to original contributors.

RepoNerve should help:

- New developers onboard faster
- Existing developers understand unfamiliar systems
- Reviewers gain context before reviewing changes
- Architects understand impact before proposing changes
- Engineering managers understand ownership and expertise
- AI coding agents obtain development context before implementation

Knowledge should be transferable through Development Experience — not through tribal knowledge or repeated file exploration.

---

## Repository Understanding

Repository understanding is a first-class outcome.

RepoNerve should answer:

- What does this repository do?
- What business capabilities exist?
- What features exist?
- What domains exist?
- How is the repository organized?
- What are the important components?

Repository Intelligence provides the foundation. Development Experience surfaces repository understanding through `ask`, `explain`, and related workflows.

---

## Feature Understanding

Humans think in features, not files. Feature Understanding is a **first-class v1.0 goal** — not a secondary capability.

RepoNerve should evolve toward understanding features as first-class concepts:

```text
Feature
    ↓
Code
    ↓
Ownership
    ↓
Decisions
    ↓
Impact
```

Examples: Authentication, Billing, Metadata Management, Notifications, Search.

Feature understanding is a **v1.0** capability. Humans think in features; RepoNerve v1.0 must support feature-level questions through Development Experience (e.g. `reponerve explain "authentication"` resolving Feature → Code → Ownership → Decisions → Impact).

Delivered as part of ISSUE-057 and the v1.0 release. There is no separate post-v1.0 product release for this capability.

---

## v1.0 Scope

RepoNerve has **one product release: v1.0.0**. All product capabilities ship together. There is no v1.x product roadmap.

Capabilities are delivered through **v0.x.0-alpha engineering iterations** (see `docs/roadmap/v1.0-iteration-plan.md`).

v1.0 includes:

**Foundation (complete)**

- Knowledge Preservation
- Repository Intelligence
- MCP server and agent services

**ISSUE-057 — Code Intelligence & Development Experience**

- Code Intelligence (Go AST, symbols, call graph)
- Repository-Code linking
- Feature Understanding
- Development Experience (`ask`, `explain`, `explain-file`, `explain-function`, `explain-struct`, `explain-interface`, `explain-type`, `plan`, `impact`, `review`)

**ISSUE-059 — Foundation Fixes**

- Expertise in scan pipeline, CLI exposure, ingestion debt fixes

**ISSUE-060 — Token Intelligence**

- Graph-aware compression, token budgets
- `--format caveman|prose|json`
- `reponerve hook install`, incremental scan

**ISSUE-061 — Evidence Graph & Session Memory**

- Graph communities, surprising connections, `reponerve explore`
- `remember` / `forget`, session writeback, agent handoff bundles
- Fixed workflow templates

**ISSUE-062 — Multi-Language Code Intelligence**

- Tree-sitter: TypeScript, Python, Rust

**Outcomes**

- Software Understanding (complete product outcome)
- Knowledge Transfer
- Token-efficient MCP and CLI delivery

RepoNerve v1.0.0 is **not tagged** until all issues above are complete and acceptance criteria pass.

### Explicit v1.0 Non-Goals

The following are **agreed out of scope** for v1.0.0. They require a post-release RFC to reconsider:

- Semantic or hybrid embedding search
- User-defined workflow composition
- Autonomous code modification or deployment
- Cloud-required core product (mandatory SaaS)
- Cross-repo enterprise federation

See `docs/roadmap/v1.x-backlog.md`.

---

## AI Agent Context

RepoNerve reduces the amount of repository exploration required by AI systems.

The objective is not simply retrieval. The objective is understanding.

AI agents should obtain before beginning implementation work:

- Code context
- Repository context
- Ownership context
- Architectural context
- Change context

Development Experience and MCP expose this context in token-efficient, evidence-backed packages — reducing what agents must rediscover.

See `docs/product/token-economics.md` for the full token cost model.

---

## Token Economics

RepoNerve reduces the **cost of software understanding** as LLMs become more expensive.

The primary waste in AI-assisted development is not generation — it is **re-exploration**: agents re-read files, re-grep, and re-summarize every session.

RepoNerve moves understanding out of the token meter:

- **Scan once** (deterministic, zero LLM tokens)
- **Query cheaply** (bounded MCP context packs)
- **Persist across sessions** (repository memory, not chat memory)

Premium models should spend tokens on building and deciding — not on re-learning what the repository already knew.

---

## Market Position

RepoNerve is **Software Understanding Infrastructure**.

Many tools provide `Code Graph → Retrieval → LLM Context`. RepoNerve provides preserved repository knowledge, code intelligence, repository-code linking, and feature understanding — with evidence on every conclusion.

RepoNerve composes with adjacent tools:

- **RTK** — compresses shell output; RepoNerve compresses understanding
- **Graphify-style graph discovery** — communities and surprises (ISSUE-061, v1.0)
- **Agent memory tools** — remember conversations; RepoNerve remembers the repository

See `docs/product/market-positioning.md`.

---

## Greenfield and Brownfield

RepoNerve is not an autonomous coding agent. It does not build a repository from an idea.

It ensures repositories — whether legacy or greenfield — **preserve understanding from early commits**: ADRs, scan, MCP, and Development Experience compound from day one on new projects.

See `docs/product/greenfield-guide.md`.

---

## Delivery Stack (Super Intelligence Layers)

RepoNerve delivers understanding through layered capabilities:

```text
INGEST → INDEX → LINK → RECALL → COMPRESS → DELIVER → LEARN

Ingest     Git, ADR, code (ISSUE-057)
Index      Memory + code store + graph
Link       Repository-code relationships
Recall     MCP + Development Experience
Compress   Context packs, token budgets (ISSUE-060)
Deliver    CLI, MCP, structured formats (ISSUE-060)
Learn      Session writeback, remember/forget (ISSUE-061)
```

Iteration plan: `docs/roadmap/v1.0-iteration-plan.md`.

---

## Implementation Status

Repository Intelligence is complete. Code Intelligence and Development Experience (ISSUE-057) are the v1.0 blocker.

Honest code-vs-documentation snapshot: `docs/product/implementation-status.md`.

---

## Architectural Pillars

RepoNerve v1.0 is built on explicit architectural pillars. See `docs/architecture/architecture-overview.md` for full subsystem definitions.

```text
Knowledge Preservation          (Core Platform Capability)
    ↓
Software Memory
    ↓
Repository Intelligence         (why)
    +
Code Intelligence               (how)
    ↓
Repository-Code Linking           (cross-authority)
    ↓
Feature Understanding            (what — feature-level)
    ↓
Development Experience          (product surface)
    ↓
Software Understanding          (outcome)
```

The **Understanding Engine** retrieves and assembles context across all intelligence sources. It evolved from the earlier Query Engine as the platform grew beyond repository memory.

| Pillar | Role | v1.0 Status |
| --- | --- | --- |
| Knowledge Preservation | Core platform foundation — all layers depend on it | Core Platform Capability |
| Repository Intelligence | Why software exists — decisions, facts, events, ownership | Complete |
| Code Intelligence | How code works — symbols, graphs, dependencies | ISSUE-057 |
| Repository-Code Linking | Connect repository entities to code entities | ISSUE-057 |
| Feature Understanding | Feature → Code → Ownership → Decisions → Impact | ISSUE-057 |
| Development Experience | ask, explain, explain-file, explain-function, explain-struct, explain-interface, explain-type, plan, impact, review | ISSUE-057 |
| Software Understanding | Complete product outcome | Blocked |

---

## Product Layers

```text
Knowledge Preservation          (Core Platform Capability)
    ↓
Software Memory
    ↓
Repository Intelligence
    +
Code Intelligence
    ↓
Repository-Code Linking
    ↓
Feature Understanding
    ↓
Development Experience
    ↓
Software Understanding
```

Repository Intelligence, Code Intelligence, and Development Experience are capabilities.

Software Understanding is the outcome.

### Knowledge Preservation

Core platform foundation — capture and retain software knowledge before it is lost. All intelligence layers depend on it.

Status: Core Platform Capability.

### Software Memory

Durable organizational knowledge that survives contributor turnover and architectural evolution. The outcome of Knowledge Preservation.

### Repository Intelligence

Answers why repository knowledge exists. One capability — not the whole product.

Status: Implemented.

### Code Intelligence

Answers how code works. One capability — not the whole product.

Status: Required for v1.0 — ISSUE-057.

### Repository-Code Linking

Connects repository entities (decisions, facts, events) to code entities (files, symbols). Required for unified explain output.

Status: Required for v1.0 — ISSUE-057.

### Feature Understanding

First-class v1.0 goal. Humans think in features, not files.

```text
Feature → Code → Ownership → Decisions → Impact
```

Status: Required for v1.0 — ISSUE-057.

### Development Experience

The primary user-facing layer. How humans and AI consume RepoNerve.

```bash
reponerve ask "Who owns billing?"
reponerve explain "authentication"
reponerve explain-file "internal/auth/oauth.go"
reponerve plan "Add OAuth login"
reponerve impact user-service
reponerve review "metadata panel"
```

Status: Required for v1.0 — ISSUE-057.

### Knowledge Transfer

Mission outcome delivered through Development Experience. See `docs/roadmap/v1.0-prd.md` Goal 6.

---

## Differentiation

Most tools focus on:

```text
Code Graph
    ↓
Retrieval
    ↓
LLM Context
```

RepoNerve focuses on:

```text
Knowledge Preservation
    ↓
Software Memory
    ↓
Code Intelligence
    +
Repository Intelligence
    +
Repository-Code Linking
    +
Feature Understanding
    ↓
Software Understanding
```

This enables software understanding rather than simple code retrieval.

RepoNerve does not replace LLMs. It preserves and transfers knowledge so agents and developers understand software before they act.

---

## Guiding Principle

Understanding first.

Evidence second.

AI third.

RepoNerve prioritizes deterministic understanding and evidence-backed intelligence before introducing AI-assisted reasoning.

---

## What Users Get

| Outcome | How RepoNerve Delivers It |
| --- | --- |
| Software understanding | Combined code, repository, and ownership context |
| Knowledge preservation | Durable memory across team changes |
| Knowledge transfer | Onboarding, review, and agent context without tribal knowledge |
| Development speed | Guided starting points through `plan` |
| Confidence | Evidence-backed answers through `ask` |
| Reduced exploration | Pre-indexed knowledge and development workflows |
| Lower AI token cost | Deterministic scan + bounded MCP context packs |
| Premium models within limits | Understanding delivered before LLM reasoning begins |

---

## Release State

RepoNerve has one product release target: **v1.0**. Everything in the product vision ships in v1.0.

| Capability | Role | v1.0 Status |
| --- | --- | --- |
| Knowledge Preservation | Core platform foundation | Core Platform Capability |
| Repository Intelligence | Capability | Complete |
| Code Intelligence | Capability | ISSUE-057 |
| Repository-Code Linking | Cross-authority subsystem | ISSUE-057 |
| Feature Understanding | First-class v1.0 goal | ISSUE-057 |
| Knowledge Transfer | Mission outcome via Development Experience | ISSUE-057 |
| Development Experience | Product surface | ISSUE-057 |
| Software Understanding | Outcome | Blocked until ISSUE-057 complete |

**RepoNerve v1.0** is not released until ISSUE-057 is complete and all v1.0 scope items above are delivered.

There is no v1.1, v1.x, or partial product release.
