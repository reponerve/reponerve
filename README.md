# RepoNerve

> Repository Intelligence for Humans and AI Agents.

RepoNerve is an open-source platform that captures repository knowledge, builds a structured memory graph, generates repository context, and exposes that intelligence through MCP (Model Context Protocol).

Software remembers code.

Software forgets why.

RepoNerve preserves the why.

---

# Vision

Every repository should be able to explain itself.

To developers.

To teams.

To AI coding agents.

---

# What RepoNerve Does

RepoNerve transforms repository artifacts into structured knowledge.

Repository
↓
Ingestion
↓
Memory Graph
↓
Query Engine
↓
Context Engine
↓
MCP Server
↓
AI Agents

RepoNerve extracts:

* Events
* Decisions
* Intents
* Facts
* Relationships

and converts them into actionable repository intelligence.

---

# Core Capabilities

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

---

## Query Engine

Explore repository knowledge.

Commands:

```bash
reponerve memory list decisions

reponerve memory get decision <id>

reponerve memory trace decision <id>

reponerve memory explain decision <id>
```

---

## Context Engine

Generate repository context.

Commands:

```bash
reponerve context generate

reponerve context export
```

Example output:

```text
Repository Context

Key Decisions
...

Key Intents
...

Key Facts
...

Recent Events
...
```

---

## MCP Server

Expose repository intelligence directly to AI coding agents.

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

---

# Philosophy

Memory First.

Context Second.

Agents Third.

---

# Current Status

Release Status:

```text
v1.0.0 release-ready
```

Completed Milestones:

```text
v0.1.0-alpha
✓ Ingestion Engine

v0.2.0-alpha
✓ Memory Engine

v0.3.0-alpha
✓ Query Engine

v0.4.0-alpha
✓ Context Engine

v0.5.0-alpha
✓ MCP Server

v0.7.0-alpha
✓ Ownership Intelligence

v0.8.0-alpha
✓ Knowledge Graph Intelligence

v0.9.0-alpha
✓ Repository Intelligence

v1.0.0
✓ Agent Context, Search, Session, Workflow, and Production Readiness
```

Current Focus:

```text
v1.x
Backlog curation and post-release follow-up
```

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

If output is empty after scan, add commits and/or ADR files, then run `reponerve scan` again.

---

# Documentation

## Start Here

* Product Overview: `README.md`
* Contributor Setup: `docs/governance/contribution-guide.md`
* MCP Troubleshooting: `docs/mcp/troubleshooting.md`

## By Goal

* Understand product direction: `docs/vision/`
* Understand architecture: `docs/architecture/`
* Track planned work: `docs/roadmap/`
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
* Repository Onboarding
* Repository Q&A
* Impact Analysis
* Architectural Guidance
* Context Compression
* Agent Context Builder
* Repository Search
* Agent Session Intelligence
* Workflow Intelligence

## Planned

* Workflow Templates
* Session Export
* Search Adapters
* Semantic Search Experiments
* Hybrid Search
* User Defined Workflows
* Agent Handoff Bundles

---

# License

RepoNerve is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.

Copyright (c) 2026 RepoNerve Contributors
