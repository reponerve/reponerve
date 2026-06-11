# RepoNerve Agent Instructions

## Purpose

RepoNerve is the intelligence layer for software understanding.

Its purpose is to preserve, organize, and transfer software knowledge so that understanding survives beyond individual contributors and remains accessible to both humans and AI systems.

RepoNerve serves as a **software memory system** — preserving understanding that would otherwise be lost through contributor turnover, architectural evolution, and documentation drift.

---

## Product Vision

RepoNerve delivers **Software Understanding** through:

```text
Knowledge Preservation          (Core Platform Capability)
    ↓
Software Memory
    ↓
Repository Intelligence         (complete)
    +
Code Intelligence               (ISSUE-057)
    ↓
Repository-Code Linking           (ISSUE-057)
    ↓
Feature Understanding           (ISSUE-057)
    ↓
Development Experience          (product surface — ISSUE-057)
    ↓
Software Understanding          (outcome)
```

Repository Intelligence, Code Intelligence, and Development Experience are capabilities — not the whole product mission.

Development Experience (`ask`, `explain`, `explain-file`, `explain-function`, `explain-struct`, `explain-interface`, `explain-type`, `plan`, `impact`, `review`) is how users consume RepoNerve.

Software Understanding is what they receive.

v1.0 is blocked until ISSUE-057 completes all v1.0 scope. See `docs/roadmap/v1.0-prd.md` and `docs/vision/vision.md`.

**Implementation status:** Repository Intelligence is shipped. Code Intelligence and most Development Experience commands are not yet implemented. See `docs/product/implementation-status.md`.

**Product strategy docs:**

- Token economics: `docs/product/token-economics.md`
- Market positioning: `docs/product/market-positioning.md`
- Greenfield workflows: `docs/product/greenfield-guide.md`
- v0.x → v1.0 iteration plan: `docs/roadmap/v1.0-iteration-plan.md`

---

## Core Principles

Understanding First.

Evidence Second.

AI Third.

RepoNerve optimizes AI usage by moving understanding out of the LLM: scan deterministically, query via MCP, deliver token-budget context packs. Premium models spend tokens on implementation, not re-exploration.

---

## Current Development Phase

Phase 7 - v0.x Iterations Toward v1.0 (ISSUE-059 through ISSUE-062)

Current Issues (all required for v1.0.0):

ISSUE-059 → ISSUE-057 → ISSUE-060 → ISSUE-061 → ISSUE-062

Current Release Target:

v1.0.0 — the only product release; delivered via v0.10–v0.15 alpha iterations (see `docs/roadmap/v1.0-iteration-plan.md`)

---

## Architecture Rules

Always follow:

- docs/architecture/architecture-overview.md
- docs/architecture/package-structure.md
- docs/architecture/repository-ingestion.md
- docs/architecture/event-flows.md

Never introduce architectural changes without RFC approval.

---

## Technology Stack

Language:

Go

CLI:

Cobra

Database:

SQLite

Configuration:

Viper

Search:

SQLite FTS5

Testing:

Go Testing Framework

---

## Repository Rules

The CLI is an interface.

The platform is the product.

Business logic must never be implemented inside CLI commands.

---

## Dependency Rules

Allowed:

CLI
→ Services
→ Storage

Forbidden:

Storage
→ CLI

Query Engine
→ CLI

Memory Engine
→ MCP

---

## Local First

Do not introduce:

- Cloud services
- SaaS dependencies
- External infrastructure

unless explicitly requested.

---

## AI Usage Policy

AI should only be used for:

- Intent extraction
- Decision extraction
- Tradeoff extraction

AI should not be used for:

- Repository scanning
- AST parsing
- File discovery

---

## Development Workflow

Before implementation:

1. Read relevant architecture documents.
2. Create implementation plan.
3. Implement.
4. Add tests.
5. Update documentation.

---

## Current Goal

Complete ISSUE-057 (Code Intelligence & Development Experience) to deliver Software Understanding and approve RepoNerve for v1.0.0 release.

# RepoNerve Architecture Rules

## Core Philosophy

Memory First.

Relationships Second.

Context Third.

Agents Fourth.

Evidence Always.

RepoNerve is a repository knowledge system.

Every capability must be derived from repository evidence.

---

# Architectural Principles

## Single Source of Truth

Repository knowledge must be extracted once.

Subsequent systems must consume repository memory rather than re-scan repository sources.

Preferred:

Repository
↓
Ingestion
↓
Memory
↓
Consumers

Avoid:

Repository
↓
Feature

Feature
↓
Repository

---

## Layering

Dependency direction must remain:

Storage
↓
Read Stores
↓
Query Engines
↓
Context Engines
↓
MCP
↓
Agents

Upper layers must not bypass lower layers.

---

## Reuse Before Reinvention

New capabilities should reuse existing engines whenever possible.

Examples:

Ownership reuses Memory.

Context reuses Query Engine.

MCP reuses Query and Context Engines.

Graph Intelligence reuses Memory and Ownership.

---

# Determinism Requirements

All outputs must be:

* Deterministic
* Reproducible
* Testable

The same repository state must produce the same results.

Deterministic ordering is required for:

* Queries
* Context Generation
* Ownership
* Graph Traversal
* MCP Outputs

---

# Explainability Requirements

Every conclusion must be explainable.

Unsupported:

* Subjective rankings
* AI-generated ownership
* Heuristic assumptions without evidence

Supported:

* Evidence-based conclusions
* Repository-derived relationships
* Traceable recommendations

---

# Evidence Requirements

Evidence is mandatory.

Invalid:

Expertise Score
↓
No Evidence

Valid:

Expertise Score
↓
Evidence

Graph Edge
↓
Evidence

Recommendation
↓
Evidence

Rule:

Evidence-Free Conclusions Are Invalid.

---

# Ownership Intelligence Rules

Contributor identity must be deterministic.

Recommended identity:

RepositoryID + Email

Ownership recommendations are derived conclusions.

Ownership recommendations are not facts.

Ownership recommendations must expose evidence.

---

# Knowledge Graph Rules

## Graph Nodes

Graph nodes wrap existing repository entities.

Graph nodes do not duplicate repository entities.

Correct:

GraphNode
↓
EntityType
↓
EntityID

Incorrect:

Decision
↓
GraphDecision

The Memory Engine remains the source of truth.

---

## Relationship Categories

Stored Relationships

* Persisted
* Extracted
* Fact-based

Derived Relationships

* Computed
* Explainable
* Evidence-backed

Rule:

Stored Relationships are facts.

Derived Relationships are conclusions.

---

## Graph Edge Evidence

Every graph edge must contain evidence.

Graph edges without evidence are invalid.

---

# MCP Rules

MCP tools must remain thin.

Preferred:

MCP
↓
Query Engine
↓
Context Engine
↓
Graph Engine

Avoid:

MCP
↓
SQLite

MCP must not contain business logic.

---

# Storage Rules

Store interfaces are mandatory.

Consumers must not access SQLite directly.

Use:

SQLite
↓
Stores
↓
Readers
↓
Services

---

# Testing Rules

Every feature must include:

* Unit Tests
* Integration Tests

Graph and ownership features must additionally verify:

* Deterministic behavior
* Evidence preservation
* Ordering guarantees

---

# Documentation Rules

Every milestone must include:

Architecture

PRD

Tasks

Implementation

Audit

Release

in that order.

Implementation must not begin before Architecture and PRD are approved.

---

# Commit Convention

Architecture:

docs(<area>): ...

Roadmap:

docs(<area>): define roadmap

Tasks:

docs(<area>): define implementation roadmap

Implementation:

feat(<area>): ...

Audit:

docs(audit): ...

Release:

release: <version> <description>

Examples:

feat(ownership): implement expertise detection

feat(graph): implement graph traversal engine

release: v0.7.0-alpha ownership intelligence complete
