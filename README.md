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
reponerve reuse-check "add OAuth middleware"
reponerve ship-check "OAuth login"
reponerve doctor
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
* Go + 19 Tree-sitter languages

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

```bash
reponerve ask "Who created metadata panel?"
reponerve explain "metadata panel"
reponerve explain-feature "Authentication"
reponerve explain-file "metadata-panel.tsx"
reponerve explain-function "BuildMetadataPanel"
reponerve explain-struct "MetadataPanel"
reponerve explain-interface "Searcher"
reponerve explain-type "HandlerFunc"
reponerve plan "Add OAuth login"
reponerve impact "user-service"
reponerve review "metadata panel"
reponerve reuse-check "add rate limiter"
reponerve ship-check "metadata panel"
reponerve onboard
reponerve doctor
```

## Local Explore UI

Browse repository knowledge as an interactive graph:

```bash
reponerve explore --serve    # http://127.0.0.1:8765/
reponerve explore -o reponerve-graph.html
```

## MCP Server

Expose intelligence directly to AI coding agents.

Start the server:

```bash
reponerve mcp
```

**49 MCP tools** across memory, ownership, graph, and Development Experience.

Compatible with:

* Claude Code
* Cursor
* Windsurf
* Cline
* Roo
* Codex
* GitHub Copilot Chat

Works in **AI chat without MCP** — run `reponerve ask --json` in the terminal; same evidence envelope as MCP.

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

**Latest release:** `v1.5.1` (2026-06-24)

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
Native Development Discipline — Complete
Reuse Protocol + Ship Readiness — Complete
Local Explore UI — Complete
Software Understanding — Delivered
v1.0.0 — Released (2026-06-18)
Post-1.0 — Semver on main (v1.5.1)
```

**Post-1.0 focus:** RFC-gated capabilities from `docs/roadmap/v1.x-backlog.md`.

See `docs/product/implementation-status.md` and `docs/releases/versioning.md`.

---

# Installation

**Full guide:** [`docs/install.md`](docs/install.md)

## npm (Node 18+)

```bash
npm install -g reponerve
# or per project:
npm install -D reponerve && npx reponerve init
```

## No Go required (shell)

```bash
curl -fsSL https://raw.githubusercontent.com/reponerve/reponerve/main/scripts/install.sh | bash
```

Or download an archive for your OS from [GitHub Releases](https://github.com/reponerve/reponerve/releases) and put `reponerve` on your `PATH`.

## Go developers

```bash
go install github.com/reponerve/reponerve/cmd/reponerve@v1.5.1
# or from a clone:
make install
```

## Homebrew

```bash
brew tap reponerve/tap
brew install reponerve
```

Until the tap is published, use the install script or release archives. See `docs/install.md`.

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

3. Verify memory health:

```bash
reponerve doctor
```

4. Inspect extracted decisions:

```bash
reponerve memory list decisions
```

5. Generate repository context:

```bash
reponerve context generate
```

6. Start MCP server for agent integrations:

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

See `docs/cursor-integration.md`. **Skill + MCP:** context-first workflow plus 49 MCP tools.

---

# Documentation

**Documentation index:** [`docs/README.md`](docs/README.md)

## Start Here

* Product Overview: `README.md`
* Install: `docs/install.md`
* AI chat in any IDE: `docs/ai-chat-integration.md`
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
* Post-1.0 planned work: `docs/roadmap/v1.x-backlog.md`
* Release history: `docs/releases/versioning.md`
* Review quality and release readiness: `docs/audits/`

---

# Roadmap

## Shipped (v1.0 + post-1.0)

| Release | Highlights |
| --- | --- |
| **v1.0.0** | First product release — complete v1.0 scope (2026-06-18) |
| **v1.0.1** | Scan reliability patch |
| **v1.1.0** | Bounded DE, feature intelligence v2, native discipline (RFC-001–003) |
| **v1.2.0** | Reuse Protocol + Ship Readiness (RFC-003 B/C) |
| **v1.3.0** | Discipline policy, PR context, document paths (RFC-003 D, RFC-004, RFC-005) |
| **v1.3.1** | Binary-first install script + release archives |
| **v1.3.2** | npm distribution (RFC-006) |
| **v1.4.0** | Doctor, scoped monorepo scan, Homebrew (RFC-007, RFC-008) |
| **v1.5.0** | Local Explore UI (RFC-009) |
| **v1.5.1** | CLI `--version` / `version` command |

See `docs/releases/v1.5.1.md` and `docs/releases/versioning.md` for semver policy.

**Next:** capabilities in `docs/roadmap/v1.x-backlog.md` (RFC-gated).

---

# License

RepoNerve is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.

Copyright (c) 2026 RepoNerve Contributors
