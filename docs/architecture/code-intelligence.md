# Code Intelligence

Version: v1.0

Status: Draft

Issue: ISSUE-057

See also:

* `docs/architecture/code-intelligence-v1.md`
* `docs/architecture/code-storage-model.md`
* `docs/architecture/issue-057-architecture.md`

---

# Overview

Code Intelligence is one capability within RepoNerve's software understanding mission.

It is the authoritative source for **code understanding** — how software works.

It answers:

* How does it work?
* Which modules, packages, and files are involved?
* Which structs, interfaces, and type aliases are involved?
* Which functions and methods are involved?
* Which endpoints are involved?
* What is the execution flow?
* What is the call graph?
* Which symbols depend on this symbol?

Code Intelligence complements Repository Intelligence (why software exists) and serves Development Experience (how users consume RepoNerve).

See `docs/vision/vision.md` for the Software Understanding model.

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

# Entities

| EntityType | Description |
| --- | --- |
| `module` | Go module from `go.mod` or `go.work` |
| `package` | Go package boundary |
| `file` | Source file path and metadata |
| `struct` | Struct declaration |
| `interface` | Interface declaration |
| `type_alias` | Type alias declaration |
| `function` | Top-level function |
| `method` | Type-associated function |
| `endpoint` | Exposed endpoint surface |

### Endpoint Types

| EndpointType | Description |
| --- | --- |
| `http` | HTTP handler or route |
| `grpc` | gRPC service method |
| `graphql` | GraphQL resolver or operation |
| `cli` | CLI command or subcommand |

Each entity must include:

* EntityID
* EntityType
* Name
* QualifiedName
* FilePath
* PackagePath
* ModulePath
* StartLine
* EndLine
* EvidenceJSON

Endpoint entities additionally include `EndpointType`.

---

# Relationships

| Relationship | Description |
| --- | --- |
| `MODULE_CONTAINS_PACKAGE` | Module contains package |
| `BELONGS_TO_MODULE` | Package belongs to module |
| `BELONGS_TO_PACKAGE` | File or symbol belongs to package |
| `DEFINED_IN_FILE` | Symbol defined in file |
| `IMPORTS` | File imports package |
| `CALLS` | Function or method invokes another symbol |
| `IMPLEMENTS` | Struct satisfies interface |
| `DEPENDS_ON` | Package depends on package |
| `REFERENCES` | Symbol references another symbol |
| `EXPOSES_ENDPOINT` | Function or method exposes endpoint |

Every relationship must include evidence:

* Source file
* Source line
* Relationship type

---

# Repository ↔ Code Linkage

Code Intelligence does not create repository memory.

Cross-authority references are stored in `repository_code_relationships`:

| Relationship | Description |
| --- | --- |
| `DECISION_REFERENCES_CODE` | Decision cites code entity |
| `FACT_REFERENCES_CODE` | Fact references code entity |
| `EVENT_REFERENCES_CODE` | Event references code entity |
| `CONTEXT_REFERENCES_CODE` | Context artifact references code entity |

Links must be deterministic. See linking rules in `docs/architecture/issue-057-architecture.md`.

---

# Authority Rules

## Code Intelligence Authority

Code Intelligence is authoritative for:

* Code structure
* Symbol resolution
* Call graphs
* Symbol dependencies

Code Intelligence must not:

* Create repository memory entities
* Generate ownership scores
* Generate reviewer recommendations
* Generate change plans
* Replace Repository Intelligence authorities

---

## Development Experience Authority

Development Experience orchestrates Code Intelligence and Repository Intelligence.

It assembles final explain, plan, impact, and review outputs.

It must not generate Purpose or History narrative fields.

It must not bypass either authority.

---

# Determinism Rules

All Code Intelligence outputs must be deterministic.

Requirements:

* Identical repository state produces identical entities and relationships
* Parsing is deterministic — no LLM extraction
* Graph traversal ordering is stable
* List outputs sort by EntityType ASC, ModulePath ASC, FilePath ASC, StartLine ASC, QualifiedName ASC, EntityID ASC

Forbidden:

* LLM parsing or ranking
* Embedding-based symbol inference
* Heuristic symbol invention without source evidence

---

# Parsing Strategy

Initial language support:

* Go — `go/ast`, `go/parser`, `go/types`
* Module discovery — `go.mod`, `go.work`

Rules:

* Parse supported language files only
* Respect repository ignore rules
* Preserve file path and line number evidence
* Re-index identical repository state to identical output

---

# Graphs

## Module Graph

Module → package containment and module membership.

## File Graph

File imports and references.

## Package Graph

Package dependencies.

## Call Graph

Function and method invocations.

All graphs must preserve evidence on every edge.

---

# Integration with Development Experience

Code Intelligence provides **Code Context**:

* Modules, files, packages
* Structs, interfaces, type aliases
* Functions, methods, endpoints
* Call graph
* Symbol dependencies

Development Experience combines Code Context with Repository Context through Repository-Code Links.

Code Intelligence alone does not produce the final user-facing explanation.

---

# Package Structure

```text
internal/code/
    models.go
    parser.go
    indexer.go
    linker.go
    graph.go
    service.go

internal/storage/sqlite/code_entity_store.go
internal/storage/sqlite/code_relationship_store.go
internal/storage/sqlite/repository_code_relationship_store.go
internal/query/storage/code_reader.go
```

---

# Service APIs

* `IndexRepository` — index modules, packages, files, and symbols
* `ResolveFile` — resolve a file path to code context
* `ResolveSymbol` — resolve a function, method, struct, interface, or type alias
* `BuildCallGraph` — build call graph from a root symbol
* `AnalyzeSymbolDependencies` — analyze symbol dependency chain
* `ListRepositoryCodeLinks` — traverse repository-code relationships

---

# Testing Requirements

Unit tests must verify:

* Module discovery from go.mod and go.work
* Struct, interface, type_alias extraction
* Endpoint extraction
* Symbol resolution
* Graph construction
* Repository-code link extraction
* Evidence preservation
* Deterministic ordering

Integration tests must verify:

* Scan → index → link → query end-to-end
* File and symbol resolution through stored indexes

---

# Release Impact

Code Intelligence is required for v1.0.0.

Tracked in ISSUE-057 — Code Intelligence & Development Experience.

Repository Intelligence alone is not sufficient for release.
