# RepoNerve Milestones

Version: 1.0

Status: Draft

Authors: RepoNerve Contributors

Last Updated: 2026-06-05

---

# Purpose

This document defines the engineering and product milestones required to deliver RepoNerve.

The roadmap is designed around a single principle:

> Build repository memory first.

Everything else is layered on top of memory.

The roadmap intentionally prioritizes:

1. Memory
2. Understanding
3. Context
4. AI Integration

in that order.

---

# Long-Term Vision

RepoNerve aims to become:

> The Memory and Context Engine for Software Repositories

The roadmap progresses from:

```text
Repository
    ▼
Memory
    ▼
Understanding
    ▼
Context
    ▼
AI Skills
```

---

# Roadmap Overview

| Phase   | Goal                    |
| ------- | ----------------------- |
| Phase 0 | Foundation              |
| Phase 1 | Repository Scanner      |
| Phase 2 | Memory Engine           |
| Phase 3 | Query Engine            |
| Phase 4 | Context Engine          |
| Phase 5 | MCP Skills              |
| Phase 6 | Repository Intelligence |
| Phase 7 | v1.0 Release            |

---

# Phase 0: Foundation

Status:

Completed

Target:

Project Initialization

---

## Objective

Establish project foundations.

---

## Deliverables

### Repository Setup

```text
Git Repository

Branch Strategy

Labels

Issue Templates

PR Templates
```

---

### Documentation

```text
Vision

Mission

Project Charter

PRD

Architecture Docs
```

---

### Development Environment

```text
Go Modules

Linting

Formatting

Testing

CI Pipeline
```

---

## Exit Criteria

Project ready for engineering work.

---

# Phase 1: Repository Scanner

Status:

Completed

Target:

Repository Discovery

---

## Objective

Build repository scanning capabilities.

---

## Deliverables

### Repository Discovery

Support:

```text
Local Git Repositories
```

---

### Git Scanner

Extract:

```text
Commits

Authors

Branches

File Changes
```

---

### ADR Scanner

Extract:

```text
Architecture Decision Records
```

---

### Documentation Scanner

Extract:

```text
Markdown Files
```

---

### Incremental Detection

Detect:

```text
New Commits

Modified Files

Updated ADRs
```

---

## CLI Commands

```bash
reponerve init

reponerve scan
```

---

## Exit Criteria

Repository artifacts successfully discovered and parsed.

---

# Phase 2: Memory Engine

Status:

Completed

Target:

Repository Memory Creation

---

## Objective

Convert repository artifacts into memory.

---

## Deliverables

### Fact Extraction

Examples:

```text
Service Uses Database

API Depends On Service
```

---

### Event Extraction

Examples:

```text
PR Merged

Commit Created
```

---

### Decision Extraction

Examples:

```text
Use Redis

Adopt Kafka
```

---

### Intent Extraction

Examples:

```text
Reduce Latency

Improve Scalability
```

---

### Evidence System

Every memory must include:

```text
Source

Confidence

Traceability
```

---

## Exit Criteria

Repository memory successfully generated.

---

# Phase 3: Query Engine

Status:

Completed

Target:

Repository Memory Retrieval

---

## Objective

Answer repository questions.

---

## Deliverables

### Search Engine

Based on:

```text
SQLite FTS5
```

---

### Memory Retrieval

Retrieve:

```text
Facts

Events

Decisions

Intent
```

---

### Evidence Retrieval

Attach:

```text
PRs

ADRs

Commits
```

---

## CLI Commands

```bash
reponerve ask
```

---

## Example Queries

```text
Why was Redis introduced?

Who introduced this service?

What problem was this solving?
```

---

## Exit Criteria

Repository questions answered with evidence.

---

# Phase 4: Context Engine

Status:

Completed

Target:

Development Context Generation

---

## Objective

Generate task-specific repository context.

---

## Deliverables

### Intent Detection

Identify:

```text
Authentication

Caching

Billing
```

---

### Context Retrieval

Retrieve:

```text
Relevant Decisions

Relevant Files

Relevant Services
```

---

### Context Ranking

Reduce noise.

---

### Context Pack Generation

Create:

```text
Task-Specific Context
```

---

## CLI Commands

```bash
reponerve context
```

---

## Example

```bash
reponerve context "Add MFA support"
```

---

## Exit Criteria

Relevant context packs generated.

---

# Phase 5: MCP Skills

Status:

Completed

Target:

AI Integration

---

## Objective

Expose RepoNerve capabilities to AI systems.

---

## Deliverables

### MCP Server

Support:

```text
Model Context Protocol
```

---

### Repository Memory Skill

Tool:

```text
get_repository_memory
```

---

### Component Explanation Skill

Tool:

```text
explain_component
```

---

### Decision Retrieval Skill

Tool:

```text
find_related_decisions
```

---

### Context Pack Skill

Tool:

```text
get_context_pack
```

---

## Exit Criteria

AI systems can consume RepoNerve memory.

---

# Phase 6: Repository Intelligence

Status:

Completed

Target:

Repository Reasoning

---

## Objective

Provide higher-level repository understanding.

---

## Deliverables

### Pattern Memory

Examples:

```text
Authentication Pattern

Logging Pattern

Repository Pattern
```

---

### Similar Implementation Discovery

Examples:

```text
Show Similar MFA Changes
```

---

### Architecture Insights

Examples:

```text
Identify Architectural Hotspots
```

---

### Technical Debt Insights

Examples:

```text
Identify Legacy Components
```

---

## Exit Criteria

Repository understanding exceeds simple retrieval.

---

# Phase 7: v1.0 Release

Status:

Release Readiness

Target:

Production-Ready Open Source Release

---

## Objective

Release stable RepoNerve platform.

---

## Deliverables

### Stable CLI

```text
init

scan

ask

explain

context
```

---

### Stable Memory Engine

---

### Stable Query Engine

---

### Stable Context Engine

---

### Stable MCP Integration

---

### Public Documentation

---

### Contributor Documentation

---

## Exit Criteria

Public v1.0 release completed.

---

# Future Roadmap

The following are intentionally outside the current roadmap.

---

## Web Dashboard

Future consideration.

---

## Cloud Platform

Future consideration.

---

## Multi-Repository Memory

Future consideration.

---

## Enterprise Features

Future consideration.

---

## Team Collaboration

Future consideration.

---

# Success Metrics

---

## Memory Quality

Accurate repository memory.

---

## Evidence Quality

Strong traceability.

---

## Query Accuracy

Reliable answers.

---

## Context Quality

Relevant context packs.

---

## Token Reduction

Reduced AI repository exploration.

---

## Adoption

Community adoption and contributions.

---

# Roadmap Principles

---

## Principle 1

Memory Before Intelligence

---

## Principle 2

Context Before Automation

---

## Principle 3

Understanding Before Generation

---

## Principle 4

Open Source First

---

## Principle 5

Local First

---

# Guiding Statement

RepoNerve will not become valuable because it generates code.

RepoNerve will become valuable because it preserves and serves the knowledge behind software systems.

Memory first.

Context second.

AI third.
