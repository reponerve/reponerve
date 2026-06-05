# RepoNerve Package Structure

Version: 1.0

Status: Draft

Authors: RepoNerve Contributors

Last Updated: 2026-06-05

---

# Purpose

This document defines the package structure, module boundaries, and dependency rules for the RepoNerve codebase.

The goal is to ensure:

* Clear ownership
* Low coupling
* High cohesion
* Testability
* Extensibility
* Long-term maintainability

This structure is designed for the MVP while allowing future expansion into MCP, APIs, and additional integrations.

---

# Architectural Philosophy

RepoNerve is not a CLI application.

RepoNerve is a platform.

The CLI is merely the first interface.

Future interfaces include:

* MCP
* HTTP APIs
* Background services
* Web UI

All interfaces must consume the same core platform.

---

# Dependency Rule

Dependencies must always flow inward.

```text
CLI
 │
 ▼
Application Layer
 │
 ▼
Domain Layer
 │
 ▼
Storage Layer
```

The Domain Layer must never depend on:

* CLI
* MCP
* APIs
* UI

---

# Repository Structure

```text
reponerve/
│
├── cmd/
│
├── internal/
│
├── pkg/
│
├── docs/
│
├── examples/
│
├── scripts/
│
├── testdata/
│
├── .github/
│
└── Makefile
```

---

# cmd/

Contains executable entrypoints.

---

## Structure

```text
cmd/
└── reponerve/
    └── main.go
```

---

## Responsibility

Bootstrap the application.

Responsibilities:

* Parse CLI arguments
* Load configuration
* Initialize dependencies
* Execute commands

Must contain minimal business logic.

---

# internal/

Contains implementation details.

Packages inside internal are not intended for external consumption.

---

## Structure

```text
internal/
├── app/
├── config/
├── storage/
├── search/
├── parser/
├── extraction/
├── ingestion/
├── context/
├── query/
├── ai/
└── mcp/
```

---

# internal/app

Application orchestration layer.

---

## Responsibilities

* Service initialization
* Dependency wiring
* Startup lifecycle

---

## Example

```go
type Application struct {
    MemoryService MemoryService
    QueryService QueryService
}
```

---

# internal/config

Configuration management.

---

## Responsibilities

* Load config files
* Environment variables
* Default values

---

## Example

```text
.reponerve/config.yaml
```

---

# internal/storage

Persistence layer.

---

## Responsibilities

* SQLite access
* Migrations
* Repository management

---

## Structure

```text
storage/
├── sqlite/
├── migrations/
└── repository/
```

---

# internal/search

Search infrastructure.

---

## Responsibilities

* FTS queries
* Ranking
* Search expansion

---

## Future

May support:

* Hybrid search
* Semantic search
* Embedding retrieval

---

# internal/parser

Artifact parsers.

---

## Responsibilities

Convert repository artifacts into structured records.

---

## Structure

```text
parser/
├── git/
├── adr/
├── markdown/
├── tree_sitter/
└── docs/
```

---

# internal/extraction

Memory extraction engine.

---

## Responsibilities

Transform parsed artifacts into memory entities.

---

## Example

```text
Commit
      ▼
Decision Memory
```

---

## Components

```text
extraction/
├── facts/
├── decisions/
├── intent/
├── ownership/
└── relationships/
```

---

# internal/ingestion

Repository ingestion pipeline.

---

## Responsibilities

* Repository scanning
* Incremental indexing
* Artifact processing
* Pipeline orchestration

---

## Flow

```text
Repository
      ▼
Discovery
      ▼
Parsing
      ▼
Extraction
      ▼
Storage
```

---

# internal/context

Context Engine.

---

## Responsibilities

Generate context packs.

---

## Example

Input:

```text
Add MFA support
```

Output:

```text
Relevant Services

Relevant ADRs

Relevant Decisions

Relevant Files
```

---

# internal/query

Query engine.

---

## Responsibilities

Answer repository questions.

---

## Example

```text
Why was Redis introduced?
```

---

## Flow

```text
Question
      ▼
Search
      ▼
Memory Retrieval
      ▼
Evidence Collection
      ▼
Response
```

---

# internal/ai

Optional AI integrations.

---

## Responsibilities

* Decision extraction
* Intent extraction
* Tradeoff extraction

---

## Design Rule

AI providers must be replaceable.

---

## Interface

```go
type Extractor interface {
    Extract(ctx context.Context, input string) (*Result, error)
}
```

---

## Supported Providers

Future:

* Ollama
* OpenAI
* Anthropic
* Local models

---

# internal/mcp

MCP server implementation.

---

## Responsibilities

Expose RepoNerve capabilities to AI systems.

---

## Future Tools

```text
get_repository_memory

get_context_pack

explain_component

find_related_decisions
```

---

# pkg/

Public reusable packages.

These packages may be imported by external projects.

---

## Structure

```text
pkg/
├── memory/
├── models/
├── types/
└── sdk/
```

---

# pkg/memory

Public memory abstractions.

---

## Example

```go
type Memory struct {
    ID string
    Type string
}
```

---

# pkg/models

Shared domain models.

---

## Examples

```go
Decision
Intent
Fact
Event
```

---

# pkg/types

Reusable public types.

---

# pkg/sdk

Future client SDK.

---

## Purpose

Allow external tools to consume RepoNerve.

---

# Domain Boundaries

---

## Memory Domain

Responsible for:

* Facts
* Events
* Decisions
* Ownership
* Intent

---

## Ingestion Domain

Responsible for:

* Discovery
* Parsing
* Processing

---

## Context Domain

Responsible for:

* Context packs
* Relevance ranking

---

## Query Domain

Responsible for:

* Retrieval
* Explanation

---

# Interface Layer

Current:

```text
CLI
```

Future:

```text
CLI

MCP

API

Web UI
```

All must consume the same services.

---

# CLI Package Structure

```text
internal/cli/
├── init/
├── scan/
├── ask/
├── explain/
└── context/
```

---

# Future HTTP API

Future package:

```text
internal/api/
```

---

## Example Endpoints

```text
GET /memory

GET /decisions

POST /context
```

---

# Future MCP Server

Future package:

```text
internal/mcp/
```

---

## Example Tools

```text
get_context_pack

get_repository_memory

explain_component
```

---

# Testing Structure

```text
tests/
├── integration/
├── e2e/
└── fixtures/
```

---

# Dependency Rules

Allowed:

```text
CLI
  ▼
Services
  ▼
Storage
```

---

Forbidden:

```text
Storage
  ▼
CLI
```

---

Forbidden:

```text
Memory
  ▼
MCP
```

---

Forbidden:

```text
Query
  ▼
CLI
```

---

# Build Targets

MVP binaries:

```text
reponerve
```

Future binaries:

```text
reponerve

reponerve-mcp

reponerve-api
```

---

# Future Evolution

The package structure must support:

* MCP integrations
* AI skills
* Multi-repository memory
* Context packs
* Cloud deployments

without requiring major refactoring.

---

# Success Criteria

The package structure succeeds when:

* Components remain loosely coupled.
* Interfaces remain replaceable.
* Business logic remains independent.
* New interfaces can be added safely.

---

# Guiding Principle

Build RepoNerve as a platform.

The CLI is the first consumer, not the product.
