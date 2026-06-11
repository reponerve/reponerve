# RepoNerve Milestones

Version: 1.0

Status: Draft

Authors: RepoNerve Contributors

Last Updated: 2026-06-11

---

# Purpose

This document defines the engineering and product milestones required to deliver RepoNerve.

The product mission is **Software Understanding** — preserving and transferring software knowledge so humans and AI can understand, change, and evolve software without repeated repository exploration.

The roadmap principle remains: build repository memory first, then deliver Software Understanding through Development Experience.

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
| Phase 7 | v0.x Iterations (ISSUE-059–062) → v1.0 scope |
| Phase 8 | v1.0.0 Release                                 |

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

# Phase 7: Code Intelligence & Development Experience

Status:

Planned

Target:

Code Understanding + Development-Facing Orchestration

Issue:

ISSUE-057

Depends On:

ARCH-001 — Architecture Realignment For Software Understanding (architecture approval)

Strategy and iteration plan:

* `docs/roadmap/v1.0-iteration-plan.md` — **authoritative v0.x → v1.0 path**
* `docs/product/implementation-status.md` — honest code snapshot
* `docs/product/token-economics.md` — AI cost optimization thesis
* `docs/product/market-positioning.md` — competitive landscape

---

## Release Strategy

**v1.0.0 is the only product release.** All scoped capabilities ship together.

**v0.x.0-alpha tags are engineering iterations** toward v1.0 — not partial product releases.

| Tag | Issue | Focus |
| --- | --- | --- |
| `v0.10.0-alpha` | ISSUE-059 | Foundation fixes (expertise, CLI exposure, debt) |
| `v0.11.0-alpha` | ISSUE-057 steps 1–4 | Code Intelligence core |
| `v0.12.0-alpha` | ISSUE-057 steps 5–9 | Development Experience + linking |
| `v0.13.0-alpha` | ISSUE-060 | Token Intelligence layer |
| `v0.14.0-alpha` | ISSUE-061 | Evidence Graph + Session Memory |
| `v0.15.0-alpha` | ISSUE-062 | Multi-language code intelligence |
| `v1.0.0` | Phase 8 | Full acceptance + release tag |

Alpha tags validate incremental progress. `v1.0.0` is tagged only when **all** rows above are complete.

---

## Objective

Deliver Code Intelligence, Repository-Code Linking, Feature Understanding, and Development Experience — completing RepoNerve v1.0.

Repository Intelligence is complete.

This phase delivers:

```text
Code Intelligence
    +
Repository-Code Linking
    +
Feature Understanding
    +
Development Experience
    =
Software Understanding
```

---

## Deliverables

### Code Intelligence

```text
Symbol extraction
File graph
Package graph
Call graph
Symbol dependency analysis
```

Entities: Files, Packages, Types, Interfaces, Functions, Methods, API Endpoints

Relationships: CALLS, IMPORTS, IMPLEMENTS, DEPENDS_ON, EXPOSES_ENDPOINT

---

### Repository-Code Linking

```text
Cross-authority link extraction
repository_code_relationships storage
Bidirectional link traversal
Evidence on every link
```

---

### Feature Understanding

```text
Feature topic resolution
Feature → Code → Ownership → Decisions → Impact
```

---

### Development Experience

```text
ask
explain
explain-file
explain-function
explain-struct
explain-interface
explain-type
plan
impact
review
```

Explain output combines Code Context + Repository Context.

---

## Exit Criteria

Code structure indexed deterministically.

All development-facing CLI commands work end-to-end.

Explain output combines code and repository context.

Evidence and provenance preserved.

Humans and AI agents can understand and evolve software with minimal repository exploration.

---

# Phase 8: v1.0 Release

Status:

Blocked on Phase 7 (ISSUE-059 through ISSUE-062)

Target:

Production-Ready Open Source Release

---

## Objective

Release stable RepoNerve v1.0 with Code Intelligence, Repository-Code Linking, Feature Understanding, and Development Experience.

---

## Deliverables

### Stable CLI

```text
init
scan
ask
explain
explain-file
explain-function
explain-struct
explain-interface
explain-type
plan
impact
review
context
mcp
```

---

### Stable Code Intelligence

---

### Stable Development Experience

---

### Stable Repository Intelligence

---

### Stable MCP Integration

---

### Public Documentation

---

### Contributor Documentation

---

## Exit Criteria

Public v1.0 release completed.

Release state:

| Capability | Required |
| --- | --- |
| Knowledge Preservation | ✅ Core platform operational |
| Repository Intelligence | ✅ |
| Code Intelligence | ✅ |
| Repository-Code Linking | ✅ |
| Feature Understanding | ✅ |
| Development Experience | ✅ |
| Software Understanding | ✅ |
| RepoNerve v1.0 | 🚀 Released |

---

# Out of Scope (Not v1.0)

The following are intentionally outside v1.0 and not planned as a separate product release:

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

RepoNerve preserves software knowledge and transfers understanding.

Understanding first. Evidence second. AI third.

Software Understanding is the outcome. Development Experience is how users consume RepoNerve.
