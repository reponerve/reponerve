# RepoNerve Architecture Overview

Version: 1.1

Status: Approved

Authors: RepoNerve Contributors

Last Updated: 2026-06-09

Related Task: `tasks/ARCH-001.md`

---

# Purpose

This document defines the high-level architecture of RepoNerve.

It describes:

* System responsibilities
* Architectural pillars and subsystems
* Data flow
* Technology decisions
* Architectural principles

This document serves as the foundation for all future engineering work.

---

# Mission

RepoNerve is a software understanding platform built around knowledge preservation. Its purpose is to ensure that software understanding survives beyond individual contributors and remains accessible to both humans and AI systems.

RepoNerve serves as a **software memory system** — preserving understanding that would otherwise be lost through contributor turnover, architectural evolution, and documentation drift.

```text
Knowledge Preservation
        ↓
Software Memory
        ↓
Software Understanding
```

---

# Architectural Vision

RepoNerve is the intelligence layer for software understanding.

Its primary responsibility is to preserve, organize, and transfer software knowledge so that **Software Understanding** remains accessible to humans and AI systems.

System responsibilities:

1. Preserve software knowledge (Knowledge Preservation)
2. Extract and serve repository knowledge (Repository Intelligence)
3. Index and serve code structure (Code Intelligence)
4. Link repository knowledge to code entities (Repository-Code Linking)
5. Resolve feature-level understanding (Feature Understanding)
6. Orchestrate intelligence into development workflows (Development Experience)
7. Retrieve, traverse, and assemble understanding (Understanding Engine)
8. Deliver Software Understanding through interfaces (CLI, MCP)

RepoNerve is not responsible for:

* Code generation
* Repository hosting
* Workflow automation
* Project management

---

# Product Mission

RepoNerve delivers **Software Understanding** — not merely repository or code intelligence.

```text
Code Understanding
    +
Repository Understanding
    +
Ownership Understanding
    +
Architectural Understanding
    +
Change Understanding
    +
Historical Understanding
    ═══════════════════════
    Software Understanding
```

---

# Architectural Pillars

RepoNerve v1.0 is organized around seven architectural pillars plus one outcome.

```text
Knowledge Preservation          (foundation subsystem)
    ↓
Repository Intelligence         (capability — why)
    +
Code Intelligence               (capability — how)
    ↓
Repository-Code Linking           (cross-authority subsystem)
    ↓
Feature Understanding            (emerging capability — what)
    ↓
Development Experience          (product surface)
    ↓
Software Understanding          (outcome)
```

The **Understanding Engine** spans retrieval and context assembly across Repository Intelligence, Code Intelligence, and Repository-Code links. It is not a separate product pillar — it is the retrieval layer that connects preservation to experience.

Authority boundaries between intelligence layers are unchanged. See layer-specific architecture documents.

---

# System Overview

```text
Repository
        │
        ▼
┌─────────────────────┐
│ Repository Scanner  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Ingestion Pipeline  │
└───────┬─────┬───────┘
        │     │
        ▼     ▼
┌─────────────────┐   ┌─────────────────┐
│ Repository Int. │   │ Code Int.       │
└────────┬────────┘   └────────┬────────┘
         │                     │
         └─────────┬───────────┘
                   ▼
        ┌───────────────────┐
        │ Repository-Code   │
        │ Linking           │
        └─────────┬─────────┘
                  ▼
        ┌───────────────────┐
        │ Feature           │
        │ Intelligence      │
        └─────────┬─────────┘
                  ▼
        ┌───────────────────┐
        │ Development Exp.  │
        └─────────┬─────────┘
                  ▼
        ┌───────────────────┐
        │ Software          │
        │ Understanding     │
        └─────────┬─────────┘
                  │
          ┌───────┼───────┐
          ▼       ▼       ▼
         CLI     MCP    API
```

All pillars rest on the **Knowledge Preservation Layer** (memory store, code store, link store, indexes). The diagram above shows the product flow; preservation is the substrate beneath every box.

---

# Core Subsystems

RepoNerve consists of nine major subsystems organized by architectural pillar.

---

# Knowledge Preservation Layer

Responsibility:

Durable storage and indexing of all software knowledge.

This is the heart of RepoNerve — not merely a philosophy, but a foundational subsystem that every other pillar depends on.

---

## Preserved Artifacts

* **Memory** — Decisions, Facts, Events, Intent, Relationships
* **Ownership** — Contributor expertise, domain ownership, reviewer signals
* **Context** — Generated repository context packages
* **Code entities** — Modules, packages, files, symbols, endpoints
* **Code relationships** — Calls, imports, implements, depends-on
* **Repository-Code links** — Cross-authority references between memory and code

---

## Storage

* SQLite (local-first, embedded)
* FTS5 search indexes
* Code entity and relationship tables
* `repository_code_relationships` for cross-authority links

---

## Requirements

* Local-first
* Queryable
* Searchable
* Portable
* Evidence-linked

---

# Repository Scanner

Responsibility:

Discover repository artifacts for ingestion.

---

## Inputs

* Source code
* Git history
* Pull requests
* ADRs
* Documentation

---

## Outputs

Normalized repository artifacts routed to the Ingestion Pipeline.

---

## Responsibilities

* Repository discovery
* File traversal
* Metadata extraction
* Change detection

---

# Ingestion Pipeline

Responsibility:

Transform repository artifacts into processable records and route them to the correct intelligence path.

---

## Inputs

Raw repository artifacts from the Scanner.

---

## Outputs

Structured ingestion records for Repository Intelligence and Code Intelligence paths.

---

## Responsibilities

* Parsing
* Normalization
* Classification
* Validation
* Routing to memory extraction or code indexing

---

# Repository Intelligence

Responsibility:

Authoritative source for repository knowledge — why software exists, who owns it, how it evolved.

Status: **Complete**

---

## Answers

* Why was this introduced?
* Who owns this domain?
* What architectural decisions led to this?
* Who should review changes?
* What areas are impacted?

---

## Components

* Memory Extraction Engine
* Memory Store (within Knowledge Preservation)
* Ownership Intelligence
* Knowledge Graph Intelligence
* Context Engine
* Repository Search
* Agent Intelligence (context packaging)

---

## Memory Types

* Facts
* Events
* Decisions
* Ownership
* Relationships
* Intent

---

## Does Not

* Parse or index code structure (Code Intelligence authority)
* Orchestrate development workflows (Development Experience authority)

---

# Code Intelligence

Responsibility:

Authoritative source for code understanding — how software works structurally.

Status: **ISSUE-057 — In Progress**

---

## Answers

* How does this code work?
* Which files, packages, and symbols are involved?
* What is the call graph?
* Which symbols depend on this symbol?
* What endpoints does this expose?

---

## Entity Hierarchy

```text
Module → Package → File → Symbol
```

Symbol types: struct, interface, type_alias, function, method, endpoint.

---

## Components

* Code Parser (Tree-sitter, Go initial)
* Code Indexer
* Code Graph (file, package, call graphs)
* Code Intelligence Service

See: `docs/architecture/code-intelligence.md`, `docs/architecture/code-storage-model.md`

---

## Does Not

* Create repository memory entities (Repository Intelligence authority)
* Link decisions to code without Repository-Code Linking subsystem
* Orchestrate development workflows (Development Experience authority)

---

# Repository-Code Linking

Responsibility:

Deterministic cross-authority references between repository entities and code entities.

Status: **ISSUE-057 — In Progress**

This subsystem is critical. Without it, Development Experience cannot combine repository context with code context.

---

## Purpose

Connect why (repository knowledge) to how (code structure).

Example:

```text
Decision: ADR-004 OAuth
    ↓ links to
oauth.go
AuthService (struct)
LoginHandler (function)
```

---

## Link Types

* `DECISION_REFERENCES_CODE`
* `FACT_REFERENCES_CODE`
* `EVENT_REFERENCES_CODE`
* Additional types as defined in `docs/architecture/issue-057-architecture.md`

---

## Components

* Repository-Code Linker (extraction during ingestion and indexing)
* `repository_code_relationships` storage
* Link traversal in Understanding Engine and Development Experience

---

## Requirements

* Deterministic link creation
* Evidence on every link
* Bidirectional traversal (repository entity → code entities, code entity → repository entities)

---

# Feature Understanding

Responsibility:

Resolve feature-level understanding across code, ownership, decisions, and impact.

Status: **ISSUE-057 — v1.0 (orchestrated via Development Experience; not a separate authority)**

Feature Understanding is not a separate authority. It is orchestrated by Development Experience using Repository Intelligence, Code Intelligence, Repository-Code links, and Repository Search.

---

## Model

```text
Feature
    ↓
Code
    ↓
Ownership
    ↓
Decisions
    ↓
Impact
```

Examples: Authentication, Billing, Metadata Management, Notifications, Search.

---

## v1.0 Delivery

Feature topics resolve through natural language input (e.g. `reponerve explain "authentication"`) via:

1. Repository Search — topic → repository entities
2. Repository-Code links — repository entities → code entities
3. Code Intelligence — code structure and call graph
4. Repository Intelligence — ownership, decisions, impact

Full feature entity modeling may evolve post-v1.0. v1.0 delivers feature-level understanding through orchestration.

---

# Development Experience

Responsibility:

Orchestrate Code Intelligence, Repository Intelligence, and Repository-Code links into development-facing workflows.

Status: **ISSUE-057 — In Progress**

Development Experience is the primary user-facing layer — how humans and AI consume RepoNerve.

---

## Workflows

```bash
reponerve ask "Who created metadata panel?"
reponerve explain "metadata panel"
reponerve explain-file "metadata-panel.go"
reponerve explain-function "BuildMetadataPanel"
reponerve explain-struct "MetadataPanel"
reponerve explain-interface "Searcher"
reponerve explain-type "HandlerFunc"
reponerve plan "Add OAuth login"
reponerve impact "user-service"
reponerve review "metadata panel"
```

---

## Rules

* Orchestration only — no duplicate intelligence authorities
* Evidence-backed output — every section traceable
* Deterministic ordering
* No free-form Purpose or History narrative fields

See: `docs/architecture/development-experience-contracts.md`

---

# Understanding Engine

Responsibility:

Retrieve, traverse, and assemble understanding across all intelligence sources.

The Query Engine concept evolved as RepoNerve grew beyond repository memory. Understanding Engine is the conceptual retrieval layer; the existing Query Engine remains its repository-memory retrieval implementation.

---

## Responsibilities

* Repository Intelligence retrieval
* Code Intelligence retrieval
* Repository-Code link traversal
* Development context assembly
* Evidence collection
* Deterministic result ordering

---

## Supported Queries

Examples:

```text
Why was Redis introduced?
Who owns billing?
How does authentication work?
What depends on user-service?
Explain the metadata panel feature.
```

---

## Implementation Mapping

| Understanding Engine Responsibility | Current Implementation |
| --- | --- |
| Repository memory retrieval | Query Engine |
| Code entity resolution | Code Intelligence Service (ISSUE-057) |
| Cross-authority traversal | Repository-Code Link Store (ISSUE-057) |
| Context assembly | Context Engine + Development Experience (ISSUE-057) |

---

# Interface Layer

Responsibility:

Expose Software Understanding to consumers.

---

## Supported Interfaces

### CLI

Primary interface. Development Experience commands plus repository management.

### MCP

Agent-native interface. Exposes intelligence to AI coding agents.

### API

Future interface.

### UI

Future interface.

---

## Rule

Interfaces consume the platform. They do not define it.

The core platform is independent of CLI, MCP, API, and UI.

---

# Deployment Model

```text
Developer Machine
       │
       ▼
RepoNerve CLI
       │
       ▼
Knowledge Preservation Layer
  (memory.db + code indexes + link store)
```

---

# Repository Workspace

RepoNerve creates:

```text
.reponerve/
```

Workspace structure:

```text
.reponerve/
│
├── config.yaml
├── memory.db
├── cache/
├── indexes/
├── snapshots/
└── logs/
```

---

# Data Flow

## Initial Scan

```text
Repository
    │
    ▼
Scanner
    │
    ▼
Ingestion Pipeline
    │
    ├──────────────────────┐
    ▼                      ▼
Memory Extraction    Code Indexing
    │                      │
    ▼                      ▼
Repository           Code
Intelligence         Intelligence
    │                      │
    └──────────┬───────────┘
               ▼
    Repository-Code Linking
               │
               ▼
    Knowledge Preservation Layer
```

---

## Understanding Flow

```text
Question / Topic
    │
    ▼
Development Experience
    │
    ▼
Understanding Engine
    │
    ├────────────────┬─────────────────┐
    ▼                ▼                 ▼
Repository      Code            Repository-Code
Intelligence    Intelligence    Link Traversal
    │                │                 │
    └────────────────┴─────────────────┘
                     │
                     ▼
           Evidence Collection
                     │
                     ▼
         Software Understanding
           (CLI / MCP response)
```

---

# Architectural Principles

## Memory First (Technical)

Repository memory is the primary ingestion asset for knowledge preservation.

All intelligence capabilities are built on preserved evidence.

Product mission: Software Understanding. Technical foundation: memory-first ingestion.

---

## Understanding First (Product)

Software Understanding is the outcome. Development Experience is the surface. Intelligence layers are capabilities.

---

## Local First

RepoNerve must work without cloud services, SaaS dependencies, or hosted infrastructure.

---

## Offline Capable

Core functionality works without internet access after repository data has been acquired.

---

## Evidence Driven

All conclusions must be traceable to repository or code artifacts.

Evidence-free conclusions are invalid.

---

## AI Optional

AI enhances memory extraction. AI is not required for the majority of system functionality.

---

## Interface Agnostic

The core platform is independent of CLI, MCP, API, and future UIs.

---

# Technology Decisions

## Programming Language

Golang — performance, portability, distribution simplicity, strong CLI ecosystem.

## Storage

SQLite — embedded, reliable, local-first, no server required.

## Search

SQLite FTS5 — integrated, lightweight, fast.

## Parsing

Tree-sitter — multi-language support, structural code analysis.

---

# AI Integration Strategy

RepoNerve should not depend on a specific model provider.

Supported approaches: local models, open models, commercial models.

---

# AI Usage Policy

AI should be used only when deterministic extraction is insufficient.

Good use cases: decision extraction, intent extraction, tradeoff extraction.

Bad use cases: basic code parsing, dependency discovery, repository traversal.

---

# Architectural Non-Goals

RepoNerve is not intended to become:

* A code editor
* A repository hosting platform
* An AI coding agent
* A workflow engine
* A project management tool

---

# Success Criteria

The architecture succeeds when:

* Developers can understand unfamiliar repositories without repeated exploration.
* Knowledge survives contributor turnover and team changes.
* AI systems require less repository exploration to begin implementation work.
* Repository and code context remain connected through deterministic links.
* Development guidance is evidence-backed and explainable.
* Repository memory is durable and retrieval is reliable.
* Evidence remains traceable across all intelligence layers.
* Multiple interfaces (CLI, MCP) consume the same understanding layer.

---

# Implementation State

| Pillar | Status |
| --- | --- |
| Knowledge Preservation | Core Platform Capability |
| Repository Intelligence | Complete |
| Code Intelligence | ISSUE-057 |
| Repository-Code Linking | ISSUE-057 |
| Feature Understanding | ISSUE-057 |
| Development Experience | ISSUE-057 |
| Understanding Engine | Partial (Query Engine complete; code + link traversal pending) |
| Software Understanding | Blocked until ISSUE-057 complete |

RepoNerve v1.0 is not released until all pillars required for Software Understanding are complete.

See: `tasks/ARCH-001.md`, `tasks/ISSUE-057.md`, `docs/architecture/issue-057-architecture.md`

---

# Guiding Principle

Understanding first. Evidence second. AI third.

Knowledge preservation is the foundation. Software Understanding is the outcome.
