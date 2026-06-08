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

# Quick Start

Initialize a workspace:

```bash
reponerve init
```

Scan a repository:

```bash
reponerve scan
```

Generate repository context:

```bash
reponerve context generate
```

Start MCP:

```bash
reponerve mcp
```

---

# Documentation

## Vision

* docs/vision/

## Architecture

* docs/architecture/

## Roadmaps

* docs/roadmap/

## Audits

* docs/audits/

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
