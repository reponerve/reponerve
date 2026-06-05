# RepoNerve Architecture Overview

Version: 1.0

Status: Draft

Authors: RepoNerve Contributors

Last Updated: 2026-06-05

---

# Purpose

This document defines the high-level architecture of RepoNerve.

It describes:

* System responsibilities
* Major components
* Data flow
* Technology decisions
* Architectural principles

This document serves as the foundation for all future engineering work.

---

# Architectural Vision

RepoNerve is an open-source memory and context engine for software repositories.

Its primary responsibility is to:

1. Discover repository knowledge
2. Extract repository memory
3. Store repository memory
4. Retrieve repository memory
5. Generate repository context

RepoNerve is not responsible for:

* Code generation
* Repository hosting
* Workflow automation
* Project management

---

# Architectural Principles

## Memory First

Repository memory is the primary system asset.

All other capabilities are built on top of memory.

---

## Local First

RepoNerve must work without:

* Cloud services
* SaaS dependencies
* Hosted infrastructure

The local machine is the default deployment target.

---

## Offline Capable

Core functionality should work without internet access after repository data has been acquired.

---

## Evidence Driven

All generated memory must be traceable to repository artifacts.

---

## AI Optional

AI enhances memory extraction.

AI is not required for the majority of system functionality.

---

## Interface Agnostic

The core platform must be independent of:

* CLI
* APIs
* MCP
* Future UIs

Interfaces consume the platform.

They do not define it.

---

# System Overview

```text
┌─────────────────────┐
│     Repository      │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Repository Scanner  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Ingestion Pipeline  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Memory Extraction   │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   Memory Store      │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   Query Engine      │
└──────────┬──────────┘
           │
   ┌───────┼────────┐
   ▼       ▼        ▼
 CLI      MCP      API
```

---

# Core Components

RepoNerve consists of six major subsystems.

---

# Repository Scanner

Responsibility:

Discover repository artifacts.

---

## Inputs

* Source code
* Git history
* Pull requests
* ADRs
* Documentation

---

## Outputs

Normalized repository artifacts.

---

## Responsibilities

* Repository discovery
* File traversal
* Metadata extraction
* Change detection

---

# Ingestion Pipeline

Responsibility:

Transform repository artifacts into processable records.

---

## Inputs

Raw repository artifacts.

---

## Outputs

Structured ingestion records.

---

## Responsibilities

* Parsing
* Normalization
* Classification
* Validation

---

# Memory Extraction Engine

Responsibility:

Generate repository memory.

---

## Inputs

Structured repository artifacts.

---

## Outputs

Memory entities.

---

## Memory Types

* Facts
* Events
* Decisions
* Ownership
* Relationships
* Intent

---

## Responsibilities

* Knowledge extraction
* Relationship creation
* Evidence linking
* Context generation support

---

# Memory Store

Responsibility:

Persist repository memory.

---

## Requirements

* Local-first
* Queryable
* Searchable
* Portable

---

## Stored Data

* Memory entities
* Sources
* Metadata
* Search indexes

---

# Query Engine

Responsibility:

Retrieve repository memory.

---

## Supported Queries

Examples:

```text
Why was Redis introduced?

Who owns billing?

Why was Kafka selected?

Explain authentication service.
```

---

## Responsibilities

* Retrieval
* Ranking
* Evidence collection
* Response generation

---

# Interface Layer

Responsibility:

Expose RepoNerve functionality.

---

## Supported Interfaces

### CLI

Primary MVP interface.

---

### MCP

Future interface.

---

### API

Future interface.

---

### UI

Future interface.

---

# Deployment Model

MVP deployment model:

```text
Developer Machine
       │
       ▼
RepoNerve CLI
       │
       ▼
Local Memory Store
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
Ingestion
    │
    ▼
Memory Extraction
    │
    ▼
Memory Store
```

---

## Query Flow

```text
Question
    │
    ▼
Query Engine
    │
    ▼
Memory Retrieval
    │
    ▼
Evidence Collection
    │
    ▼
Response
```

---

# Technology Decisions

## Programming Language

Golang

Reasons:

* Performance
* Portability
* Distribution simplicity
* Strong CLI ecosystem

---

## Storage

SQLite

Reasons:

* Embedded
* Reliable
* Local-first
* No server required

---

## Search

SQLite FTS5

Reasons:

* Integrated
* Lightweight
* Fast

---

## Parsing

Tree-sitter

Reasons:

* Multi-language support
* Structural code analysis

---

# AI Integration Strategy

RepoNerve should not depend on a specific model provider.

Supported approaches:

* Local models
* Open models
* Commercial models

Examples:

* Ollama
* Local LLMs
* Remote APIs

---

# AI Usage Policy

AI should be used only when deterministic extraction is insufficient.

Good use cases:

* Decision extraction
* Intent extraction
* Tradeoff extraction

Bad use cases:

* Basic code parsing
* Dependency discovery
* Repository traversal

---

# Future Architecture Evolution

## Phase 1

CLI

Memory Engine

Local Storage

---

## Phase 2

Context Engine

Pattern Memory

---

## Phase 3

MCP Integration

Agent Skills

---

## Phase 4

API Layer

Remote Consumption

---

## Phase 5

Optional UI

Visualization

---

# Architectural Non-Goals

RepoNerve is not intended to become:

* A code editor
* A repository hosting platform
* An AI coding agent
* A workflow engine
* A project management tool

These concerns remain outside the architecture.

---

# Success Criteria

The architecture succeeds when:

* Repository memory is durable.
* Memory retrieval is reliable.
* Evidence remains traceable.
* Context generation becomes efficient.
* Multiple interfaces can consume the same memory layer.

---

# Guiding Principle

Memory is the platform.

Everything else is an interface.
