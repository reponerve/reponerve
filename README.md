# RepoNerve

> The intelligence layer for software understanding.

RepoNerve preserves, organizes, and transfers software knowledge so that understanding survives beyond individual contributors and remains accessible to both humans and AI systems.

Software remembers code.

Software forgets context.

RepoNerve prevents knowledge loss and reduces the cost of software understanding.

---

# Vision

RepoNerve reduces the time, effort, and token consumption required for humans and AI systems to understand and evolve software.

Every repository should be self-explaining — how it works, why it exists, who owns it, and how to change it safely.

---

# What Problems Does RepoNerve Solve?

RepoNerve answers the questions developers and AI agents ask every day:

* **Explain this code** — Which files, packages, functions, and APIs are involved? What is the call graph?
* **Explain this feature** — Why does it exist? What decisions shaped it?
* **Who owns this area?** — Who created it? Who has expertise?
* **What breaks if I change this?** — What depends on it? What is impacted?
* **Who should review this?** — Which reviewers have the required expertise?
* **Add OAuth login** — What areas are affected? Where should I start?

Examples:

```bash
reponerve explain "metadata panel"
reponerve explain-file "metadata-panel.tsx"
reponerve explain-function "BuildMetadataPanel"
reponerve ask "Who owns authentication?"
reponerve impact "user-service"
reponerve review "metadata panel"
reponerve plan "Add OAuth login"
```

---

# What RepoNerve Does

RepoNerve delivers **Software Understanding** through knowledge preservation, software memory, intelligence capabilities, and Development Experience.

```text
Knowledge Preservation          (Core Platform Capability)
    ↓
Software Memory
    ↓
Repository Intelligence + Code Intelligence
    ↓
Repository-Code Linking
    ↓
Feature Understanding
    ↓
Development Experience
    ↓
Software Understanding
    ↓
CLI / MCP
    ↓
Developers and AI Agents
```

**Knowledge Preservation** — core platform foundation. Stores memory, ownership, context, code entities, and repository-code links. All intelligence layers depend on it.

**Software Memory** — durable organizational knowledge that survives contributor turnover and architectural evolution.

**Repository Intelligence** — why the software exists (memory, context, ownership, graph, discovery, reviewers, change planning).

**Code Intelligence** — how the software works (modules, packages, symbols, call graphs).

**Repository-Code Linking** — deterministic connections between repository entities (decisions, facts, events) and code entities (files, symbols). Required for unified explain output.

**Feature Understanding** — feature-level resolution: Feature → Code → Ownership → Decisions → Impact.

**Development Experience** — how users consume RepoNerve (`ask`, `explain`, `plan`, `impact`, `review`, and symbol-level explain commands).

---

# Core Capabilities

## Code Intelligence

Deterministic code structure extraction and analysis.

* File, package, type, and function indexing
* Symbol resolution
* Call graph and dependency analysis

## Memory Engine

Build a repository memory graph from:

* Git history
* ADRs
* Repository metadata

Extract:

* Events
* Decisions
* Intents
* Facts
* Relationships

## Query Engine

Explore repository knowledge. Part of the **Understanding Engine** — the retrieval layer spanning repository memory, code intelligence, and repository-code links.

Commands:

```bash
reponerve memory list decisions
reponerve memory get decision <id>
reponerve memory trace decision <id>
reponerve memory explain decision <id>
```

## Context Engine

Generate repository context.

Commands:

```bash
reponerve context generate
reponerve context export
```

## Development Experience

Development-facing workflows that orchestrate Code Intelligence and Repository Intelligence.

v1.0 commands (ISSUE-057):

```bash
reponerve ask "Who created metadata panel?"
reponerve explain "metadata panel"
reponerve explain-file "metadata-panel.tsx"
reponerve explain-function "BuildMetadataPanel"
reponerve explain-struct "MetadataPanel"
reponerve explain-interface "Searcher"
reponerve explain-type "HandlerFunc"
reponerve plan "Add OAuth login"
reponerve impact "user-service"
reponerve review "metadata panel"
```

## MCP Server

Expose intelligence directly to AI coding agents.

Start the server:

```bash
reponerve mcp
```

Supported MCP capabilities:

* Memory Queries, Trace, and Explain
* Repository Context Generation and Export
* Ownership and Contributor Queries
* Repository Intelligence
* Knowledge Graph Traversal and Impact Analysis

Compatible with:

* Claude Code
* Cursor
* Windsurf
* Cline
* Roo
* Codex
* GitHub Copilot Chat

---

# Philosophy

Understanding first.

Evidence second.

AI third.

Software Understanding is the outcome. Development Experience is the product surface.

---

# Why RepoNerve (Token Economics)

Premium LLM models are expensive and getting costlier. Most agent cost is not generation — it is **re-exploring the repository** every session (file reads, greps, summaries).

RepoNerve inverts that:

```text
EXPENSIVE:  LLM reads repo → LLM understands → LLM acts
CHEAP:      reponerve scan (0 LLM tokens) → MCP context pack → LLM acts
```

* **Scan once** — deterministic extraction, no LLM required
* **Query cheaply** — structured MCP tools return bounded evidence
* **Persist understanding** — session 50 does not re-pay the exploration tax

See `docs/product/token-economics.md`.

---

# Market Position

RepoNerve is **Software Understanding Infrastructure** — not another code graph, not generic chat memory, not an autonomous coding agent.

Code-graph tools answer *where*. RepoNerve answers *why*, *who*, and *what breaks* — with mandatory evidence.

See `docs/product/market-positioning.md`.

---

# Greenfield Projects

RepoNerve does not build a repository from an idea. It ensures a repository built from an idea **stays understandable** — capture ADRs and scan from the first commit so agents never accumulate amnesia.

See `docs/product/greenfield-guide.md`.

---

# Current Status

Release Status:

```text
Knowledge Preservation — Core Platform Capability
Repository Intelligence — Complete
Code Intelligence — Complete (Go + 19 languages)
Repository-Code Linking — Complete
Feature Understanding — Complete
Development Experience — Complete
Token Intelligence — Complete
Evidence Graph & Session Memory — Complete
Multi-Language Code Intelligence — Complete
Software Understanding — Delivered
v1.0.0 — Released (2026-06-18); latest v1.1.0
```

Completed Milestones:

```text
v0.1.0-alpha  ✓ Ingestion Engine
v0.2.0-alpha  ✓ Memory Engine
v0.3.0-alpha  ✓ Query Engine
v0.4.0-alpha  ✓ Context Engine
v0.5.0-alpha  ✓ MCP Server
v0.7.0-alpha  ✓ Ownership Intelligence
v0.8.0-alpha  ✓ Knowledge Graph Intelligence
v0.9.0-alpha  ✓ Repository Intelligence
v0.10.0-alpha ✓ Foundation Fixes (ISSUE-059)
v0.12.0-alpha ✓ Code Intelligence + DE (ISSUE-057)
v0.13.0-alpha ✓ Token Intelligence (ISSUE-060)
v0.14.0-alpha ✓ Graph + Session Memory (ISSUE-061)
v0.15.0-alpha ✓ Multi-Language (ISSUE-062)
```

Current Focus:

```text
v1.0.0 release — git tag and publish release notes
```

See `docs/releases/v1.0.0-checklist.md` and `docs/audits/v1.0-release-review.md`.

---

# Installation

## Option 1: Install From Release Artifacts

Download the archive for your OS and architecture from GitHub Releases, then place the `reponerve` binary on your `PATH`.

## Option 2: Build From Source

```bash
git clone https://github.com/reponerve/reponerve.git
cd reponerve
make build
```

The binary is produced at the repository root as `./reponerve`.

## Option 3: Homebrew (After Tap Is Published)

```bash
brew tap reponerve/reponerve
brew install reponerve
```

---

# Quick Start

From the root of the repository you want to analyze:

1. Initialize workspace metadata and database:

```bash
reponerve init
```

2. Ingest repository signals (commits, ADRs, metadata):

```bash
reponerve scan
```

3. Inspect extracted decisions:

```bash
reponerve memory list decisions
```

4. Generate repository context:

```bash
reponerve context generate
```

5. Start MCP server for agent integrations:

```bash
reponerve mcp
```

## AI Chat Integration

RepoNerve is built for **direct conversation in AI chat** — in Cursor, VS Code Copilot, JetBrains, Windsurf, Continue, Claude, and other MCP hosts. Type natural language; the assistant calls RepoNerve tools and answers from evidence. **Any LLM** the IDE provides works (GPT, Claude, Gemini, Llama, …) — RepoNerve does not require its own API keys.

```bash
reponerve init && reponerve scan   # once per repository; init installs skill + MCP
```

**Full guide:** `docs/ai-chat-integration.md`  
**IDE matrix:** `docs/mcp/compatibility-matrix.md`  
**Refresh IDE files:** `reponerve integrate` (or `reponerve integrate --force`)

| IDE / client | Config in this repo |
| --- | --- |
| Cursor | `.cursor/mcp.json` + `.cursor/skills/reponerve/` |
| VS Code + Copilot | `.vscode/mcp.json` |
| Continue | `.continue/mcpServers/reponerve.json` |

Example prompts in chat: "Onboard me", paste a ticket → "Where do I start?", "Why do we use SQLite?", "Explain internal/mcp/server/server.go".

### GitHub Copilot Chat (VS Code)

See `docs/copilot-chat-integration.md`. Open `.vscode/mcp.json` → Start → Copilot Chat → **Agent** mode.

### Cursor

See `docs/cursor-integration.md`. **Skill + MCP:** context-first workflow plus 43 MCP tools.

---

# Documentation

## Start Here

* Product Overview: `README.md`
* Contributor Setup: `docs/governance/contribution-guide.md`
* MCP Troubleshooting: `docs/mcp/troubleshooting.md`

## By Goal

* **AI chat in any IDE or LLM:** `docs/ai-chat-integration.md`
* Understand product direction: `docs/vision/`
* Market positioning and competitors: `docs/product/market-positioning.md`
* Token economics and AI cost optimization: `docs/product/token-economics.md`
* Greenfield / build-from-scratch workflows: `docs/product/greenfield-guide.md`
* Honest code vs docs status: `docs/product/implementation-status.md`
* Understand architecture: `docs/architecture/`
* Track planned work: `docs/roadmap/`
* v0.x → v1.0 iteration plan: `docs/roadmap/v1.0-iteration-plan.md`
* Review quality and release readiness: `docs/audits/`
* Follow release process: `docs/releases/v1.0.0-checklist.md` and `docs/releases/v1.0.0.md`

---

# Roadmap

## Completed

* Ingestion Engine
* Memory Engine
* Query Engine
* Context Engine
* MCP Server
* Ownership Intelligence
* Knowledge Graph Intelligence
* Repository Intelligence
* Agent Context Builder
* Repository Search
* Agent Session Intelligence
* Workflow Intelligence

## v1.0.0 Release (shipped)

| Iteration | Issue | Status |
| --- | --- | --- |
| v0.10.0-alpha | ISSUE-059 | ✅ Foundation fixes |
| v0.11–v0.12.0-alpha | ISSUE-057 | ✅ Code Intelligence + Development Experience |
| v0.13.0-alpha | ISSUE-060 | ✅ Token Intelligence (in v1.0.0) |
| v0.14.0-alpha | ISSUE-061 | ✅ Evidence Graph + Session Memory (in v1.0.0) |
| v0.15.0-alpha | ISSUE-062 | ✅ Multi-language code intelligence (in v1.0.0) |
| **v1.0.0** | — | ✅ Tagged 2026-06-18 |
| **v1.0.1** | — | ✅ Patch — scan reliability |
| **v1.1.0** | — | ✅ Bounded DE, feature intelligence, native discipline |

See `docs/releases/v1.1.0.md` and `docs/releases/versioning.md` for post-1.0 semver policy.

---

# License

RepoNerve is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.

Copyright (c) 2026 RepoNerve Contributors
