# RepoNerve MVP Scope

Version: 1.0

Status: Draft

Authors: RepoNerve Contributors

Last Updated: 2026-06-05

---

# Purpose

This document defines the Minimum Viable Product (MVP) scope for RepoNerve.

The purpose of the MVP is to validate the core hypothesis:

> Repository memory can significantly reduce the effort required to understand software systems.

The MVP must remain intentionally focused.

Any feature that does not directly contribute to validating the core hypothesis should be deferred.

---

# MVP Objective

RepoNerve MVP should answer one question exceptionally well:

> Why does this exist?

The MVP should help developers understand:

* Why components were created
* Why technologies were adopted
* Why architectural decisions were made
* What historical context influenced repository evolution

---

# MVP Success Criteria

The MVP is successful if a developer can:

* Connect a repository
* Generate repository memory
* Query repository memory
* Receive evidence-backed answers

without manually searching:

* Git history
* Pull requests
* Issues
* ADRs

---

# MVP Philosophy

The MVP focuses on:

```text
Memory First
```

Not:

```text
Code Generation
```

Not:

```text
Autonomous Agents
```

Not:

```text
Workflow Automation
```

---

# Target Users

Primary users:

* Individual Developers
* Open Source Maintainers

Secondary users:

* Staff Engineers

AI agent integrations are intentionally deferred.

---

# MVP Product Shape

RepoNerve MVP is:

```text
CLI First
Local First
Single Repository Focus
```

No web UI is required.

No cloud services are required.

No accounts are required.

---

# MVP User Journey

## Step 1

Initialize RepoNerve

```bash
reponerve init
```

Creates local RepoNerve workspace.

---

## Step 2

Scan Repository

```bash
reponerve scan
```

Analyzes repository artifacts.

Creates repository memory.

---

## Step 3

Ask Questions

```bash
reponerve ask "Why was Redis introduced?"
```

Returns evidence-backed answers.

---

## Step 4

Explain Components

```bash
reponerve explain services/auth
```

Provides component understanding.

---

# MVP Features

---

# Feature 1: Repository Initialization

Priority: P0

Command:

```bash
reponerve init
```

Responsibilities:

* Create RepoNerve workspace
* Generate configuration
* Initialize local storage

---

# Feature 2: Repository Scanning

Priority: P0

Command:

```bash
reponerve scan
```

Responsibilities:

* Discover repository artifacts
* Build repository memory
* Create searchable knowledge

---

# Feature 3: Git History Ingestion

Priority: P0

Inputs:

* Commits
* Authors
* File changes
* Commit messages

Outputs:

* Events
* Facts
* Relationships

---

# Feature 4: Pull Request Ingestion

Priority: P0

Inputs:

* Pull requests
* Reviews
* Discussions

Outputs:

* Decisions
* Intent
* Historical context

---

# Feature 5: ADR Ingestion

Priority: P0

Inputs:

* Architecture Decision Records

Outputs:

* Decisions
* Alternatives
* Tradeoffs

---

# Feature 6: Memory Extraction Engine

Priority: P0

Memory Types:

* Facts
* Events
* Decisions
* Ownership
* Relationships
* Intent

Responsibilities:

* Extract repository knowledge
* Link repository artifacts
* Create searchable memory

---

# Feature 7: Repository Question Engine

Priority: P0

Command:

```bash
reponerve ask
```

Example Questions:

```text
Why was Redis introduced?

Who introduced this service?

Why was Kafka selected?

What problem was this solving?
```

---

# Feature 8: Component Explanation Engine

Priority: P0

Command:

```bash
reponerve explain
```

Example:

```bash
reponerve explain services/auth
```

Outputs:

* Purpose
* Dependencies
* Related decisions
* Ownership
* Historical context

---

# Feature 9: Evidence Engine

Priority: P0

Every answer must include sources.

Example:

```text
Decision:
Use Redis

Reason:
Reduce database latency

Sources:
- PR #143
- ADR-14
- Commit abc123
```

Evidence is mandatory.

---

# MVP Memory Types

Supported:

* Facts
* Events
* Decisions
* Intent

Partial support:

* Ownership
* Relationships

Deferred:

* Pattern Memory

---

# Supported Inputs

## Repository Source Code

Required.

---

## Git History

Required.

---

## Pull Requests

Required.

---

## ADRs

Required.

---

## Documentation

Optional.

Limited support.

---

# Storage Requirements

MVP must work locally.

Suggested storage:

## Metadata

SQLite

---

## Search

SQLite FTS5

---

## Memory Store

SQLite

---

# AI Requirements

AI should be used only when necessary.

Examples:

* Decision extraction
* Intent extraction
* Tradeoff extraction

The majority of repository understanding should rely on deterministic analysis.

---

# Explicit Non-Goals

The following are intentionally excluded from MVP.

---

## Code Generation

Not supported.

---

## Autonomous Coding

Not supported.

---

## PR Generation

Not supported.

---

## Bug Fixing

Not supported.

---

## Agent Workflows

Not supported.

---

## Multi-Agent Systems

Not supported.

---

## Knowledge Graph Visualization

Not supported.

---

## Web Dashboard

Not supported.

---

## Cloud Platform

Not supported.

---

## Multi-Repository Reasoning

Not supported.

---

## Team Collaboration

Not supported.

---

# Deferred Features

These features may be implemented after MVP validation.

---

## Context Packs

Future Version

---

## AI Agent Skills

Future Version

---

## MCP Integration

Future Version

---

## Repository Pattern Memory

Future Version

---

## Knowledge Graph

Future Version

---

## Web Interface

Future Version

---

## Multi-Repository Memory

Future Version

---

# MVP Deliverables

RepoNerve MVP is complete when:

* Repository memory can be generated.
* Historical decisions can be retrieved.
* Repository questions can be answered.
* Answers include evidence.
* Developers can understand repository context faster.

---

# MVP Exit Criteria

The MVP exits validation when users can reliably answer:

* Why does this exist?
* Who introduced it?
* What problem was it solving?
* What decision led to this change?

without manually reconstructing repository history.

At that point, RepoNerve has successfully demonstrated the value of repository memory.

---

# Guiding Principle

The MVP should optimize for understanding software systems, not generating software.

Memory is the product.

Everything else is built on top of memory.
