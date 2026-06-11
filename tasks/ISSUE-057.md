# ISSUE-057 — Code Intelligence & Development Experience

Status: Planned

Milestone: v1.0

Depends On:

- ISSUE-049 — Reviewer Recommendations
- ISSUE-050 — Change Planning Engine
- ISSUE-051 — Agent Context Builder
- ISSUE-052 — Workflow Intelligence
- ISSUE-053 — Repository Search
- ISSUE-054 — Agent Session Intelligence

Blocks:

- v1.0.0 Release Finalization

---

# Objective

Deliver Code Intelligence, Repository-Code Linking, Feature Understanding, and Development Experience — completing RepoNerve's path to **Software Understanding**.

RepoNerve preserves and transfers software knowledge. Repository Intelligence (complete) and the ISSUE-057 deliverables are required capabilities that serve that mission.

Repository Intelligence alone does not deliver Software Understanding.

RepoNerve v1.0 is not complete until this issue is done.

---

# Release State

| Capability | Status |
| --- | --- |
| Knowledge Preservation | Core Platform Capability |
| Repository Intelligence | ✅ Complete |
| Code Intelligence | ❌ This issue |
| Repository-Code Linking | ❌ This issue |
| Feature Understanding | ❌ This issue |
| Development Experience | ❌ This issue |
| Software Understanding | 🚀 Blocked |
| RepoNerve v1.0 | 🚀 Blocked |

---

# Background

RepoNerve successfully delivers Repository Intelligence:

* Memory, Context, Ownership, Knowledge Graph
* Discovery, Learning, Reviewer Recommendations, Change Planning
* Search, Sessions, Workflows, MCP Integration

However, Repository Intelligence alone is not the final product.

RepoNerve's mission is to help humans and AI understand software systems with minimal repository exploration and minimal token consumption.

That requires:

**Code Intelligence** — how code works

**Development Experience** — development-facing orchestration of code and repository intelligence

---

# Philosophy

Orchestrate First. Generate Second. Evidence Always.

* Code Intelligence is authoritative for code understanding
* Repository Intelligence is authoritative for repository knowledge
* Development Experience orchestrates both
* No duplicate intelligence systems

---

# Part 1 — Code Intelligence

See: docs/architecture/code-intelligence.md

## Scope

* Symbol extraction
* File graph
* Package graph
* Call graph
* Symbol dependency analysis

## Entities

* Modules
* Files
* Packages
* Structs
* Interfaces
* Type Aliases
* Functions
* Methods
* Endpoints

## Relationships

* MODULE_CONTAINS_PACKAGE
* BELONGS_TO_MODULE
* BELONGS_TO_PACKAGE
* DEFINED_IN_FILE
* CALLS
* IMPORTS
* IMPLEMENTS
* DEPENDS_ON
* REFERENCES
* EXPOSES_ENDPOINT

## Repository-Code Links

* DECISION_REFERENCES_CODE
* FACT_REFERENCES_CODE
* EVENT_REFERENCES_CODE
* CONTEXT_REFERENCES_CODE

## Package

Create:

```text
internal/code/
    models.go
    parser.go
    parser_test.go
    indexer.go
    indexer_test.go
    graph.go
    graph_test.go
    service.go
    service_test.go
```

Storage:

```text
internal/storage/sqlite/code_store.go
internal/query/storage/code_reader.go
```

## Service APIs

```go
func (s *Service) IndexRepository(ctx context.Context, repositoryID, repositoryPath string) error
func (s *Service) ResolveFile(ctx context.Context, repositoryID, filePath string) (*CodeExplanationContext, error)
func (s *Service) ResolveSymbol(ctx context.Context, repositoryID, symbol string) (*CodeExplanationContext, error)
func (s *Service) BuildCallGraph(ctx context.Context, repositoryID, entityID string) (*CallGraph, error)
func (s *Service) AnalyzeSymbolDependencies(ctx context.Context, repositoryID, entityID string) (*SymbolDependencyReport, error)
```

## Parsing

* Initial language: Go (`go/ast`, `go/parser`, `go/types`)
* Deterministic only — no LLM parsing
* Incremental re-indexing on file changes
* Integrate with scan pipeline

## Acceptance Criteria — Code Intelligence

* Code structure indexed deterministically
* File and symbol resolution works
* Call graph and dependency analysis works
* Evidence preserved on all entities and relationships
* All tests pass

---

# Part 2 — Development Experience

See: docs/architecture/development-intelligence-v1.md

## Scope

Development-facing workflows that orchestrate Code Intelligence and Repository Intelligence into **Software Understanding** and **Knowledge Transfer**.

## Feature Understanding (v1.0)

Feature-level understanding is a v1.0 requirement. Development Experience must resolve feature topics (e.g. "authentication", "metadata panel") to:

```text
Feature → Code → Ownership → Decisions → Impact
```

## CLI Commands (all required for v1.0)

```bash
reponerve ask "Who created metadata panel?"
reponerve explain "metadata panel"
reponerve explain-file "metadata-panel.tsx"
reponerve explain-function "BuildMetadataPanel"
reponerve explain-struct "MetadataPanel"
reponerve explain-interface "Searcher"
reponerve explain-type "HandlerFunc"
reponerve plan "Add OAuth login"
reponerve impact "user-service"
reponerve review "metadata panel"
```

## Engines

### 1. Natural Language Question Answering

Examples: Who created X? Why Redis? Who owns authentication?

Orchestrates: Search, Ownership, Q&A, Graph.

### 2. Repository Explanation

Examples: Explain metadata panel, Explain authentication flow.

Orchestrates: Code Intelligence, Search, Context, Guidance, Ownership.

Output combines Code Context + Repository Context.

### 3. Development Planning

Examples: Add OAuth login, Add audit logging.

Orchestrates: Search, Context, Discovery, Learning, Reviewers, Change Planning, Workflows.

Output: Impacted areas, decisions, facts, owners, reviewers, workflow, starting points.

### 4. Development Impact

Examples: What breaks if I change user-service?

Orchestrates: Search, Graph Impact, Agent Impact, Code dependencies, Ownership.

### 5. Review Preparation

Examples: Who should review this?

Orchestrates: Search, Reviewers, Ownership, Workflows.

## Package

Create:

```text
internal/agent/development/
    models.go
    router.go
    router_test.go
    service.go
    service_test.go
```

CLI (thin wrappers only):

```text
internal/cli/ask/
internal/cli/explain/
internal/cli/explainfile/
internal/cli/explainfunction/
internal/cli/plan/
internal/cli/impact/
internal/cli/review/
```

## Explain Output Contract

**Code Context** (from Code Intelligence)

* Modules, files, packages, structs, interfaces, type aliases, functions, methods, endpoints
* Call graph, symbol dependencies

**Repository Context** (from Repository Intelligence)

* Decisions, facts, events
* Ownership, expertise, reviewers
* Impact, change plans

**Repository-Code Links**

* Deterministic connections between repository entities and code entities

No Purpose or History fields. Rationale appears as structured Decisions, Facts, and Events.

## Reuse Requirements

Must reuse without duplication:

* Repository Search
* Agent Context Builder
* Workflow Intelligence
* Agent Session Intelligence
* Repository Q&A, Guidance, Impact, Onboarding
* Discovery, Learning, Reviewers, Change Planning
* Knowledge Graph Impact, Context Engine, Ownership Intelligence
* Code Intelligence

Must NOT:

* Access SQLite directly from CLI
* Introduce new intelligence scoring systems
* Duplicate code parsing or repository intelligence

## Acceptance Criteria — Development Experience

* All CLI commands work end-to-end
* Explain output combines code and repository context
* Natural language topics resolve through Search and Code Intelligence
* Evidence and provenance preserved
* Deterministic ordering
* All tests pass

---

# Combined Acceptance Criteria

RepoNerve v1.0 is complete when all v1.0 scope items in `docs/vision/vision.md` are delivered.

```bash
reponerve ask "Who created metadata panel?"
reponerve explain "metadata panel"
reponerve explain-file "metadata-panel.tsx"
reponerve explain-function "BuildMetadataPanel"
reponerve explain-struct "MetadataPanel"
reponerve explain-interface "Searcher"
reponerve explain-type "HandlerFunc"
reponerve plan "Add OAuth login"
reponerve impact "user-service"
reponerve review "metadata panel"
```

All execute successfully with evidence-backed, deterministic output.

| v1.0 Capability | Required |
| --- | --- |
| Knowledge Preservation | ✅ Core platform operational |
| Repository Intelligence | ✅ |
| Code Intelligence | ✅ |
| Repository-Code Linking | ✅ |
| Feature Understanding | ✅ |
| Development Experience | ✅ |
| Software Understanding | ✅ |

There is no v1.x product release. v1.0.0 is the only release, delivered via v0.10–v0.15 alpha iterations.

---

# Documentation

* docs/architecture/issue-057-architecture.md  ← **Architecture (approve before implementation)**
* docs/architecture/code-storage-model.md
* docs/architecture/development-experience-contracts.md
* docs/examples/development-experience.md      ← **Acceptance criteria**

Update on completion:

* docs/architecture/cli-reference-v1.md
* README.md

---

# Related v1.0 Issues

Foundation fixes moved to **ISSUE-059** (`v0.10.0-alpha`).

Token Intelligence, Evidence Graph, and Multi-language scope: **ISSUE-060**, **ISSUE-061**, **ISSUE-062**.

Full iteration map: `docs/roadmap/v1.0-iteration-plan.md`.

---

# Constraints

Implementation must not begin until:

* `docs/architecture/issue-057-architecture.md` is approved
* `docs/architecture/architecture-overview.md` v1.1 realignment (ARCH-001) is approved

Acceptance criteria: `docs/examples/development-experience.md`

Do NOT:

* Add LLM-required parsing or routing
* Add embeddings or vector search
* Duplicate Repository Intelligence or Code Intelligence authorities
* Finalize v1.0.0 before this issue is complete

Initial Code Intelligence language support: Go.

Additional language adapters may follow without changing authority boundaries.
