# Code Storage Model

Status: Draft

Related Milestone:

* ISSUE-057 — Code Intelligence & Development Experience

Related:

* `docs/architecture/issue-057-architecture.md`

---

# Objective

Define how Code Intelligence is represented, stored, queried, and consumed inside RepoNerve.

Code Intelligence is the authoritative source for code understanding.

Repository Intelligence remains the authoritative source for repository understanding.

Cross-authority references are stored in `repository_code_relationships`.

---

# Entity Hierarchy

```text
Module
    ↓
Package
    ↓
File
    ↓
Symbols
    ├── struct
    ├── interface
    ├── type_alias
    ├── function
    ├── method
    └── endpoint
```

---

# Design Principles

Code Intelligence must be:

* Deterministic
* Explainable
* Reproducible
* Language-aware
* Independent from LLMs

Code Intelligence must not:

* Use embeddings
* Use semantic search
* Infer ownership
* Infer reviewer recommendations
* Infer repository decisions

These responsibilities belong to Repository Intelligence.

---

# Authority Boundaries

## Code Intelligence Owns

Modules

Packages

Files

Structs

Interfaces

Type Aliases

Functions

Methods

Endpoints

Imports

Call Graphs

Dependency Graphs

## Repository Intelligence Owns

Decisions

Facts

Events

Ownership

Expertise

Reviewers

Change Plans

Learning Paths

Repository Context

## Repository-Code Links Own

Deterministic references between repository entities and code entities.

Stored separately. Neither authority duplicates the other.

---

# Core Entities

## Module

Represents a Go module boundary from `go.mod` or `go.work`.

```go
type Module struct {
    ID           string
    RepositoryID string
    Name         string
    ModulePath   string
    Language     string
    EvidenceJSON string
    IndexedAt    time.Time
}
```

## File

Represents a physical source file.

Fields:

* ID, RepositoryID, Path, PackagePath, ModulePath, Language, EvidenceJSON

## Package

Represents a logical Go package.

Fields:

* ID, RepositoryID, Name, PackagePath, ModulePath

## Struct

Represents a struct declaration.

Fields:

* ID, RepositoryID, Name, QualifiedName, PackagePath, ModulePath, FilePath, StartLine, EndLine

## Interface

Represents an interface declaration.

Fields:

* ID, RepositoryID, Name, QualifiedName, PackagePath, ModulePath, FilePath, StartLine, EndLine

## TypeAlias

Represents a type alias declaration.

Fields:

* ID, RepositoryID, Name, QualifiedName, PackagePath, ModulePath, FilePath, StartLine, EndLine

## Function

Represents a standalone function.

Fields:

* ID, RepositoryID, Name, QualifiedName, PackagePath, ModulePath, FilePath, StartLine, EndLine, Signature

## Method

Represents a receiver-based method.

Fields:

* ID, RepositoryID, Name, QualifiedName, Receiver, PackagePath, ModulePath, FilePath, StartLine, EndLine, Signature

## Endpoint

Represents an exposed endpoint surface.

Fields:

* ID, RepositoryID, Name, QualifiedName, EndpointType, PackagePath, ModulePath, FilePath, StartLine, EndLine

EndpointType values: `http`, `grpc`, `graphql`, `cli`

---

# Code Relationships

| RelationshipType | From → To |
| --- | --- |
| `MODULE_CONTAINS_PACKAGE` | module → package |
| `BELONGS_TO_MODULE` | package → module |
| `BELONGS_TO_PACKAGE` | file, symbol → package |
| `DEFINED_IN_FILE` | symbol → file |
| `IMPORTS` | file → package |
| `CALLS` | function/method → function/method |
| `IMPLEMENTS` | struct → interface |
| `DEPENDS_ON` | package → package |
| `REFERENCES` | symbol → symbol |
| `EXPOSES_ENDPOINT` | function/method → endpoint |

Relationships must be deterministic and originate from parser output.

`DEFINED_IN` is not used.

---

# Repository-Code Relationships

Stored in `repository_code_relationships`.

| RelationshipType | From → To |
| --- | --- |
| `DECISION_REFERENCES_CODE` | decision → code entity |
| `FACT_REFERENCES_CODE` | fact → code entity |
| `EVENT_REFERENCES_CODE` | event → code entity |
| `CONTEXT_REFERENCES_CODE` | context artifact → code entity |

```go
type RepositoryCodeRelationship struct {
    ID                   string
    RepositoryID         string
    RepositoryEntityID   string
    RepositoryEntityType string
    CodeEntityID         string
    CodeEntityType       string
    RelationshipType     string
    EvidenceJSON         string
    IndexedAt            time.Time
}
```

Linking rules: see `docs/architecture/issue-057-architecture.md`.

---

# Storage Tables

* `code_entities`
* `code_relationships`
* `repository_code_relationships`
* `code_index_state`

Full DDL: `docs/architecture/issue-057-architecture.md`

---

# Storage Rules

Code Intelligence must be queryable independently from Repository Intelligence.

Repository-code links must be queryable independently from both.

Recommended package:

```text
internal/code/
    models.go
    parser.go
    indexer.go
    linker.go
    graph.go
    service.go
```

---

# Ingestion Pipeline

```text
Repository
    ↓
Module Discovery (go.mod, go.work)
    ↓
Parser
    ↓
Entity Extraction
    ↓
Relationship Extraction
    ↓
Code Store
    ↓
Repository-Code Linking
    ↓
Repository-Code Link Store
    ↓
Query Services
```

---

# Parser Requirements

Go support is required for v1.

Implementation:

* `go/parser`
* `go/ast`
* `go/types`

Module discovery:

* `go.mod`
* `go.work` (monorepo support)

No LLM participation.

No semantic inference.

---

# Query Requirements

Code Intelligence must support:

* Find Module
* Find Package
* Find File
* Find Struct
* Find Interface
* Find Type Alias
* Find Function
* Find Method
* Find Endpoint
* Find Callers
* Find Callees
* Find Imports
* Find Dependencies

Repository-Code link queries must support:

* List links by repository entity
* List links by code entity
* Traverse repository entity → code entities
* Traverse code entity → repository entities

---

# Determinism Requirements

The same repository state must produce:

* The same entities
* The same relationships
* The same repository-code links
* The same query results

No runtime randomness is permitted.

---

# Future Language Support

Not part of v1.

Potential future support:

* TypeScript
* Java
* Python
* Rust

The storage model must remain language-neutral where possible.

Module and endpoint abstractions support multi-language extension.
