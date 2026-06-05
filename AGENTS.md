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

Phase 0 - Foundation

Current Release Target:

v0.1.0-alpha

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

Deliver v0.1.0-alpha.