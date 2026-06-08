# RepoNerve Agent Instructions

## Purpose

RepoNerve is an open-source memory and context engine for software repositories.

Its purpose is to preserve repository knowledge and generate optimized context for humans and AI systems.

---

## Core Principles

Memory First.

Context Second.

AI Third.

---

## Current Development Phase

Phase 7 - Release Readiness

Current Release Target:

v1.0.0

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

Approve RepoNerve for v1.0.0 release.

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
